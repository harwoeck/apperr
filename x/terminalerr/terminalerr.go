package terminalerr

import (
	"fmt"

	"github.com/harwoeck/apperr/apperr"
)

func Convert(rendered *apperr.RenderedError) string {
	s := fmt.Sprintf("%s: %s", rendered.Code.String(), rendered.Message)

	if rendered.Localized != nil {
		s += fmt.Sprintf(" ([%s] %s: %s)", rendered.Localized.Locale.String(), rendered.Localized.UserMessageShort, rendered.Localized.UserMessage)
	}

	return s
}
