package apperr

import (
	"fmt"

	"github.com/harwoeck/apperr/apperr/code"
)

// AppError represents a general application error
type AppError struct {
	code    code.Code
	message string
	opts    []Option

	optLocalizedMsgID    *string
	optLocalizedLanguage *string
	optLocalizedConfig   interface{}
}

func (a *AppError) Error() string {
	return fmt.Sprintf("%s: %s", a.code.String(), a.message)
}

func newAppError(status code.Code, msg string, opts []Option) *AppError {
	return &AppError{
		code:    status,
		message: msg,
		opts:    opts,
	}
}

// Canceled indicates the operation was canceled (typically by the caller).
func Canceled(msg string, opts ...Option) *AppError {
	return newAppError(code.Canceled, msg, opts)
}

// Unknown error. An example of where this error may be returned is
// if a Status value received from another address space belongs to
// an error-space that is not known in this address space. Also
// errors raised by APIs that do not return enough error information
// may be converted to this error.
func Unknown(msg string, opts ...Option) *AppError {
	return newAppError(code.Unknown, msg, opts)
}

// InvalidArgument indicates client specified an invalid argument.
// Note that this differs from FailedPrecondition. It indicates arguments
// that are problematic regardless of the state of the system
// (e.g., a malformed file name).
func InvalidArgument(msg string, opts ...Option) *AppError {
	return newAppError(code.InvalidArgument, msg, opts)
}

// DeadlineExceeded means operation expired before completion.
// For operations that change the state of the system, this error may be
// returned even if the operation has completed successfully. For
// example, a successful response from a server could have been delayed
// long enough for the deadline to expire.
func DeadlineExceeded(msg string, opts ...Option) *AppError {
	return newAppError(code.DeadlineExceeded, msg, opts)
}

// NotFound means some requested entity (e.g., file or directory) was
// not found.
func NotFound(msg string, opts ...Option) *AppError {
	return newAppError(code.NotFound, msg, opts)
}

// AlreadyExists means an attempt to create an entity failed because one
// already exists.
func AlreadyExists(msg string, opts ...Option) *AppError {
	return newAppError(code.AlreadyExists, msg, opts)
}

// PermissionDenied indicates the caller does not have permission to
// execute the specified operation. It must not be used for rejections
// caused by exhausting some resource (use ResourceExhausted
// instead for those errors). It must not be
// used if the caller cannot be identified (use Unauthenticated
// instead for those errors).
func PermissionDenied(msg string, opts ...Option) *AppError {
	return newAppError(code.PermissionDenied, msg, opts)
}

// ResourceExhausted indicates some resource has been exhausted, perhaps
// a per-user quota, or perhaps the entire file system is out of space.
func ResourceExhausted(msg string, opts ...Option) *AppError {
	return newAppError(code.ResourceExhausted, msg, opts)
}

// FailedPrecondition indicates operation was rejected because the
// system is not in a state required for the operation's execution.
// For example, directory to be deleted may be non-empty, an rmdir
// operation is applied to a non-directory, etc.
//
// A litmus test that may help a service implementor in deciding
// between FailedPrecondition, Aborted, and Unavailable:
//  (a) Use Unavailable if the client can retry just the failing call.
//  (b) Use Aborted if the client should retry at a higher-level
//      (e.g., restarting a read-modify-write sequence).
//  (c) Use FailedPrecondition if the client should not retry until
//      the system state has been explicitly fixed. E.g., if an "rmdir"
//      fails because the directory is non-empty, FailedPrecondition
//      should be returned since the client should not retry unless
//      they have first fixed up the directory by deleting files from it.
//  (d) Use FailedPrecondition if the client performs conditional
//      REST Get/Update/Delete on a resource and the resource on the
//      server does not match the condition. E.g., conflicting
//      read-modify-write on the same resource.
func FailedPrecondition(msg string, opts ...Option) *AppError {
	return newAppError(code.FailedPrecondition, msg, opts)
}

// Aborted indicates the operation was aborted, typically due to a
// concurrency issue like sequencer check failures, transaction aborts,
// etc.
//
// See litmus test above for deciding between FailedPrecondition,
// Aborted, and Unavailable.
func Aborted(msg string, opts ...Option) *AppError {
	return newAppError(code.Aborted, msg, opts)
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
	return newAppError(code.OutOfRange, msg, opts)
}

// Unimplemented indicates operation is not implemented or not
// supported/enabled in this service.
func Unimplemented(msg string, opts ...Option) *AppError {
	return newAppError(code.Unimplemented, msg, opts)
}

// Internal errors. Means some invariants expected by underlying
// system has been broken. If you see one of these errors,
// something is very broken.
func Internal(msg string, opts ...Option) *AppError {
	return newAppError(code.Internal, msg, opts)
}

// Unavailable indicates the service is currently unavailable.
// This is a most likely a transient condition and may be corrected
// by retrying with a backoff. Note that it is not always safe to retry
// non-idempotent operations.
//
// See litmus test above for deciding between FailedPrecondition,
// Aborted, and Unavailable.
func Unavailable(msg string, opts ...Option) *AppError {
	return newAppError(code.Unavailable, msg, opts)
}

// DataLoss indicates unrecoverable data loss or corruption.
func DataLoss(msg string, opts ...Option) *AppError {
	return newAppError(code.DataLoss, msg, opts)
}

// Unauthenticated indicates the request does not have valid
// authentication credentials for the operation.
func Unauthenticated(msg string, opts ...Option) *AppError {
	return newAppError(code.Unauthenticated, msg, opts)
}
