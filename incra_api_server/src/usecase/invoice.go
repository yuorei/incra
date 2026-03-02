package usecase

import (
	"fmt"
	"time"

	"github.com/yuorei/incra_api_server/src/domain"
	"github.com/yuorei/incra_api_server/src/domain/repository"
	"github.com/yuorei/incra_api_server/src/infrastructure"
)

type InvoiceUseCase interface {
	GetInvoice(invoiceId string) (domain.Invoice, error)
	ListInvoices(issuerSlackUserId string, status string, limit int, lastKey string) ([]domain.Invoice, string, error)
	CreateInvoice(invoice domain.Invoice) (domain.Invoice, error)
	UpdateInvoice(invoice domain.Invoice) (domain.Invoice, error)
	TransitionStatus(invoiceId string, status domain.InvoiceStatus, changedBy string, changedByUserId string) (domain.Invoice, error)
	DeleteInvoice(invoiceId string) error
}

type invoiceUseCase struct {
	invoiceRepository repository.InvoiceRepository
	clientRepository  repository.ClientRepository
}

func NewInvoiceUseCase() InvoiceUseCase {
	return &invoiceUseCase{
		invoiceRepository: infrastructure.NewInvoiceRepository(),
		clientRepository:  infrastructure.NewClientRepository(),
	}
}

func (u *invoiceUseCase) GetInvoice(invoiceId string) (domain.Invoice, error) {
	return u.invoiceRepository.GetInvoice(invoiceId)
}

func (u *invoiceUseCase) ListInvoices(issuerSlackUserId string, status string, limit int, lastKey string) ([]domain.Invoice, string, error) {
	return u.invoiceRepository.ListInvoices(issuerSlackUserId, status, limit, lastKey)
}

func (u *invoiceUseCase) CreateInvoice(invoice domain.Invoice) (domain.Invoice, error) {
	year := time.Now().Year()
	invoiceId, err := u.invoiceRepository.NextInvoiceNumber(year)
	if err != nil {
		return domain.Invoice{}, err
	}
	invoice.InvoiceId = invoiceId
	invoice.Status = domain.InvoiceStatusDraft
	total := 0
	for _, item := range invoice.Items {
		total += item.Amount
	}
	invoice.TotalAmount = total
	now := time.Now().UTC().Format(time.RFC3339)
	invoice.CreatedAt = now
	invoice.UpdatedAt = now
	return u.invoiceRepository.CreateInvoice(invoice)
}

func (u *invoiceUseCase) UpdateInvoice(invoice domain.Invoice) (domain.Invoice, error) {
	existing, err := u.invoiceRepository.GetInvoice(invoice.InvoiceId)
	if err != nil {
		return domain.Invoice{}, err
	}
	if existing.Status != domain.InvoiceStatusDraft {
		return domain.Invoice{}, fmt.Errorf("only draft invoices can be updated")
	}
	total := 0
	for _, item := range invoice.Items {
		total += item.Amount
	}
	invoice.TotalAmount = total
	invoice.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	invoice.CreatedAt = existing.CreatedAt
	invoice.IssuerSlackUserId = existing.IssuerSlackUserId
	invoice.IssuerSlackRealName = existing.IssuerSlackRealName
	return u.invoiceRepository.UpdateInvoice(invoice)
}

func (u *invoiceUseCase) TransitionStatus(invoiceId string, status domain.InvoiceStatus, changedBy string, changedByUserId string) (domain.Invoice, error) {
	existing, err := u.invoiceRepository.GetInvoice(invoiceId)
	if err != nil {
		return domain.Invoice{}, err
	}
	if err := validateStatusTransition(existing.Status, status); err != nil {
		return domain.Invoice{}, err
	}
	if err := validatePermission(existing, status, changedByUserId); err != nil {
		return domain.Invoice{}, err
	}
	updated, err := u.invoiceRepository.UpdateInvoiceStatus(invoiceId, status, changedBy)
	if err != nil {
		return domain.Invoice{}, err
	}

	// Handle side effects based on transition
	switch status {
	case domain.InvoiceStatusSent:
		if existing.Status == domain.InvoiceStatusDraft {
			// Fresh send: PDF generation + notification with pay button
			var billingClientSlackUserId string
			if updated.BillingClientId != "" {
				client, err := u.clientRepository.GetClient(updated.BillingClientId)
				if err != nil {
					fmt.Printf("warning: failed to get client for notification: %v\n", err)
				} else {
					billingClientSlackUserId = client.SlackUserId
				}
			}
			if err := infrastructure.SendPDFGenerateMessage(updated, billingClientSlackUserId); err != nil {
				fmt.Printf("warning: failed to send PDF generate message: %v\n", err)
			}
			if billingClientSlackUserId != "" {
				if err := infrastructure.SendInvoiceNotificationWithPayButton(billingClientSlackUserId, updated); err != nil {
					fmt.Printf("warning: failed to send invoice notification DM: %v\n", err)
				}
			}
		}
	case domain.InvoiceStatusPaid:
		// Notify issuer with confirm/reject buttons
		if updated.IssuerSlackUserId != "" {
			if err := infrastructure.SendPaymentConfirmationRequestDM(updated.IssuerSlackUserId, updated); err != nil {
				fmt.Printf("warning: failed to send payment confirmation request DM: %v\n", err)
			}
		}
	case domain.InvoiceStatusConfirmed:
		// Notify client that payment is confirmed
		if updated.BillingClientId != "" {
			client, err := u.clientRepository.GetClient(updated.BillingClientId)
			if err != nil {
				fmt.Printf("warning: failed to get client for notification: %v\n", err)
			} else if client.SlackUserId != "" {
				if err := infrastructure.SendPaymentConfirmedDM(client.SlackUserId, updated); err != nil {
					fmt.Printf("warning: failed to send payment confirmed DM: %v\n", err)
				}
			}
		}
	}

	// Handle paid→sent rejection: notify client with pay button again
	if existing.Status == domain.InvoiceStatusPaid && status == domain.InvoiceStatusSent {
		if updated.BillingClientId != "" {
			client, err := u.clientRepository.GetClient(updated.BillingClientId)
			if err != nil {
				fmt.Printf("warning: failed to get client for notification: %v\n", err)
			} else if client.SlackUserId != "" {
				if err := infrastructure.SendPaymentRejectedDM(client.SlackUserId, updated); err != nil {
					fmt.Printf("warning: failed to send payment rejected DM: %v\n", err)
				}
			}
		}
	}

	return updated, nil
}

func (u *invoiceUseCase) DeleteInvoice(invoiceId string) error {
	existing, err := u.invoiceRepository.GetInvoice(invoiceId)
	if err != nil {
		return err
	}
	if existing.Status != domain.InvoiceStatusDraft {
		return fmt.Errorf("only draft invoices can be deleted")
	}
	return u.invoiceRepository.DeleteInvoice(invoiceId)
}

func validateStatusTransition(from, to domain.InvoiceStatus) error {
	allowed := map[domain.InvoiceStatus][]domain.InvoiceStatus{
		domain.InvoiceStatusDraft:     {domain.InvoiceStatusSent, domain.InvoiceStatusCancelled},
		domain.InvoiceStatusSent:      {domain.InvoiceStatusPaid, domain.InvoiceStatusCancelled},
		domain.InvoiceStatusPaid:      {domain.InvoiceStatusConfirmed, domain.InvoiceStatusSent},
		domain.InvoiceStatusConfirmed: {},
		domain.InvoiceStatusCancelled: {},
	}
	for _, s := range allowed[from] {
		if s == to {
			return nil
		}
	}
	return fmt.Errorf("invalid status transition from %s to %s", from, to)
}

func validatePermission(invoice domain.Invoice, targetStatus domain.InvoiceStatus, changedByUserId string) error {
	switch targetStatus {
	case domain.InvoiceStatusPaid:
		// sent→paid: only non-issuer (recipient) can mark as paid
		if changedByUserId == invoice.IssuerSlackUserId {
			return fmt.Errorf("issuer cannot mark their own invoice as paid")
		}
	case domain.InvoiceStatusConfirmed:
		// paid→confirmed: only issuer can confirm
		if changedByUserId != invoice.IssuerSlackUserId {
			return fmt.Errorf("only the issuer can confirm payment")
		}
	case domain.InvoiceStatusSent:
		// paid→sent (rejection): only issuer can reject
		if invoice.Status == domain.InvoiceStatusPaid && changedByUserId != invoice.IssuerSlackUserId {
			return fmt.Errorf("only the issuer can reject payment")
		}
	}
	return nil
}
