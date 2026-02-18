package grpcerr

import (
	"fmt"

	"github.com/harwoeck/apperr/errdetails"
	"github.com/harwoeck/apperr/x/protoerr"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/protoadapt"
)

// Convert transforms a resolved apperr error into a *status.Status with
// the appropriate gRPC status code and rich error details attached via
// the standard googleapis rpc errdetails protos.
//
// The keepDebugInfo parameter controls whether [errdetails.DebugInfo] is
// included in the response. Pass false in production to ensure server-side
// debugging information (stack traces, internal details) is never leaked to
// clients. When false, resolved.DebugInfo is stripped before conversion.
func Convert(resolved *errdetails.ResolvedError, keepDebugInfo bool) (*status.Status, error) {
	if !keepDebugInfo {
		resolved.DebugInfo = nil
	}

	st := status.New(mapCode(resolved.Code), resolved.Message)

	msgs := protoerr.Details(resolved)
	if len(msgs) > 0 {
		details := make([]protoadapt.MessageV1, len(msgs))
		for i, msg := range msgs {
			details[i] = protoadapt.MessageV1Of(msg)
		}

		var err error
		st, err = st.WithDetails(details...)
		if err != nil {
			return nil, fmt.Errorf("grpcerr.Convert: failed to attach error details to gRPC status: %w", err)
		}
	}

	return st, nil
}
