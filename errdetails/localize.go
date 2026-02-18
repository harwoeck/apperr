package errdetails

import "golang.org/x/text/language"

// LocalizationProvider is implemented by localization adapters (e.g. x/i18n)
// to translate message IDs into localized strings.
type LocalizationProvider interface {
	Localize(messageID string, languages []language.Tag) (msg string, tag language.Tag, notFound bool, err error)
	LocalizeAny(v any, languages []language.Tag) (msg string, tag language.Tag, notFound bool, err error)
}
