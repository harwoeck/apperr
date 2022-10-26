package terminalerr

import (
	"fmt"

	"github.com/harwoeck/apperr/utils/finalizer"
)

func Convert(rendered *finalizer.Error) string {
	s := fmt.Sprintf("%s: %s", rendered.Code.String(), rendered.Message)

	if rendered.Localized != nil {
		s += fmt.Sprintf(" ([%s] %s: %s)", rendered.Localized.Locale.String(), rendered.Localized.Title, rendered.Localized.Text)
	}

	return s
}
