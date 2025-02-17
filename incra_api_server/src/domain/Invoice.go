package domain

import "time"

type Invoice struct {
	// AdditionalInfo その他の説明
	AdditionalInfo *string `json:"additional_info,omitempty"`

	// BillingSlackRealName 請求先のSlackの本名
	BillingSlackRealName *string `json:"billing_slack_real_name,omitempty"`

	// BillingSlackUserId 請求先のSlackのユーザーID
	BillingSlackUserId *string `json:"billing_slack_user_id,omitempty"`

	// InvoiceId 請求書ID
	InvoiceId *string `json:"invoice_id,omitempty"`

	// IssuerSlackRealName 請求者のSlackの本名
	IssuerSlackRealName *string `json:"issuer_slack_real_name,omitempty"`

	// IssuerSlackUserId 請求者のSlackのユーザーID
	IssuerSlackUserId *string `json:"issuer_slack_user_id,omitempty"`

	// PaidAmount 支払い金額 (円)
	PaidAmount *int `json:"paid_amount,omitempty"`

	// PaidDate 支払い日 (yyyyddmm)
	PaidDate *time.Time `json:"paid_date,omitempty"`

	// PaidMethod 支払い方法
	PaidMethod *string `json:"paid_method,omitempty"`

	// Status 請求書のステータス
	Status *string `json:"status,omitempty"`
}


type InvoiceResponse struct {
	// AdditionalInfo その他の説明
	AdditionalInfo *string `json:"additional_info,omitempty"`

	// BillingSlackRealName 請求先のSlackの本名
	BillingSlackRealName *string `json:"billing_slack_real_name,omitempty"`

	// BillingSlackUserId 請求先のSlackのユーザーID
	BillingSlackUserId *string `json:"billing_slack_user_id,omitempty"`

	// InvoiceId 請求書ID
	InvoiceId *string `json:"invoice_id,omitempty"`

	// IssuerSlackRealName 請求者のSlackの本名
	IssuerSlackRealName *string `json:"issuer_slack_real_name,omitempty"`

	// IssuerSlackUserId 請求者のSlackのユーザーID
	IssuerSlackUserId *string `json:"issuer_slack_user_id,omitempty"`

	// PaidAmount 支払い金額 (円)
	PaidAmount *int `json:"paid_amount,omitempty"`

	// PaidDate 支払い日 (yyyyddmm)
	PaidDate *time.Time `json:"paid_date,omitempty"`

	// PaidMethod 支払い方法
	PaidMethod *string `json:"paid_method,omitempty"`

	// Status 請求書のステータス
	Status *string `json:"status,omitempty"`
}
