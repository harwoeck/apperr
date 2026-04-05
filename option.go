package apperr

import (
	"net/http"
	"time"

	"github.com/harwoeck/apperr/errdetails"
)

// Option is the interface implemented by all error detail options. Each
// concrete option type (e.g. ResourceInfoOption, ErrorInfoOption) satisfies
// Option but is also a distinct type, allowing packages like stricterr to
// require specific detail types at compile time.
//
// Options are applied in order during resolution. For scalar detail types
// (RequestInfo, ResourceInfo, ErrorInfo, Localize, RetryInfo) the last
// option of a given type wins — earlier values are silently overwritten.
// For slice-based detail types (HelpLink, FieldViolation,
// PreconditionViolation, QuotaViolation) values are appended and all
// entries are preserved.
type Option interface {
	// Apply applies this option to the UnresolvedError being built.
	Apply(*errdetails.UnresolvedError) error
}

// WrapOption wraps an underlying error as the cause of the AppError.
// It is extracted during construction and does not implement Apply in a
// meaningful way — it is never applied to a ResolvedError.
//
// The wrapped error participates in errors.Is and errors.As chains via
// the AppError.Unwrap method. The wrapped error's message is included in
// AppError.Error() output for logging/debugging but is NOT exposed in the
// resolved error sent to clients (to avoid leaking internal details).
type WrapOption struct {
	Err error
}

// Apply is a no-op. WrapOption is consumed by the constructor, not during
// resolution.
func (o WrapOption) Apply(*errdetails.UnresolvedError) error { return nil }

// Wrap returns an Option that wraps the given error as the underlying
// cause of the AppError. This enables errors.Is and errors.As to
// traverse the error chain.
//
//	apperr.NotFound("user not found", apperr.Wrap(sql.ErrNoRows))
func Wrap(err error) WrapOption {
	return WrapOption{Err: err}
}

// RequestInfoOption is an Option that sets request metadata.
type RequestInfoOption struct {
	RequestID   string
	ServingData string
}

// Apply implements Option.
func (o RequestInfoOption) Apply(ue *errdetails.UnresolvedError) error {
	ue.RequestInfo = &errdetails.RequestInfo{
		RequestID:   o.RequestID,
		ServingData: o.ServingData,
	}
	return nil
}

// RequestInfo adds metadata about the request that clients can attach when
// filling a bug or providing other forms of feedback.
//
// The `requestID` should be an opaque non-confidential string. For example, it
// can be used to identify requests in the service's logs or across the
// infrastructure.
//
// `servingData` can be any data that was used to serve this request. For
// example, an encrypted stack trace that can be sent back to the service
// provider for debugging.
func RequestInfo(requestID string, servingData string) RequestInfoOption {
	return RequestInfoOption{
		RequestID:   requestID,
		ServingData: servingData,
	}
}

// ResourceInfoOption is an Option that sets the resource being accessed.
type ResourceInfoOption struct {
	ResourceType string
	Name         string
	Owner        string
	Description  string
}

// Apply implements Option.
func (o ResourceInfoOption) Apply(ue *errdetails.UnresolvedError) error {
	ue.ResourceInfo = &errdetails.ResourceInfo{
		Type:        o.ResourceType,
		Name:        o.Name,
		Owner:       o.Owner,
		Description: o.Description,
	}
	return nil
}

// ResourceInfo adds information about the resource being accessed.
//
// The `resourceType` should be a unique name of the resource, e.g.
// "example.com/store.v1.Book".
//
// The `name` must be the unique identifier of the resource being accessed.
//
// `owner` can be populated if it doesn't impose any security and privacy
// risks, e.g. the ownership is public knowledge anyway.
//
// `description` should explain what error is encountered when accessing this
// resource. For example, updating a project may require the "writer"
// permission for the project. The description is only intended for client
// developers and should not be localized.
func ResourceInfo(resourceType string, name string, owner string, description string) ResourceInfoOption {
	return ResourceInfoOption{
		ResourceType: resourceType,
		Name:         name,
		Owner:        owner,
		Description:  description,
	}
}

// ErrorInfoOption is an Option that sets structured error cause details.
type ErrorInfoOption struct {
	Reason   string
	Domain   string
	Metadata map[string]string
}

// Apply implements Option.
func (o ErrorInfoOption) Apply(ue *errdetails.UnresolvedError) error {
	ue.ErrorInfo = &errdetails.ErrorInfo{
		Reason:   o.Reason,
		Domain:   o.Domain,
		Metadata: o.Metadata,
	}
	return nil
}

// ErrorInfo should describe the cause of the error with more structured
// details.
//
// The `reason` should be a constant value that identifies the proximate cause
// of the errors.
//
// The `domain` refers to the logical grouping to which the reason belongs. The
// value is typically the registered service name of the service generating the
// error, like "api.store.example.com". The domain should be a globally unique
// value and should be constant within the service infrastructure.
//
// `metadata` can attach further structured meta information to the error. The
// key must not exceed 64 characters in length.
func ErrorInfo(reason string, domain string, metadata map[string]string) ErrorInfoOption {
	return ErrorInfoOption{
		Reason:   reason,
		Domain:   domain,
		Metadata: metadata,
	}
}

// LocalizeOption is an Option that sets a message ID for localization.
type LocalizeOption struct {
	TextID *string
	Any    any
}

// Apply implements Option.
func (o LocalizeOption) Apply(ue *errdetails.UnresolvedError) error {
	ue.RawLocalized = &errdetails.RawLocalized{
		TextID: o.TextID,
		Any:    o.Any,
	}
	return nil
}

// Localize sets a unique message ID that can later be resolved using a
// LocalizationProvider. The message must be safe to return to the end user
// and should be printable for GUI applications.
func Localize(messageID string) LocalizeOption {
	return LocalizeOption{TextID: &messageID}
}

// LocalizeAny is like Localize but provides an untyped object for the
// LocalizationProvider instead of a string message ID.
func LocalizeAny(v any) LocalizeOption {
	return LocalizeOption{Any: v}
}

// HelpLinkOption is an Option that appends a help link.
type HelpLinkOption struct {
	URL         string
	Description string
}

// Apply implements Option.
func (o HelpLinkOption) Apply(ue *errdetails.UnresolvedError) error {
	ue.HelpLinks = append(ue.HelpLinks, &errdetails.HelpLink{
		URL:         o.URL,
		Description: o.Description,
	})
	return nil
}

// HelpLink provides URLs to documentation or for performing an out-of-band
// action. For example, if a quota check failed with an error indicating the
// calling project hasn't enabled the accessed service, this can contain a URL
// pointing directly to the right place in a dashboard to flip the bit.
//
// The `description` should explain what the link offers and is only intended
// for client developers and should not be localized.
func HelpLink(url string, description string) HelpLinkOption {
	return HelpLinkOption{
		URL:         url,
		Description: description,
	}
}

// FieldViolationOption is an Option that appends a field violation.
type FieldViolationOption struct {
	Field          string
	Description    *string
	DescriptionID  *string
	DescriptionAny any
}

// Apply implements Option.
func (o FieldViolationOption) Apply(ue *errdetails.UnresolvedError) error {
	ue.RawFieldViolations = append(ue.RawFieldViolations, &errdetails.RawFieldViolation{
		Field:          o.Field,
		Description:    o.Description,
		DescriptionID:  o.DescriptionID,
		DescriptionAny: o.DescriptionAny,
	})
	return nil
}

// FieldViolation describes a single bad request field in a client request.
//
// The `field` must focus on the syntactic aspects of the request, e.g. a path
// leading to the field in the response body, like "book.author_id". The path
// in the field value must be a sequence of dot-separated identifiers.
//
// The `description` should explain why the request element is bad. The value
// must be safe to return to the end user and should be printable for GUI
// applications.
func FieldViolation(field string, description string) FieldViolationOption {
	return FieldViolationOption{
		Field:       field,
		Description: &description,
	}
}

// FieldViolationLocalize is like FieldViolation, but localizes the description
// using the descriptionID in the same way as Localize.
func FieldViolationLocalize(field string, descriptionID string) FieldViolationOption {
	return FieldViolationOption{
		Field:         field,
		DescriptionID: &descriptionID,
	}
}

// FieldViolationLocalizeAny is like FieldViolation, but localizes the
// description using the descriptionAny in the same way as LocalizeAny.
func FieldViolationLocalizeAny(field string, descriptionAny any) FieldViolationOption {
	return FieldViolationOption{
		Field:          field,
		DescriptionAny: descriptionAny,
	}
}

// PreconditionViolationOption is an Option that appends a precondition
// violation.
type PreconditionViolationOption struct {
	ViolationType string
	Subject       string
	Description   string
}

// Apply implements Option.
func (o PreconditionViolationOption) Apply(ue *errdetails.UnresolvedError) error {
	ue.PreconditionViolations = append(ue.PreconditionViolations, &errdetails.PreconditionViolation{
		Type:        o.ViolationType,
		Subject:     o.Subject,
		Description: o.Description,
	})
	return nil
}

// PreconditionViolation describes a single precondition violation. For
// example, conflicting object revisions during an update call.
//
// The `violationType` should be a service-specific enum type to define the
// supported precondition violation subjects. For example, "UNKNOWN_AUTHOR".
//
// The `subject` references the object, relative to the type, that failed, like
// "book.author".
//
// The `description` should explain how the precondition failed. Developers can
// use this description to understand how to fix the failure.
func PreconditionViolation(violationType string, subject string, description string) PreconditionViolationOption {
	return PreconditionViolationOption{
		ViolationType: violationType,
		Subject:       subject,
		Description:   description,
	}
}

// QuotaViolationOption is an Option that appends a quota violation.
type QuotaViolationOption struct {
	Subject     string
	Description string
}

// Apply implements Option.
func (o QuotaViolationOption) Apply(ue *errdetails.UnresolvedError) error {
	ue.QuotaViolations = append(ue.QuotaViolations, &errdetails.QuotaViolation{
		Subject:     o.Subject,
		Description: o.Description,
	})
	return nil
}

// QuotaViolation describes a single quota violation. For example, a daily
// quota or a custom quota that was exceeded.
//
// The subject must reference the object on which the quota check failed.
// For example, "ip:<ip address of client>" or "project:<project id>".
//
// The description should contain more information about how the quota check
// failed. Clients can use this description to find more about the quota
// configuration in the service's public documentation. For example: "Service
// disabled" or "Daily Limit for read operations exceeded".
func QuotaViolation(subject string, description string) QuotaViolationOption {
	return QuotaViolationOption{
		Subject:     subject,
		Description: description,
	}
}

// RetryInfoOption is an Option that sets retry delay information.
type RetryInfoOption struct {
	Delay time.Duration
}

// Apply implements Option. The Duration is converted to milliseconds for
// the resolved RetryInfo.
func (o RetryInfoOption) Apply(ue *errdetails.UnresolvedError) error {
	ue.RetryInfo = &errdetails.RetryInfo{
		Delay: int(o.Delay.Milliseconds()),
	}
	return nil
}

// RetryInfo sets a minimum delay when the clients can retry a failed request.
// The delay is stored as milliseconds in the resolved error detail. In
// general clients should always use this in combination with exponential
// backoff, e.g. if the first request after the `delay` timeout fails, clients
// should gradually increase the delay between retries, until either a maximum
// number of retries have been reached or a maximum retry delay cap has been
// reached.
func RetryInfo(delay time.Duration) RetryInfoOption {
	return RetryInfoOption{Delay: delay}
}

// RetryInfoPtr is a convenience wrapper that returns a pointer to a
// RetryInfoOption. This is useful when a *RetryInfoOption parameter is
// optional and you want to provide a value inline.
func RetryInfoPtr(delay time.Duration) *RetryInfoOption {
	o := RetryInfo(delay)
	return &o
}

// DebugInfoOption is an Option that sets debugging information.
type DebugInfoOption struct {
	StackEntries []string
	Detail       string
}

// Apply implements Option.
func (o DebugInfoOption) Apply(ue *errdetails.UnresolvedError) error {
	ue.DebugInfo = &errdetails.DebugInfo{
		StackEntries: o.StackEntries,
		Detail:       o.Detail,
	}
	return nil
}

// DebugInfo attaches server-side debugging information such as stack traces.
// This detail type should only be used during development or when the
// information is safe to expose; it must NOT be returned to untrusted
// clients in production.
//
// `stackEntries` is a list of stack trace entries indicating where the
// error occurred.
//
// `detail` is additional debugging information provided by the server.
func DebugInfo(stackEntries []string, detail string) DebugInfoOption {
	return DebugInfoOption{
		StackEntries: stackEntries,
		Detail:       detail,
	}
}

// HeadersOption is an Option that adds HTTP headers to the resolved error.
// Unlike most scalar detail types, HeadersOption uses aggregate semantics:
// values from multiple HeadersOption instances are merged (via
// [http.Header.Add]) rather than overwritten. This allows separate parts
// of the application to contribute headers independently.
type HeadersOption http.Header

// Apply implements Option. Headers are aggregated: values from multiple
// HeadersOption instances are merged rather than overwritten.
func (o HeadersOption) Apply(ue *errdetails.UnresolvedError) error {
	if ue.Headers == nil {
		ue.Headers = make(http.Header)
	}
	for key, values := range http.Header(o) {
		for _, v := range values {
			ue.Headers.Add(key, v)
		}
	}
	return nil
}

// Headers returns an Option that attaches the given HTTP headers to the
// error. These headers are forwarded by transport-layer interceptors and
// middlewares (e.g. set on the [http.ResponseWriter] by httperr, or added
// to [connect.Error] metadata by connecterr).
//
// Multiple Headers options are aggregated — all entries are preserved.
// This differs from scalar detail types where the last option wins.
func Headers(headers http.Header) Option {
	return HeadersOption(headers)
}
