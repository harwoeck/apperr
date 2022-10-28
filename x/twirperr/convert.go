package twirperr

import (
	"github.com/twitchtv/twirp"

	"github.com/harwoeck/apperr/utils/finalizer"
)

func Convert(rendered *finalizer.Error) twirp.Error {
	te := twirp.NewError(mapCode(rendered.Code), rendered.Message)

	if rendered.Localized != nil {
		te = te.WithMeta("localized_message", rendered.Localized.Text)
		te = te.WithMeta("localized_message_short", rendered.Localized.Title)
		te = te.WithMeta("localized_locale", rendered.Localized.Locale.String())
	}

	return te
}
