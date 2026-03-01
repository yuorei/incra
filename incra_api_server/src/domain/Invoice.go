package domain

import "time"

type InvoiceStatus string

const (
	InvoiceStatusDraft     InvoiceStatus = "draft"
	InvoiceStatusSent      InvoiceStatus = "sent"
	InvoiceStatusPaid      InvoiceStatus = "paid"
	InvoiceStatusCancelled InvoiceStatus = "cancelled"
)

type InvoiceItem struct {
	Date        string `json:"date" dynamodbav:"date"`
	Description string `json:"description" dynamodbav:"description"`
	Quantity    int    `json:"quantity" dynamodbav:"quantity"`
	UnitPrice   int    `json:"unit_price" dynamodbav:"unit_price"`
	Amount      int    `json:"amount" dynamodbav:"amount"`
	Memo        string `json:"memo,omitempty" dynamodbav:"memo,omitempty"`
}

type HistoryEntry struct {
	ChangedAt string `json:"changed_at" dynamodbav:"changed_at"`
	OldStatus string `json:"old_status" dynamodbav:"old_status"`
	NewStatus string `json:"new_status" dynamodbav:"new_status"`
	ChangedBy string `json:"changed_by" dynamodbav:"changed_by"`
}

type Invoice struct {
	InvoiceId           string         `json:"invoice_id" dynamodbav:"invoice_id"`
	Status              InvoiceStatus  `json:"status" dynamodbav:"status"`
	BillingClientId     string         `json:"billing_client_id" dynamodbav:"billing_client_id"`
	BillingClientName   string         `json:"billing_client_name" dynamodbav:"billing_client_name"`
	TotalAmount         int            `json:"total_amount" dynamodbav:"total_amount"`
	DueDate             string         `json:"due_date" dynamodbav:"due_date"`
	BankDetails         string         `json:"bank_details" dynamodbav:"bank_details"`
	AdditionalInfo      string         `json:"additional_info,omitempty" dynamodbav:"additional_info,omitempty"`
	PdfUrl              string         `json:"pdf_url,omitempty" dynamodbav:"pdf_url,omitempty"`
	Items               []InvoiceItem  `json:"items" dynamodbav:"items"`
	History             []HistoryEntry `json:"history,omitempty" dynamodbav:"history,omitempty"`
	IssuerSlackUserId   string         `json:"issuer_slack_user_id" dynamodbav:"issuer_slack_user_id"`
	IssuerSlackRealName string         `json:"issuer_slack_real_name" dynamodbav:"issuer_slack_real_name"`
	PaidAmount          int            `json:"paid_amount,omitempty" dynamodbav:"paid_amount,omitempty"`
	PaidDate            string         `json:"paid_date,omitempty" dynamodbav:"paid_date,omitempty"`
	PaidMethod          string         `json:"paid_method,omitempty" dynamodbav:"paid_method,omitempty"`
	CreatedAt           string         `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt           string         `json:"updated_at" dynamodbav:"updated_at"`
}

// InvoiceResponse is an alias for Invoice for backward compatibility
type InvoiceResponse = Invoice

var _ = time.Now
