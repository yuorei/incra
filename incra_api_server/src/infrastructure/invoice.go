package infrastructure

import (
	"github.com/yuorei/incra_api_server/src/domain"
	"github.com/yuorei/incra_api_server/src/domain/repository"
)

type InvoiceRepository struct {
}

func NewInvoiceRepository() repository.InvoiceRepository {
	return &InvoiceRepository{}
}

func (r *InvoiceRepository) GetInvoice(invoiceId string) (domain.InvoiceResponse, error) {
	return domain.InvoiceResponse{}, nil
}
