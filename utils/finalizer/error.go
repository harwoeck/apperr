package finalizer

import (
	"golang.org/x/text/language"

	"github.com/harwoeck/apperr/utils/code"
	"github.com/harwoeck/apperr/utils/dto"
)

type Error struct {
	Code                   code.Code                    `json:"code"`
	Message                string                       `json:"message"`
	RawLocalized           *dto.Localized               `json:"-"`
	Localized              *Localized                   `json:"localized,omitempty"`
	RequestInfo            *dto.RequestInfo             `json:"requestInfo,omitempty"`
	ResourceInfo           *dto.ResourceInfo            `json:"resourceInfo,omitempty"`
	ErrorInfo              *dto.ErrorInfo               `json:"errorInfo,omitempty"`
	HelpLinks              []*dto.HelpLink              `json:"helpLinks,omitempty"`
	RawFieldViolations     []*dto.FieldViolation        `json:"-"`
	FieldViolations        []*FieldViolation            `json:"fieldViolations,omitempty"`
	PreconditionViolations []*dto.PreconditionViolation `json:"preconditionViolations,omitempty"`
	QuotaViolations        []*dto.QuotaViolation        `json:"quotaViolations,omitempty"`
	RetryInfo              *dto.RetryInfo               `json:"retryInfo,omitempty"`
}

type Localized struct {
	Locale language.Tag `json:"locale"`
	Title  string       `json:"title"`
	Text   string       `json:"text"`
}

type FieldViolation struct {
	Locale      *language.Tag `json:"locale,omitempty"`
	Field       string        `json:"field"`
	Description string        `json:"description"`
}
