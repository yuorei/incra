package usecase

import (
	"github.com/yuorei/incra_api_server/src/domain"
	"github.com/yuorei/incra_api_server/src/domain/repository"
	"github.com/yuorei/incra_api_server/src/infrastructure"
)

type InvoiceUseCase interface {
	GetInvoice(invoiceId string) (domain.InvoiceResponse, error)
	CreateInvoice(invoice domain.Invoice) (domain.InvoiceResponse, error)
}

type invoiceUseCase struct {
	invoiceRepository repository.InvoiceRepository
}

func NewInvoiceUseCase() InvoiceUseCase {
	return &invoiceUseCase{
		invoiceRepository: infrastructure.NewInvoiceRepository(),
	}
}

func (u *invoiceUseCase) GetInvoice(invoiceId string) (domain.InvoiceResponse, error) {
	return domain.InvoiceResponse{}, nil
}

func (u *invoiceUseCase) CreateInvoice(invoice domain.Invoice) (domain.InvoiceResponse, error) {
	invoiceResponse, err := u.invoiceRepository.CreateInvoice(invoice)
	if err != nil {
		return domain.InvoiceResponse{}, err
	}
	return domain.InvoiceResponse{
		InvoiceId: invoiceResponse.InvoiceId,
	}, nil
}
