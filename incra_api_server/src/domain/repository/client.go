package repository

import "github.com/yuorei/incra_api_server/src/domain"

type ClientRepository interface {
	GetClient(clientId string) (domain.Client, error)
	GetClientBySlackUserId(slackUserId string) (domain.Client, error)
	ListClients(registeredBy string) ([]domain.Client, error)
	CreateClient(client domain.Client) (domain.Client, error)
	UpdateClient(client domain.Client) (domain.Client, error)
	DeleteClient(clientId string) error
}
