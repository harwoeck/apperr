// Package i18n provides a [errdetails.LocalizationProvider] backed by the
// [github.com/nicksnyder/go-i18n/v2/i18n] library.
//
// The [LocalizeAny] method expects a [*i18n.LocalizeConfig] as its value
// argument (matching the [apperr.LocalizeAny] option). Passing any other
// type returns an error.
package i18n

import (
	"errors"
	"fmt"

	"github.com/harwoeck/apperr/errdetails"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type adapter struct {
	bundle *i18n.Bundle
}

// NewAdapter creates a LocalizationProvider backed by a
// nicksnyder/go-i18n Bundle.
func NewAdapter(bundle *i18n.Bundle) errdetails.LocalizationProvider {
	return &adapter{
		bundle: bundle,
	}
}

func (a *adapter) Localize(messageID string, languages []language.Tag) (msg string, tag language.Tag, notFound bool, err error) {
	msg, tag, err = i18n.NewLocalizer(a.bundle, tagsToStrings(languages)...).LocalizeWithTag(&i18n.LocalizeConfig{
		MessageID: messageID,
	})
	if isNotFound(err) {
		return "", language.Und, true, nil
	}
	return
}

func (a *adapter) LocalizeAny(v any, languages []language.Tag) (msg string, tag language.Tag, notFound bool, err error) {
	locCfg, ok := v.(*i18n.LocalizeConfig)
	if !ok {
		return "", language.Und, false, fmt.Errorf("i18n.LocalizeAny: expected *i18n.LocalizeConfig, got %T; see apperr.LocalizeAny documentation", v)
	}

	msg, tag, err = i18n.NewLocalizer(a.bundle, tagsToStrings(languages)...).LocalizeWithTag(locCfg)
	if isNotFound(err) {
		return "", language.Und, true, nil
	}
	return
}

func tagsToStrings(tags []language.Tag) []string {
	s := make([]string, len(tags))
	for i, t := range tags {
		s[i] = t.String()
	}
	return s
}

func isNotFound(err error) bool {
	var nf *i18n.MessageNotFoundErr
	return errors.As(err, &nf)
}
