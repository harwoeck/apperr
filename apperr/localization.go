package apperr

import "golang.org/x/text/language"

// LocalizationProvider is an interface for localization implementors (libraries).
// One provided implementation for that uses github.com/nicksnyder/go-i18n as it's
// backend can be found at adapter/i18n
type LocalizationProvider interface {
	Localize(msgID string, languages []string) (msg string, tag language.Tag, notFound bool, err error)
	LocalizeFromConfig(cfg interface{}, languages []string) (msg string, tag language.Tag, notFound bool, err error)
}
