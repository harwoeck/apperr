package connecterr

import (
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/harwoeck/apperr/errdetails"
	"github.com/harwoeck/apperr/x/protoerr"
)

// Convert transforms a resolved apperr error into a *connect.Error with the
// appropriate Connect RPC status code and rich error details attached via
// the standard googleapis rpc errdetails protos.
//
// The keepDebugInfo parameter controls whether [errdetails.DebugInfo] is
// included in the response. Pass false in production to ensure server-side
// debugging information (stack traces, internal details) is never leaked to
// clients. When false, resolved.DebugInfo is stripped before conversion.
func Convert(resolved *errdetails.ResolvedError, keepDebugInfo bool) (*connect.Error, error) {
	if !keepDebugInfo {
		resolved.DebugInfo = nil
	}

	ce := connect.NewError(mapCode(resolved.Code), errors.New(resolved.Message))

	for _, msg := range protoerr.Details(resolved) {
		detail, err := connect.NewErrorDetail(msg)
		if err != nil {
			return nil, fmt.Errorf("connecterr.Convert: failed to create error detail: %w", err)
		}
		ce.AddDetail(detail)
	}

	return ce, nil
}
