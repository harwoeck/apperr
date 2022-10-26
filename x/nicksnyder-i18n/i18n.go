package nicksnyder_i18n

import (
	"fmt"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"

	"github.com/harwoeck/apperr/utils/finalizer"
)

type adapter struct {
	bundle *i18n.Bundle
}

func (a *adapter) Localize(messageID string, languages []language.Tag) (msg string, tag language.Tag, notFound bool, err error) {
	msg, tag, err = i18n.NewLocalizer(a.bundle, languages...).LocalizeWithTag(&i18n.LocalizeConfig{
		MessageID: messageID,
	})
	if err != nil && strings.Contains(err.Error(), "not found") {
		return "", language.Und, true, nil
	}
	return
}

func (a *adapter) LocalizeAny(any interface{}, languages []language.Tag) (msg string, tag language.Tag, notFound bool, err error) {
	locCfg, ok := any.(*i18n.LocalizeConfig)
	if !ok {
		return "", language.Und, false, fmt.Errorf("apperr/adapter/nicksnyder-i18n.LocalizeFromConfig: unable to type assert interface cfg to *nicksnyder-i18n.LocalizeConfig")
	}

	msg, tag, err = i18n.NewLocalizer(a.bundle, languages...).LocalizeWithTag(locCfg)
	if err != nil && strings.Contains(err.Error(), "not found") {
		return "", language.Und, true, nil
	}
	return
}

func NewI18nAdapter(bundle *i18n.Bundle) finalizer.LocalizationProvider {
	return &adapter{
		bundle: bundle,
	}
}
