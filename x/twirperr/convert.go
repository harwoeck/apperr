package twirperr

import (
	"encoding/json"
	"fmt"

	"github.com/harwoeck/apperr/errdetails"
	"github.com/twitchtv/twirp"
)

// Convert transforms a resolved apperr error into a twirp.Error with the
// appropriate Twirp error code. Since Twirp does not support structured
// error details natively, the full [errdetails.ResolvedError] is
// JSON-encoded and attached as the "error_details" metadata key, giving
// clients access to all detail types (localized messages, field violations,
// etc.). Clients should parse that key as a JSON-encoded ResolvedError.
//
// The keepDebugInfo parameter controls whether [errdetails.DebugInfo] is
// included in the response. Pass false in production to ensure server-side
// debugging information (stack traces, internal details) is never leaked to
// clients. When false, resolved.DebugInfo is stripped before conversion.
func Convert(resolved *errdetails.ResolvedError, keepDebugInfo bool) (twirp.Error, error) {
	if !keepDebugInfo {
		resolved.DebugInfo = nil
	}

	te := twirp.NewError(mapCode(resolved.Code), resolved.Message)

	data, err := json.Marshal(resolved)
	if err != nil {
		return nil, fmt.Errorf("twirperr.Convert: failed to marshal resolved error to JSON: %w", err)
	}
	te = te.WithMeta("error_details", string(data))

	return te, nil
}
