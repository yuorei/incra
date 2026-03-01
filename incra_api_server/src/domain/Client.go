package domain

type Client struct {
	ClientId      string `json:"client_id" dynamodbav:"client_id"`
	Name          string `json:"name" dynamodbav:"name"`
	SlackUserId   string `json:"slack_user_id,omitempty" dynamodbav:"slack_user_id,omitempty"`
	SlackRealName string `json:"slack_real_name,omitempty" dynamodbav:"slack_real_name,omitempty"`
	Email         string `json:"email,omitempty" dynamodbav:"email,omitempty"`
	Phone         string `json:"phone,omitempty" dynamodbav:"phone,omitempty"`
	Address       string `json:"address,omitempty" dynamodbav:"address,omitempty"`
	BankDetails   string `json:"bank_details,omitempty" dynamodbav:"bank_details,omitempty"`
	RegisteredBy  string `json:"registered_by" dynamodbav:"registered_by"`
	CreatedAt     string `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt     string `json:"updated_at" dynamodbav:"updated_at"`
}
