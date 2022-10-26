package dto

import (
	"time"
)

type Localized struct {
	TextID *string     `json:"-"`
	Any    interface{} `json:"-"`
}

type RequestInfo struct {
	RequestID           string         `json:"requestId"`
	RequestDuration     *time.Time     `json:"requestDuration,omitempty"`
	ServingData         string         `json:"servingData"`
	ApproximatedLatency *time.Duration `json:"approximatedLatency,omitempty"`
}

type ResourceInfo struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Owner       string `json:"owner"`
	Description string `json:"description"`
}

type ErrorInfo struct {
	Reason   string            `json:"reason"`
	Domain   string            `json:"domain"`
	Metadata map[string]string `json:"metadata"`
}

type HelpLink struct {
	URL         string `json:"url"`
	Description string `json:"description"`
}

type FieldViolation struct {
	Field          string      `json:"-"`
	Description    *string     `json:"-"`
	DescriptionID  *string     `json:"-"`
	DescriptionAny interface{} `json:"-"`
}

type PreconditionViolation struct {
	Type        string `json:"type"`
	Subject     string `json:"subject"`
	Description string `json:"description"`
}

type QuotaViolation struct {
	Subject     string `json:"subject"`
	Description string `json:"description"`
}

type RetryInfo struct {
	Delay time.Duration `json:"delay"`
}
