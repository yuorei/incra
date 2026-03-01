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
	TransitionStatus(invoiceId string, status domain.InvoiceStatus, changedBy string) (domain.Invoice, error)
	DeleteInvoice(invoiceId string) error
}

type invoiceUseCase struct {
	invoiceRepository repository.InvoiceRepository
}

func NewInvoiceUseCase() InvoiceUseCase {
	return &invoiceUseCase{
		invoiceRepository: infrastructure.NewInvoiceRepository(),
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

func (u *invoiceUseCase) TransitionStatus(invoiceId string, status domain.InvoiceStatus, changedBy string) (domain.Invoice, error) {
	existing, err := u.invoiceRepository.GetInvoice(invoiceId)
	if err != nil {
		return domain.Invoice{}, err
	}
	if err := validateStatusTransition(existing.Status, status); err != nil {
		return domain.Invoice{}, err
	}
	updated, err := u.invoiceRepository.UpdateInvoiceStatus(invoiceId, status, changedBy)
	if err != nil {
		return domain.Invoice{}, err
	}
	if status == domain.InvoiceStatusSent {
		if err := infrastructure.SendPDFGenerateMessage(updated); err != nil {
			fmt.Printf("warning: failed to send PDF generate message: %v\n", err)
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
		domain.InvoiceStatusPaid:      {},
		domain.InvoiceStatusCancelled: {},
	}
	for _, s := range allowed[from] {
		if s == to {
			return nil
		}
	}
	return fmt.Errorf("invalid status transition from %s to %s", from, to)
}
