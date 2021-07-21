package httperr

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/harwoeck/apperr/apperr"
	"github.com/harwoeck/apperr/apperr/code"
)

type jsonObjLocalized struct {
	UserMessage      string `json:"userMessage"`
	UserMessageShort string `json:"userMessageShort"`
	Locale           string `json:"locale"`
}

type jsonObj struct {
	Message   string            `json:"message"`
	Code      string            `json:"code"`
	Localized *jsonObjLocalized `json:"localized"`
}

func codeToHttpStatus(c code.Code) int {
	// mapping copied from https://cloud.google.com/apis/design/errors#handling_errors
	switch c {
	case code.Canceled:
		return 499 // https://httpstatuses.com/499
	case code.Unknown:
		return http.StatusInternalServerError
	case code.InvalidArgument:
		return http.StatusBadRequest
	case code.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case code.NotFound:
		return http.StatusNotFound
	case code.AlreadyExists:
		return http.StatusConflict
	case code.PermissionDenied:
		return http.StatusForbidden
	case code.ResourceExhausted:
		return http.StatusTooManyRequests
	case code.FailedPrecondition:
		return http.StatusBadRequest
	case code.Aborted:
		return http.StatusConflict
	case code.OutOfRange:
		return http.StatusBadRequest
	case code.Unimplemented:
		return http.StatusNotImplemented
	case code.Internal:
		return http.StatusInternalServerError
	case code.Unavailable:
		return http.StatusServiceUnavailable
	case code.DataLoss:
		return http.StatusInternalServerError
	case code.Unauthenticated:
		return http.StatusUnauthorized
	default:
		// THIS SHOULD NEVER HAPPEN
		return http.StatusInternalServerError
	}
}

func Convert(rendered *apperr.RenderedError) (httpStatusCode int, httpBody []byte, err error) {
	// copy rendered information to jsonObj
	obj := &jsonObj{
		Message: rendered.Message,
		Code:    rendered.Code.String(),
	}
	if rendered.Localized != nil {
		obj.Localized = &jsonObjLocalized{
			UserMessage:      rendered.Localized.UserMessage,
			UserMessageShort: rendered.Localized.UserMessageShort,
			Locale:           rendered.Localized.Locale.String(),
		}
	}

	// json encode jsonObj into bytes buffer
	buf, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return 0, nil, fmt.Errorf("apper/renderer/httperr.Convert: failed to encode json object with: %v", err)
	}

	// return final http error
	return codeToHttpStatus(rendered.Code), buf, nil
}
