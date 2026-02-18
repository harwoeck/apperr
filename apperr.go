package apperr

import (
	"errors"
	"fmt"

	"github.com/harwoeck/apperr/errdetails"
)

// AppError represents a general application error tied to a single request.
//
// An AppError is inherently request-scoped. It is designed to be created,
// enriched with options, resolved, and converted within the lifetime of a
// single request. It is NOT safe for concurrent use by multiple goroutines.
// Do not store an *AppError in a package-level variable and mutate it from
// multiple requests — create a new instance per request instead.
type AppError struct {
	code    errdetails.Code
	message string
	cause   error
	opts    []Option
}

// Code returns the code
func (a *AppError) Code() errdetails.Code {
	return a.code
}

// Message returns the message
func (a *AppError) Message() string {
	return a.message
}

// Opts returns the accumulated options
func (a *AppError) Opts() []Option {
	return a.opts
}

// Error implements Go's error interface
func (a *AppError) Error() string {
	if a.cause != nil {
		return fmt.Sprintf("%s: %s: %s", a.code.String(), a.message, a.cause.Error())
	}
	return fmt.Sprintf("%s: %s", a.code.String(), a.message)
}

// Unwrap returns the underlying wrapped error, if any. This allows
// *AppError to participate in Go's errors.Is and errors.As chains.
func (a *AppError) Unwrap() error {
	return a.cause
}

// AppendOptions adds further Option to the AppError instance.
//
// AppError is inherently request-scoped and is NOT safe for concurrent use.
// Do not call AppendOptions on a shared *AppError from multiple goroutines.
// This method mutates the receiver's option slice in place.
func (a *AppError) AppendOptions(opts ...Option) {
	a.opts = append(a.opts, opts...)
}

// IsCode reports whether err's chain contains an *AppError whose code
// matches the given code. It uses errors.As to unwrap the chain.
func IsCode(err error, code errdetails.Code) bool {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae.Code() == code
	}
	return false
}

// AsAppError extracts an *AppError from err's chain using errors.As.
// It returns the *AppError and true if found, or nil and false otherwise.
func AsAppError(err error) (*AppError, bool) {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}

func newAppError(status errdetails.Code, msg string, opts []Option) *AppError {
	if !status.Valid() {
		panic(fmt.Sprintf("apperr: invalid error code %d; must be a defined errdetails constant (0-16)", uint32(status)))
	}

	// Extract WrapOption from opts (if any) to set the cause field.
	var cause error
	filtered := make([]Option, 0, len(opts))
	for _, opt := range opts {
		if w, ok := opt.(WrapOption); ok {
			cause = w.Err
		} else {
			filtered = append(filtered, opt)
		}
	}

	return &AppError{
		code:    status,
		message: msg,
		cause:   cause,
		opts:    filtered,
	}
}

// Canceled indicates the operation was canceled (typically by the caller).
func Canceled(msg string, opts ...Option) *AppError {
	return newAppError(errdetails.Canceled, msg, opts)
}

// Unknown error. An example of where this error may be returned is
// if a Status value received from another address space belongs to
// an error-space that is not known in this address space. Also,
// errors raised by APIs that do not return enough error information
// may be converted to this error.
func Unknown(msg string, opts ...Option) *AppError {
	return newAppError(errdetails.Unknown, msg, opts)
}

// InvalidArgument indicates client specified an invalid argument.
// Note that this differs from FailedPrecondition. It indicates arguments
// that are problematic regardless of the state of the system
// (e.g., a malformed file name).
func InvalidArgument(msg string, opts ...Option) *AppError {
	return newAppError(errdetails.InvalidArgument, msg, opts)
}

// DeadlineExceeded means operation expired before completion.
// For operations that change the state of the system, this error may be
// returned even if the operation has completed successfully. For
// example, a successful response from a server could have been delayed
// long enough for the deadline to expire.
func DeadlineExceeded(msg string, opts ...Option) *AppError {
	return newAppError(errdetails.DeadlineExceeded, msg, opts)
}

// NotFound means some requested entity (e.g., file or directory) was
// not found.
func NotFound(msg string, opts ...Option) *AppError {
	return newAppError(errdetails.NotFound, msg, opts)
}

// AlreadyExists means an attempt to create an entity failed because one
// already exists.
func AlreadyExists(msg string, opts ...Option) *AppError {
	return newAppError(errdetails.AlreadyExists, msg, opts)
}

// PermissionDenied indicates the caller does not have permission to
// execute the specified operation. It must not be used for rejections
// caused by exhausting some resource (use ResourceExhausted
// instead for those errors). It must not be
// used if the caller cannot be identified (use Unauthenticated
// instead for those errors).
func PermissionDenied(msg string, opts ...Option) *AppError {
	return newAppError(errdetails.PermissionDenied, msg, opts)
}

// ResourceExhausted indicates some resource has been exhausted, perhaps
// a per-user quota, or perhaps the entire file system is out of space.
func ResourceExhausted(msg string, opts ...Option) *AppError {
	return newAppError(errdetails.ResourceExhausted, msg, opts)
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
func FailedPrecondition(msg string, opts ...Option) *AppError {
	return newAppError(errdetails.FailedPrecondition, msg, opts)
}

// Aborted indicates the operation was aborted, typically due to a
// concurrency issue like sequencer check failures, transaction aborts,
// etc.
//
// See litmus test above for deciding between FailedPrecondition,
// Aborted, and Unavailable.
func Aborted(msg string, opts ...Option) *AppError {
	return newAppError(errdetails.Aborted, msg, opts)
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
func OutOfRange(msg string, opts ...Option) *AppError {
	return newAppError(errdetails.OutOfRange, msg, opts)
}

// Unimplemented indicates operation is not implemented or not
// supported/enabled in this service.
func Unimplemented(msg string, opts ...Option) *AppError {
	return newAppError(errdetails.Unimplemented, msg, opts)
}

// Internal errors. Means some invariants expected by underlying
// system has been broken. If you see one of these errors,
// something is very broken.
func Internal(msg string, opts ...Option) *AppError {
	return newAppError(errdetails.Internal, msg, opts)
}

// Unavailable indicates the service is currently unavailable.
// This is a most likely a transient condition and may be corrected
// by retrying with a backoff. Note that it is not always safe to retry
// non-idempotent operations.
//
// See litmus test above for deciding between FailedPrecondition,
// Aborted, and Unavailable.
func Unavailable(msg string, opts ...Option) *AppError {
	return newAppError(errdetails.Unavailable, msg, opts)
}

// DataLoss indicates unrecoverable data loss or corruption.
func DataLoss(msg string, opts ...Option) *AppError {
	return newAppError(errdetails.DataLoss, msg, opts)
}

// Unauthenticated indicates the request does not have valid
// authentication credentials for the operation.
func Unauthenticated(msg string, opts ...Option) *AppError {
	return newAppError(errdetails.Unauthenticated, msg, opts)
}
