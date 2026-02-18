package httperr

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/harwoeck/apperr/errdetails"
)

func mapHTTPStatus(c errdetails.Code) int {
	// Mapping according to https://cloud.google.com/apis/design/errors#handling_errors
	switch c {
	case errdetails.Canceled:
		return 499
	case errdetails.Unknown:
		return http.StatusInternalServerError
	case errdetails.InvalidArgument:
		return http.StatusBadRequest
	case errdetails.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case errdetails.NotFound:
		return http.StatusNotFound
	case errdetails.AlreadyExists:
		return http.StatusConflict
	case errdetails.PermissionDenied:
		return http.StatusForbidden
	case errdetails.ResourceExhausted:
		return http.StatusTooManyRequests
	case errdetails.FailedPrecondition:
		return http.StatusBadRequest
	case errdetails.Aborted:
		return http.StatusConflict
	case errdetails.OutOfRange:
		return http.StatusBadRequest
	case errdetails.Unimplemented:
		return http.StatusNotImplemented
	case errdetails.Internal:
		return http.StatusInternalServerError
	case errdetails.Unavailable:
		return http.StatusServiceUnavailable
	case errdetails.DataLoss:
		return http.StatusInternalServerError
	case errdetails.Unauthenticated:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

// Convert transforms a resolved apperr error into an HTTP status code and
// a JSON-encoded body. The body is the JSON representation of the full
// ResolvedError, giving clients access to all detail types (localized
// messages, field violations, error info, etc.).
//
// The keepDebugInfo parameter controls whether [errdetails.DebugInfo] is
// included in the response. Pass false in production to ensure server-side
// debugging information (stack traces, internal details) is never leaked to
// clients. When false, resolved.DebugInfo is stripped before conversion.
func Convert(resolved *errdetails.ResolvedError, keepDebugInfo bool) (httpStatusCode int, httpBody []byte, err error) {
	if !keepDebugInfo {
		resolved.DebugInfo = nil
	}

	body, err := json.Marshal(resolved)
	if err != nil {
		return 0, nil, fmt.Errorf("httperr.Convert: failed to marshal resolved error: %w", err)
	}

	return mapHTTPStatus(resolved.Code), body, nil
}
