package usecase

import "github.com/yuorei/incra_api_server/src/domain"

type InvoiceUseCase interface {
	GetInvoice(invoiceId string) (domain.InvoiceResponse, error)
}

type invoiceUseCase struct {
}

func NewInvoiceUseCase() InvoiceUseCase {
	return &invoiceUseCase{}
}

func (u *invoiceUseCase) GetInvoice(invoiceId string) (domain.InvoiceResponse, error) {
	return domain.InvoiceResponse{}, nil
}
