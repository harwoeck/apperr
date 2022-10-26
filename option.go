package apperr

import (
	"time"

	"github.com/harwoeck/apperr/utils/dto"
	"github.com/harwoeck/apperr/utils/finalizer"
)

// Option provide functional modifiers for finalizer.Error instances.
type Option func(*finalizer.Error) error

// RequestInfo adds metadata about the request that clients can attach when
// filling a bug or providing other forms of feedback.
//
// The `requestID` should be an opaque non-confidential string. For example, it
// can be used to identify requests in the service's logs or across the
// infrastructure.
//
// `requestDuration` is the duration between the start and the end of this
// request. It can be useful to identify inconsistencies between latency and
// computation time on the server. It should not be specified for endpoints
// that perform cryptographic operations to prevent timing side channel
// attacks.
//
// `servingData` can be any data that was used to serve this request. For
// example, an encrypted stack trace that can be sent back to the service
// provider for debugging.
//
// `approximatedLatency` should be the approximated client-to-server latency.
func RequestInfo(requestID string, requestDuration *time.Time, servingData string, approximatedLatency *time.Duration) Option {
	return func(error *finalizer.Error) error {
		error.RequestInfo = &dto.RequestInfo{
			RequestID:           requestID,
			RequestDuration:     requestDuration,
			ServingData:         servingData,
			ApproximatedLatency: approximatedLatency,
		}
		return nil
	}
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
// `description` should explain what error  is encountered when accessing this
// resource. For example, updating a project may require the "writer"
// permission for the project. The description is only intended for client
// developers and should not be localized.
func ResourceInfo(resourceType string, name string, owner string, description string) Option {
	return func(error *finalizer.Error) error {
		error.ResourceInfo = &dto.ResourceInfo{
			Type:        resourceType,
			Name:        name,
			Owner:       owner,
			Description: description,
		}
		return nil
	}
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
func ErrorInfo(reason string, domain string, metadata map[string]string) Option {
	return func(error *finalizer.Error) error {
		error.ErrorInfo = &dto.ErrorInfo{
			Reason:   reason,
			Domain:   domain,
			Metadata: metadata,
		}
		return nil
	}
}

// Localize sets a unique message ID that can later be resolved using a
// LocalizationProvider. The message must be safe to return to the end user
// and should be printable for GUI applications.
func Localize(messageID string) Option {
	return func(error *finalizer.Error) error {
		error.RawLocalized = &dto.Localized{
			TextID: &messageID,
		}
		return nil
	}
}

// LocalizeAny is like Localize but provides an untyped object for the
// LocalizationProvider instead of a string message ID.
func LocalizeAny(any interface{}) Option {
	return func(error *finalizer.Error) error {
		error.RawLocalized = &dto.Localized{
			Any: any,
		}
		return nil
	}
}

// HelpLink provides URLs to documentation or for performing an out-of-band
// action. For example, if a quota check failed with an error indicating the
// calling project hasn't enabled the accessed service, this can contain a URL
// pointing directly to the right place in a dashboard to flip the bit.
//
// The `description` should explain what the link offers and is only intended
// for client developers and should not be localized.
func HelpLink(url string, description string) Option {
	return func(error *finalizer.Error) error {
		if error.HelpLinks == nil {
			error.HelpLinks = make([]*dto.HelpLink, 0)
		}
		error.HelpLinks = append(error.HelpLinks, &dto.HelpLink{
			URL:         url,
			Description: description,
		})
		return nil
	}
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
func FieldViolation(field string, description string) Option {
	return func(error *finalizer.Error) error {
		if error.RawFieldViolations == nil {
			error.RawFieldViolations = make([]*dto.FieldViolation, 0)
		}
		error.RawFieldViolations = append(error.RawFieldViolations, &dto.FieldViolation{
			Field:       field,
			Description: &description,
		})
		return nil
	}
}

// FieldViolationLocalize is like FieldViolation, but localizes the description
// using the descriptionID in the same way as Localize.
func FieldViolationLocalize(field string, descriptionID string) Option {
	return func(error *finalizer.Error) error {
		if error.RawFieldViolations == nil {
			error.RawFieldViolations = make([]*dto.FieldViolation, 0)
		}
		error.RawFieldViolations = append(error.RawFieldViolations, &dto.FieldViolation{
			Field:         field,
			DescriptionID: &descriptionID,
		})
		return nil
	}
}

// FieldViolationLocalizeAny is like FieldViolation, but localizes the
// description using the descriptionAny in the same way as LocalizeAny.
func FieldViolationLocalizeAny(field string, descriptionAny interface{}) Option {
	return func(error *finalizer.Error) error {
		if error.RawFieldViolations == nil {
			error.RawFieldViolations = make([]*dto.FieldViolation, 0)
		}
		error.RawFieldViolations = append(error.RawFieldViolations, &dto.FieldViolation{
			Field:          field,
			DescriptionAny: descriptionAny,
		})
		return nil
	}
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
func PreconditionViolation(violationType string, subject string, description string) Option {
	return func(error *finalizer.Error) error {
		if error.PreconditionViolations == nil {
			error.PreconditionViolations = make([]*dto.PreconditionViolation, 0)
		}
		error.PreconditionViolations = append(error.PreconditionViolations, &dto.PreconditionViolation{
			Type:        violationType,
			Subject:     subject,
			Description: description,
		})
		return nil
	}
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
func QuotaViolation(subject string, description string) Option {
	return func(error *finalizer.Error) error {
		if error.QuotaViolations == nil {
			error.QuotaViolations = make([]*dto.QuotaViolation, 0)
		}
		error.QuotaViolations = append(error.QuotaViolations, &dto.QuotaViolation{
			Subject:     subject,
			Description: description,
		})
		return nil
	}
}

// RetryInfo sets a minimum delay when the clients can retry a failed request.
// In general clients should always use this in combination with exponential
// backoff, e.g. if the first request after the `delay` timeout fails, clients
// should gradually increase the delay between retries, until either a maximum
// number of retries have been reached or a maximum retry delay cap has been
// reached.
func RetryInfo(delay time.Duration) Option {
	return func(error *finalizer.Error) error {
		error.RetryInfo = &dto.RetryInfo{
			Delay: delay,
		}
		return nil
	}
}
