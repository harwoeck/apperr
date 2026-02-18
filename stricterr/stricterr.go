// Package stricterr provides an [apperr.AppError] factory that enforces
// [AIP-193]'s requirement that every error carries an [errdetails.ErrorInfo]
// with a domain, reason and optional metadata. For error codes that AIP-193
// associates with specific detail types, the corresponding detail is required
// as a typed parameter at compile time:
//
//   - [Factory.NotFound], [Factory.AlreadyExists]: require [apperr.ResourceInfoOption]
//   - [Factory.InvalidArgument]: requires []FieldViolationOption
//   - [Factory.ResourceExhausted]: requires []QuotaViolationOption
//   - [Factory.FailedPrecondition]: requires []PreconditionViolationOption
//   - [Factory.Unavailable]: accepts an optional *RetryInfoOption
//
// Usage:
//
//	// In your main.go or a shared package:
//	var errors = stricterr.NewFactory("api.store.example.com")
//
//	// At every call site (reason + metadata are now required):
//	return errors.NotFound("book not found", "BOOK_NOT_FOUND", nil,
//	    apperr.ResourceInfo("store.v1.Book", bookID, "", ""),
//	    apperr.Localize("BOOK_NOT_FOUND"))
//
// In environments with generated code (e.g. protobuf/gRPC), the magic strings
// can be replaced with typed constants:
//
//	return errors.NotFound("book not found", rpc.BookResourceErrNotFound, nil,
//	    apperr.ResourceInfo(rpc.BookResource, bookID, "", ""),
//	    apperr.Localize(rpc.BookResourceErrNotFound))
//
// When the reason key doubles as the localization lookup key (common in
// greenfield projects), use [LocalizeWithReason] to avoid repeating it:
//
//	var errors = stricterr.NewFactory("api.store.example.com", stricterr.LocalizeWithReason())
//
//	// apperr.Localize(reason) is now added automatically:
//	return errors.NotFound("book not found", rpc.BookResourceErrNotFound, nil,
//	    apperr.ResourceInfo(rpc.BookResource, bookID, "", ""))
//
// Since options are applied in order and the last write wins, a caller can
// still override the automatic localization by passing an explicit
// apperr.Localize or apperr.LocalizeAny in opts.
//
// [AIP-193]: https://google.aip.dev/193
package stricterr

import (
	"github.com/harwoeck/apperr"
	"github.com/harwoeck/apperr/errdetails"
)

// FactoryOption configures a [Factory].
type FactoryOption func(*Factory)

// LocalizeWithReason returns a [FactoryOption] that automatically appends
// apperr.Localize(reason) to every error produced by the factory. This is
// useful when the ErrorInfo reason key is also used as the localization
// message ID, avoiding repetition at every call site.
func LocalizeWithReason() FactoryOption {
	return func(f *Factory) {
		f.localizeWithReason = true
	}
}

// Factory produces [apperr.AppError] values that always carry an
// [errdetails.ErrorInfo] populated with the configured domain.
type Factory struct {
	domain             string
	localizeWithReason bool
}

// NewFactory returns a new Factory bound to the given ErrorInfo domain.
// The domain should be a globally unique value, typically the registered
// service name, like "api.store.example.com".
func NewFactory(domain string, opts ...FactoryOption) *Factory {
	f := &Factory{domain: domain}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

// Enrich adds an ErrorInfo with the factory's domain to an existing
// *apperr.AppError. This is useful when you receive an AppError from a
// lower-level package that doesn't use stricterr and want to attach
// structured cause information.
//
// Enrich mutates appErr in place; it does not create a new error.
func (f *Factory) Enrich(appErr *apperr.AppError, reason string, metadata map[string]string) {
	appErr.AppendOptions(apperr.ErrorInfo(reason, f.domain, metadata))
}

// build prepends an ErrorInfo option and delegates to the base apperr
// constructor for the given code.
func (f *Factory) build(code errdetails.Code, msg, reason string, metadata map[string]string, extra []apperr.Option) *apperr.AppError {
	opts := []apperr.Option{apperr.ErrorInfo(reason, f.domain, metadata)}
	if f.localizeWithReason {
		opts = append(opts, apperr.Localize(reason))
	}
	opts = append(opts, extra...)

	switch code {
	case errdetails.Canceled:
		return apperr.Canceled(msg, opts...)
	case errdetails.Unknown:
		return apperr.Unknown(msg, opts...)
	case errdetails.InvalidArgument:
		return apperr.InvalidArgument(msg, opts...)
	case errdetails.DeadlineExceeded:
		return apperr.DeadlineExceeded(msg, opts...)
	case errdetails.NotFound:
		return apperr.NotFound(msg, opts...)
	case errdetails.AlreadyExists:
		return apperr.AlreadyExists(msg, opts...)
	case errdetails.PermissionDenied:
		return apperr.PermissionDenied(msg, opts...)
	case errdetails.ResourceExhausted:
		return apperr.ResourceExhausted(msg, opts...)
	case errdetails.FailedPrecondition:
		return apperr.FailedPrecondition(msg, opts...)
	case errdetails.Aborted:
		return apperr.Aborted(msg, opts...)
	case errdetails.OutOfRange:
		return apperr.OutOfRange(msg, opts...)
	case errdetails.Unimplemented:
		return apperr.Unimplemented(msg, opts...)
	case errdetails.Internal:
		return apperr.Internal(msg, opts...)
	case errdetails.Unavailable:
		return apperr.Unavailable(msg, opts...)
	case errdetails.DataLoss:
		return apperr.DataLoss(msg, opts...)
	case errdetails.Unauthenticated:
		return apperr.Unauthenticated(msg, opts...)
	default:
		panic("unreachable code: invalid error code") // should never happen since code is an enum
	}
}

// Canceled indicates the operation was canceled (typically by the caller).
//
// The `reason` should be a constant value that identifies the proximate cause
// of the error. `metadata` can attach further structured meta information to
// the error. The key must not exceed 64 characters in length.
//
// (Further resources: https://google.aip.dev/193)
func (f *Factory) Canceled(msg, reason string, metadata map[string]string, opts ...apperr.Option) *apperr.AppError {
	return f.build(errdetails.Canceled, msg, reason, metadata, opts)
}

// Unknown error. An example of where this error may be returned is
// if a Status value received from another address space belongs to
// an error-space that is not known in this address space. Also,
// errors raised by APIs that do not return enough error information
// may be converted to this error.
//
// The `reason` should be a constant value that identifies the proximate cause
// of the error. `metadata` can attach further structured meta information to
// the error. The key must not exceed 64 characters in length.
//
// (Further resources: https://google.aip.dev/193)
func (f *Factory) Unknown(msg, reason string, metadata map[string]string, opts ...apperr.Option) *apperr.AppError {
	return f.build(errdetails.Unknown, msg, reason, metadata, opts)
}

// InvalidArgument indicates client specified an invalid argument.
// Note that this differs from FailedPrecondition. It indicates arguments
// that are problematic regardless of the state of the system
// (e.g., a malformed file name).
//
// The `reason` should be a constant value that identifies the proximate cause
// of the error. `metadata` can attach further structured meta information to
// the error. The key must not exceed 64 characters in length.
//
// AIP-193 requires at least one FieldViolation for this code.
// The caller must supply a non-empty violations slice; passing an empty
// slice defeats the purpose of stricterr and violates AIP-193.
// (Further resources: https://google.aip.dev/193)
func (f *Factory) InvalidArgument(msg, reason string, metadata map[string]string, violations []apperr.FieldViolationOption, opts ...apperr.Option) *apperr.AppError {
	extra := make([]apperr.Option, 0, len(violations)+len(opts))
	for _, v := range violations {
		extra = append(extra, v)
	}
	extra = append(extra, opts...)
	return f.build(errdetails.InvalidArgument, msg, reason, metadata, extra)
}

// DeadlineExceeded means operation expired before completion.
// For operations that change the state of the system, this error may be
// returned even if the operation has completed successfully. For
// example, a successful response from a server could have been delayed
// long enough for the deadline to expire.
//
// The `reason` should be a constant value that identifies the proximate cause
// of the error. `metadata` can attach further structured meta information to
// the error. The key must not exceed 64 characters in length.
//
// (Further resources: https://google.aip.dev/193)
func (f *Factory) DeadlineExceeded(msg, reason string, metadata map[string]string, opts ...apperr.Option) *apperr.AppError {
	return f.build(errdetails.DeadlineExceeded, msg, reason, metadata, opts)
}

// NotFound means some requested entity (e.g., file or directory) was
// not found.
//
// The `reason` should be a constant value that identifies the proximate cause
// of the error. `metadata` can attach further structured meta information to
// the error. The key must not exceed 64 characters in length.
//
// AIP-193 requires ResourceInfo for this code.
// (Further resources: https://google.aip.dev/193)
func (f *Factory) NotFound(msg, reason string, metadata map[string]string, resource apperr.ResourceInfoOption, opts ...apperr.Option) *apperr.AppError {
	return f.build(errdetails.NotFound, msg, reason, metadata, append([]apperr.Option{resource}, opts...))
}

// AlreadyExists means an attempt to create an entity failed because one
// already exists.
//
// The `reason` should be a constant value that identifies the proximate cause
// of the error. `metadata` can attach further structured meta information to
// the error. The key must not exceed 64 characters in length.
//
// AIP-193 requires ResourceInfo for this code.
// (Further resources: https://google.aip.dev/193)
func (f *Factory) AlreadyExists(msg, reason string, metadata map[string]string, resource apperr.ResourceInfoOption, opts ...apperr.Option) *apperr.AppError {
	return f.build(errdetails.AlreadyExists, msg, reason, metadata, append([]apperr.Option{resource}, opts...))
}

// PermissionDenied indicates the caller does not have permission to
// execute the specified operation. It must not be used for rejections
// caused by exhausting some resource (use ResourceExhausted
// instead for those errors). It must not be
// used if the caller cannot be identified (use Unauthenticated
// instead for those errors).
//
// The `reason` should be a constant value that identifies the proximate cause
// of the error. `metadata` can attach further structured meta information to
// the error. The key must not exceed 64 characters in length.
//
// (Further resources: https://google.aip.dev/193)
func (f *Factory) PermissionDenied(msg, reason string, metadata map[string]string, opts ...apperr.Option) *apperr.AppError {
	return f.build(errdetails.PermissionDenied, msg, reason, metadata, opts)
}

// ResourceExhausted indicates some resource has been exhausted, perhaps
// a per-user quota, or perhaps the entire file system is out of space.
//
// The `reason` should be a constant value that identifies the proximate cause
// of the error. `metadata` can attach further structured meta information to
// the error. The key must not exceed 64 characters in length.
//
// AIP-193 requires at least one QuotaViolation for this code.
// The caller must supply a non-empty violations slice; passing an empty
// slice defeats the purpose of stricterr and violates AIP-193.
// (Further resources: https://google.aip.dev/193)
func (f *Factory) ResourceExhausted(msg, reason string, metadata map[string]string, violations []apperr.QuotaViolationOption, opts ...apperr.Option) *apperr.AppError {
	extra := make([]apperr.Option, 0, len(violations)+len(opts))
	for _, v := range violations {
		extra = append(extra, v)
	}
	extra = append(extra, opts...)
	return f.build(errdetails.ResourceExhausted, msg, reason, metadata, extra)
}

// FailedPrecondition indicates operation was rejected because the
// system is not in a state required for the operation's execution.
// For example, directory to be deleted may be non-empty, a rmdir
// operation is applied to a non-directory, etc.
//
// A litmus test that may help a service implementor in deciding
// between FailedPrecondition, Aborted, and Unavailable:
//
//	(a) Use Unavailable if the client can retry just the failing call.
//	(b) Use Aborted if the client should retry at a higher-level
//	    (e.g., restarting a read-modify-write sequence).
//	(c) Use FailedPrecondition if the client should not retry until
//	    the system state has been explicitly fixed. E.g., if an "rmdir"
//	    fails because the directory is non-empty, FailedPrecondition
//	    should be returned since the client should not retry unless
//	    they have first fixed up the directory by deleting files from it.
//	(d) Use FailedPrecondition if the client performs conditional
//	    REST Get/Update/Delete on a resource and the resource on the
//	    server does not match the condition. E.g., conflicting
//	    read-modify-write on the same resource.
//
// The `reason` should be a constant value that identifies the proximate cause
// of the error. `metadata` can attach further structured meta information to
// the error. The key must not exceed 64 characters in length.
//
// AIP-193 requires at least one PreconditionViolation for this code.
// The caller must supply a non-empty violations slice; passing an empty
// slice defeats the purpose of stricterr and violates AIP-193.
// (Further resources: https://google.aip.dev/193)
func (f *Factory) FailedPrecondition(msg, reason string, metadata map[string]string, violations []apperr.PreconditionViolationOption, opts ...apperr.Option) *apperr.AppError {
	extra := make([]apperr.Option, 0, len(violations)+len(opts))
	for _, v := range violations {
		extra = append(extra, v)
	}
	extra = append(extra, opts...)
	return f.build(errdetails.FailedPrecondition, msg, reason, metadata, extra)
}

// Aborted indicates the operation was aborted, typically due to a
// concurrency issue like sequencer check failures, transaction aborts,
// etc.
//
// See litmus test in [Factory.FailedPrecondition] for deciding between
// FailedPrecondition, Aborted, and Unavailable.
//
// The `reason` should be a constant value that identifies the proximate cause
// of the error. `metadata` can attach further structured meta information to
// the error. The key must not exceed 64 characters in length.
//
// (Further resources: https://google.aip.dev/193)
func (f *Factory) Aborted(msg, reason string, metadata map[string]string, opts ...apperr.Option) *apperr.AppError {
	return f.build(errdetails.Aborted, msg, reason, metadata, opts)
}

// OutOfRange means operation was attempted past the valid range.
// E.g., seeking or reading past end of file.
//
// Unlike InvalidArgument, this error indicates a problem that may
// be fixed if the system state changes. For example, a 32-bit file
// system will generate InvalidArgument if asked to read at an
// offset that is not in the range [0,2^32-1], but it will generate
// OutOfRange if asked to read from an offset past the current
// file size.
//
// There is a fair bit of overlap between FailedPrecondition and
// OutOfRange. We recommend using OutOfRange (the more specific
// error) when it applies so that callers who are iterating through
// a space can easily look for an OutOfRange error to detect when
// they are done.
//
// The `reason` should be a constant value that identifies the proximate cause
// of the error. `metadata` can attach further structured meta information to
// the error. The key must not exceed 64 characters in length.
//
// (Further resources: https://google.aip.dev/193)
func (f *Factory) OutOfRange(msg, reason string, metadata map[string]string, opts ...apperr.Option) *apperr.AppError {
	return f.build(errdetails.OutOfRange, msg, reason, metadata, opts)
}

// Unimplemented indicates operation is not implemented or not
// supported/enabled in this service.
//
// The `reason` should be a constant value that identifies the proximate cause
// of the error. `metadata` can attach further structured meta information to
// the error. The key must not exceed 64 characters in length.
//
// (Further resources: https://google.aip.dev/193)
func (f *Factory) Unimplemented(msg, reason string, metadata map[string]string, opts ...apperr.Option) *apperr.AppError {
	return f.build(errdetails.Unimplemented, msg, reason, metadata, opts)
}

// Internal errors. Means some invariants expected by underlying
// system has been broken. If you see one of these errors,
// something is very broken.
//
// The `reason` should be a constant value that identifies the proximate cause
// of the error. `metadata` can attach further structured meta information to
// the error. The key must not exceed 64 characters in length.
//
// (Further resources: https://google.aip.dev/193)
func (f *Factory) Internal(msg, reason string, metadata map[string]string, opts ...apperr.Option) *apperr.AppError {
	return f.build(errdetails.Internal, msg, reason, metadata, opts)
}

// Unavailable indicates the service is currently unavailable.
// This is a most likely a transient condition and may be corrected
// by retrying with a backoff. Note that it is not always safe to retry
// non-idempotent operations.
//
// See litmus test in [Factory.FailedPrecondition] for deciding between
// FailedPrecondition, Aborted, and Unavailable.
//
// The `reason` should be a constant value that identifies the proximate cause
// of the error. `metadata` can attach further structured meta information to
// the error. The key must not exceed 64 characters in length.
//
// AIP-193 recommends RetryInfo for this code when the client can retry.
// Pass nil if no retry information is available.
// (Further resources: https://google.aip.dev/193)
func (f *Factory) Unavailable(msg, reason string, metadata map[string]string, retry *apperr.RetryInfoOption, opts ...apperr.Option) *apperr.AppError {
	if retry != nil {
		opts = append([]apperr.Option{*retry}, opts...)
	}
	return f.build(errdetails.Unavailable, msg, reason, metadata, opts)
}

// DataLoss indicates unrecoverable data loss or corruption.
//
// The `reason` should be a constant value that identifies the proximate cause
// of the error. `metadata` can attach further structured meta information to
// the error. The key must not exceed 64 characters in length.
//
// (Further resources: https://google.aip.dev/193)
func (f *Factory) DataLoss(msg, reason string, metadata map[string]string, opts ...apperr.Option) *apperr.AppError {
	return f.build(errdetails.DataLoss, msg, reason, metadata, opts)
}

// Unauthenticated indicates the request does not have valid
// authentication credentials for the operation.
//
// The `reason` should be a constant value that identifies the proximate cause
// of the error. `metadata` can attach further structured meta information to
// the error. The key must not exceed 64 characters in length.
//
// (Further resources: https://google.aip.dev/193)
func (f *Factory) Unauthenticated(msg, reason string, metadata map[string]string, opts ...apperr.Option) *apperr.AppError {
	return f.build(errdetails.Unauthenticated, msg, reason, metadata, opts)
}
