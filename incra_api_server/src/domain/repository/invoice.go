package repository

import "github.com/yuorei/incra_api_server/src/domain"

type InvoiceRepository interface {
	GetInvoice(invoiceId string) (domain.InvoiceResponse, error)
	CreateInvoice(invoice domain.Invoice) (domain.InvoiceResponse, error)
}
