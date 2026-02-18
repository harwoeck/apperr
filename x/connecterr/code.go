package connecterr

import (
	"connectrpc.com/connect"
	"github.com/harwoeck/apperr/errdetails"
)

func mapCode(c errdetails.Code) connect.Code {
	switch c {
	case errdetails.Canceled:
		return connect.CodeCanceled
	case errdetails.Unknown:
		return connect.CodeUnknown
	case errdetails.InvalidArgument:
		return connect.CodeInvalidArgument
	case errdetails.DeadlineExceeded:
		return connect.CodeDeadlineExceeded
	case errdetails.NotFound:
		return connect.CodeNotFound
	case errdetails.AlreadyExists:
		return connect.CodeAlreadyExists
	case errdetails.PermissionDenied:
		return connect.CodePermissionDenied
	case errdetails.ResourceExhausted:
		return connect.CodeResourceExhausted
	case errdetails.FailedPrecondition:
		return connect.CodeFailedPrecondition
	case errdetails.Aborted:
		return connect.CodeAborted
	case errdetails.OutOfRange:
		return connect.CodeOutOfRange
	case errdetails.Unimplemented:
		return connect.CodeUnimplemented
	case errdetails.Internal:
		return connect.CodeInternal
	case errdetails.Unavailable:
		return connect.CodeUnavailable
	case errdetails.DataLoss:
		return connect.CodeDataLoss
	case errdetails.Unauthenticated:
		return connect.CodeUnauthenticated
	default:
		// THIS SHOULD NEVER HAPPEN
		return connect.CodeInternal
	}
}
