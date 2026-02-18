package errdetails

import (
	"golang.org/x/text/language"
)

// RawLocalized holds the unresolved localization input. Either TextID or Any
// is set by the caller; the resolve step translates them into a
// LocalizedMessage.
type RawLocalized struct {
	TextID *string `json:"-"`
	Any    any     `json:"-"`
}

// RequestInfo contains metadata about the request.
type RequestInfo struct {
	RequestID   string `json:"requestId"`
	ServingData string `json:"servingData"`
}

// ResourceInfo describes the resource being accessed.
type ResourceInfo struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Owner       string `json:"owner"`
	Description string `json:"description"`
}

// ErrorInfo describes the cause of the error with structured details.
type ErrorInfo struct {
	Reason   string            `json:"reason"`
	Domain   string            `json:"domain"`
	Metadata map[string]string `json:"metadata"`
}

// HelpLink is a URL pointing to documentation or an out-of-band action.
type HelpLink struct {
	URL         string `json:"url"`
	Description string `json:"description"`
}

// RawFieldViolation holds the unresolved field violation input.
type RawFieldViolation struct {
	Field          string  `json:"-"`
	Description    *string `json:"-"`
	DescriptionID  *string `json:"-"`
	DescriptionAny any     `json:"-"`
}

// PreconditionViolation describes a single precondition failure.
type PreconditionViolation struct {
	Type        string `json:"type"`
	Subject     string `json:"subject"`
	Description string `json:"description"`
}

// QuotaViolation describes a single quota failure.
type QuotaViolation struct {
	Subject     string `json:"subject"`
	Description string `json:"description"`
}

// RetryInfo advises the client on when to retry. The Delay field is
// expressed in milliseconds.
type RetryInfo struct {
	Delay int `json:"delay"`
}

// DebugInfo describes additional debugging information provided by the
// server, such as stack traces. This detail type should only be used
// during development or when the information is safe to expose; it must
// NOT be returned to untrusted clients in production.
type DebugInfo struct {
	StackEntries []string `json:"stackEntries,omitempty"`
	Detail       string   `json:"detail"`
}

// UnresolvedError is the accumulation target used during option application
// and localization. It carries both the raw intermediate fields (RawLocalized,
// RawFieldViolations) and the fields that are already resolved at apply time.
// After resolution the resolve package projects it into a [ResolvedError],
// which contains only the final consumer-facing fields.
type UnresolvedError struct {
	Code                   Code
	Message                string
	RawLocalized           *RawLocalized
	Localized              *LocalizedMessage
	RequestInfo            *RequestInfo
	ResourceInfo           *ResourceInfo
	ErrorInfo              *ErrorInfo
	HelpLinks              []*HelpLink
	RawFieldViolations     []*RawFieldViolation
	FieldViolations        []*ResolvedFieldViolation
	PreconditionViolations []*PreconditionViolation
	QuotaViolations        []*QuotaViolation
	RetryInfo              *RetryInfo
	DebugInfo              *DebugInfo
}

// ResolvedError is the fully resolved form of an AppError, ready for
// conversion to a framework-specific error type by an x/ converter.
type ResolvedError struct {
	Code                   Code                      `json:"code"`
	Message                string                    `json:"message"`
	Localized              *LocalizedMessage         `json:"localized,omitempty"`
	RequestInfo            *RequestInfo              `json:"requestInfo,omitempty"`
	ResourceInfo           *ResourceInfo             `json:"resourceInfo,omitempty"`
	ErrorInfo              *ErrorInfo                `json:"errorInfo,omitempty"`
	HelpLinks              []*HelpLink               `json:"helpLinks,omitempty"`
	FieldViolations        []*ResolvedFieldViolation `json:"fieldViolations,omitempty"`
	PreconditionViolations []*PreconditionViolation  `json:"preconditionViolations,omitempty"`
	QuotaViolations        []*QuotaViolation         `json:"quotaViolations,omitempty"`
	RetryInfo              *RetryInfo                `json:"retryInfo,omitempty"`
	DebugInfo              *DebugInfo                `json:"debugInfo,omitempty"`
}

// LocalizedMessage is a translated user-facing message.
type LocalizedMessage struct {
	Locale language.Tag `json:"locale"`
	Text   string       `json:"text"`
}

// ResolvedFieldViolation is a field violation with its description resolved
// (possibly translated).
type ResolvedFieldViolation struct {
	Locale      *language.Tag `json:"locale,omitempty"`
	Field       string        `json:"field"`
	Description string        `json:"description"`
}
