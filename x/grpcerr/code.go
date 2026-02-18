package grpcerr

import (
	"github.com/harwoeck/apperr/errdetails"
	"google.golang.org/grpc/codes"
)

func mapCode(c errdetails.Code) codes.Code {
	switch c {
	case errdetails.Canceled:
		return codes.Canceled
	case errdetails.Unknown:
		return codes.Unknown
	case errdetails.InvalidArgument:
		return codes.InvalidArgument
	case errdetails.DeadlineExceeded:
		return codes.DeadlineExceeded
	case errdetails.NotFound:
		return codes.NotFound
	case errdetails.AlreadyExists:
		return codes.AlreadyExists
	case errdetails.PermissionDenied:
		return codes.PermissionDenied
	case errdetails.ResourceExhausted:
		return codes.ResourceExhausted
	case errdetails.FailedPrecondition:
		return codes.FailedPrecondition
	case errdetails.Aborted:
		return codes.Aborted
	case errdetails.OutOfRange:
		return codes.OutOfRange
	case errdetails.Unimplemented:
		return codes.Unimplemented
	case errdetails.Internal:
		return codes.Internal
	case errdetails.Unavailable:
		return codes.Unavailable
	case errdetails.DataLoss:
		return codes.DataLoss
	case errdetails.Unauthenticated:
		return codes.Unauthenticated
	default:
		// THIS SHOULD NEVER HAPPEN
		return codes.Internal
	}
}
