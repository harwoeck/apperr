package finalizer

import "golang.org/x/text/language"

// LocalizationProvider is an interface for localization implementors
type LocalizationProvider interface {
	Localize(messageID string, languages []language.Tag) (msg string, tag language.Tag, notFound bool, err error)
	LocalizeAny(any interface{}, languages []language.Tag) (msg string, tag language.Tag, notFound bool, err error)
}
