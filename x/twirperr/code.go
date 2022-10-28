package twirperr

import (
	"github.com/twitchtv/twirp"

	"github.com/harwoeck/apperr/utils/code"
)

func mapCode(c code.Code) twirp.ErrorCode {
	switch c {
	case code.Canceled:
		return twirp.Canceled
	case code.Unknown:
		return twirp.Unknown
	case code.InvalidArgument:
		return twirp.InvalidArgument
	case code.DeadlineExceeded:
		return twirp.DeadlineExceeded
	case code.NotFound:
		return twirp.NotFound
	case code.AlreadyExists:
		return twirp.AlreadyExists
	case code.PermissionDenied:
		return twirp.PermissionDenied
	case code.ResourceExhausted:
		return twirp.ResourceExhausted
	case code.FailedPrecondition:
		return twirp.FailedPrecondition
	case code.Aborted:
		return twirp.Aborted
	case code.OutOfRange:
		return twirp.OutOfRange
	case code.Unimplemented:
		return twirp.Unimplemented
	case code.Internal:
		return twirp.Internal
	case code.Unavailable:
		return twirp.Unavailable
	case code.DataLoss:
		return twirp.DataLoss
	case code.Unauthenticated:
		return twirp.Unauthenticated
	default:
		// THIS SHOULD NEVER HAPPEN
		return twirp.Internal
	}
}
