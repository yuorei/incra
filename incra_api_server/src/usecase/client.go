package usecase

import (
	"github.com/yuorei/incra_api_server/src/domain"
	"github.com/yuorei/incra_api_server/src/domain/repository"
	"github.com/yuorei/incra_api_server/src/infrastructure"
)

type ClientUseCase interface {
	GetClient(clientId string) (domain.Client, error)
	GetClientBySlackUserId(slackUserId string) (domain.Client, error)
	ListClients(registeredBy string) ([]domain.Client, error)
	CreateClient(client domain.Client) (domain.Client, error)
	UpdateClient(client domain.Client) (domain.Client, error)
	DeleteClient(clientId string) error
}

type clientUseCase struct {
	clientRepository repository.ClientRepository
}

func NewClientUseCase() ClientUseCase {
	return &clientUseCase{
		clientRepository: infrastructure.NewClientRepository(),
	}
}

func (u *clientUseCase) GetClient(clientId string) (domain.Client, error) {
	return u.clientRepository.GetClient(clientId)
}

func (u *clientUseCase) GetClientBySlackUserId(slackUserId string) (domain.Client, error) {
	return u.clientRepository.GetClientBySlackUserId(slackUserId)
}

func (u *clientUseCase) ListClients(registeredBy string) ([]domain.Client, error) {
	return u.clientRepository.ListClients(registeredBy)
}

func (u *clientUseCase) CreateClient(client domain.Client) (domain.Client, error) {
	return u.clientRepository.CreateClient(client)
}

func (u *clientUseCase) UpdateClient(client domain.Client) (domain.Client, error) {
	return u.clientRepository.UpdateClient(client)
}

func (u *clientUseCase) DeleteClient(clientId string) error {
	return u.clientRepository.DeleteClient(clientId)
}
