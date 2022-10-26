package nicksnyder_i18n

import (
	"fmt"
	"strings"

	"github.com/harwoeck/apperr/apperr"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type adapter struct {
	bundle *i18n.Bundle
}

func (a *adapter) Localize(msgID string, languages []string) (msg string, tag language.Tag, notFound bool, err error) {
	msg, tag, err = i18n.NewLocalizer(a.bundle, languages...).LocalizeWithTag(&i18n.LocalizeConfig{
		MessageID: msgID,
	})
	if err != nil && strings.Contains(err.Error(), "not found") {
		return "", language.Und, true, nil
	}
	return
}

func (a *adapter) LocalizeFromConfig(cfg interface{}, languages []string) (msg string, tag language.Tag, notFound bool, err error) {
	locCfg, ok := cfg.(*i18n.LocalizeConfig)
	if !ok {
		return "", language.Und, false, fmt.Errorf("apperr/adapter/nicksnyder-i18n.LocalizeFromConfig: unable to type assert interface cfg to *nicksnyder-i18n.LocalizeConfig")
	}

	msg, tag, err = i18n.NewLocalizer(a.bundle, languages...).LocalizeWithTag(locCfg)
	if err != nil && strings.Contains(err.Error(), "not found") {
		return "", language.Und, true, nil
	}
	return
}

func NewI18nAdapter(bundle *i18n.Bundle) apperr.LocalizationProvider {
	return &adapter{
		bundle: bundle,
	}
}
