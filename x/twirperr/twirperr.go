package twirperr

import (
	"github.com/twitchtv/twirp"

	"github.com/harwoeck/apperr/utils/code"
	"github.com/harwoeck/apperr/utils/finalizer"
)

func codeToTwirp(c code.Code) twirp.ErrorCode {
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

func Convert(rendered *finalizer.Error) twirp.Error {
	te := twirp.NewError(codeToTwirp(rendered.Code), rendered.Message)

	if rendered.Localized != nil {
		te = te.WithMeta("localized_message", rendered.Localized.Text)
		te = te.WithMeta("localized_message_short", rendered.Localized.Title)
		te = te.WithMeta("localized_locale", rendered.Localized.Locale.String())
	}

	return te
}
