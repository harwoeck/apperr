package grpcerr

import (
	"fmt"

	"github.com/harwoeck/apperr/apperr"
	"github.com/harwoeck/apperr/apperr/code"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func codeToGrpc(c code.Code) codes.Code {
	switch c {
	case code.Canceled:
		return codes.Canceled
	case code.Unknown:
		return codes.Unknown
	case code.InvalidArgument:
		return codes.InvalidArgument
	case code.DeadlineExceeded:
		return codes.DeadlineExceeded
	case code.NotFound:
		return codes.NotFound
	case code.AlreadyExists:
		return codes.AlreadyExists
	case code.PermissionDenied:
		return codes.PermissionDenied
	case code.ResourceExhausted:
		return codes.ResourceExhausted
	case code.FailedPrecondition:
		return codes.FailedPrecondition
	case code.Aborted:
		return codes.Aborted
	case code.OutOfRange:
		return codes.OutOfRange
	case code.Unimplemented:
		return codes.Unimplemented
	case code.Internal:
		return codes.Internal
	case code.Unavailable:
		return codes.Unavailable
	case code.DataLoss:
		return codes.DataLoss
	case code.Unauthenticated:
		return codes.Unauthenticated
	default:
		// THIS SHOULD NEVER HAPPEN
		return codes.Internal
	}
}

func Convert(rendered *apperr.RenderedError) (*status.Status, error) {
	st := status.New(codeToGrpc(rendered.Code), rendered.Message)

	if rendered.Localized != nil {
		var err error
		st, err = st.WithDetails(&errdetails.LocalizedMessage{
			Locale:  rendered.Localized.Locale.String(),
			Message: fmt.Sprintf("%s: %s", rendered.Localized.UserMessageShort, rendered.Localized.UserMessage),
		})
		if err != nil {
			return nil, fmt.Errorf("apperr/renderer/grpc.Convert: failed to append localized message to grpc status with: %v", err)
		}
	}

	return st, nil
}
