package repository

import "github.com/yuorei/incra_api_server/src/domain"

type InvoiceRepository interface {
	GetInvoice(invoiceId string) (domain.Invoice, error)
	ListInvoices(issuerSlackUserId string, status string, limit int, lastKey string) ([]domain.Invoice, string, error)
	ListReceivedInvoices(billingSlackUserId string, status string, limit int, lastKey string) ([]domain.Invoice, string, error)
	CreateInvoice(invoice domain.Invoice) (domain.Invoice, error)
	UpdateInvoice(invoice domain.Invoice) (domain.Invoice, error)
	UpdateInvoiceStatus(invoiceId string, status domain.InvoiceStatus, changedBy string) (domain.Invoice, error)
	DeleteInvoice(invoiceId string) error
	NextInvoiceNumber(year int) (string, error)
}
