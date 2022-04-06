package apperr

import "golang.org/x/text/language"

// LocalizationProvider is an interface for localization implementors
type LocalizationProvider interface {
	Localize(msgID string, languages []string) (msg string, tag language.Tag, notFound bool, err error)
	LocalizeFromConfig(cfg interface{}, languages []string) (msg string, tag language.Tag, notFound bool, err error)
}
