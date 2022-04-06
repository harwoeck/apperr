package apperr

import (
	"fmt"

	"github.com/harwoeck/apperr/code"
	"github.com/harwoeck/liblog"
	"golang.org/x/text/language"
)

type RenderConfig struct {
	Logger                liblog.Logger
	LocalizationProvider  LocalizationProvider
	LocalizationLanguages []string
}

type RenderOption func(renderer *RenderConfig) error

func RenderLocalized(provider LocalizationProvider, languages ...string) RenderOption {
	return func(renderer *RenderConfig) error {
		if renderer.LocalizationProvider != nil {
			return fmt.Errorf("apperr.RenderLocalized: cannot provide multiple LocalizationProvider")
		}

		renderer.LocalizationProvider = provider
		renderer.LocalizationLanguages = languages
		return nil
	}
}

func EnableLogging(log liblog.Logger) RenderOption {
	return func(renderer *RenderConfig) error {
		renderer.Logger = log.Named("apperr")
		return nil
	}
}

type RenderedError struct {
	Code      code.Code
	Message   string
	Localized *RenderedErrorLocalized
}

type RenderedErrorLocalized struct {
	UserMessage      string
	UserMessageShort string
	Locale           language.Tag
}

func Render(appError *AppError, opts ...RenderOption) (*RenderedError, error) {
	config := &RenderConfig{
		Logger: liblog.MustNewStd(liblog.DisableLogWrites()).Named("apperr"),
	}
	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, config.Logger.ErrorReturn("failed to apply option to *RenderConfig", field("error", err))
		}
	}

	for _, opt := range appError.opts {
		if err := opt(appError); err != nil {
			return nil, config.Logger.ErrorReturn("failed to apply option to *AppError", field("error", err))
		}
	}

	rendered := &RenderedError{
		Code:    appError.code,
		Message: appError.message,
	}

	// when localization provider is available localize error
	if config.LocalizationProvider != nil {
		config.Logger.Debug("localization provider is available. going to localize *AppError")
		loc := config.LocalizationProvider

		// use user-provided support languages, as long as the error doesn't
		// specify an explicit language
		languages := config.LocalizationLanguages
		if appError.optLocalizedLanguage != nil {
			languages = []string{*appError.optLocalizedLanguage}
		}
		config.Logger.Debug("using languages", field("languages", languages))

		var msg, short string
		var tag language.Tag
		var notFound bool
		var err error

		// localize code to short message
		shortID := "CODE_" + appError.code.String()
		short, _, notFound, err = loc.Localize(shortID, languages)
		if err != nil {
			return nil, config.Logger.ErrorReturn("localization provider failed for short message",
				field("error", err),
				field("short_id", shortID))
		}
		if notFound {
			config.Logger.Debug("localization provider didn't find short message", field("short_id", shortID))
		}

		var notLocalized bool

		// localize msg id
		if appError.optLocalizedConfig != nil {
			config.Logger.Debug("using localized config")
			msg, tag, notFound, err = loc.LocalizeFromConfig(appError.optLocalizedConfig, languages)
		} else if appError.optLocalizedMsgID != nil {
			config.Logger.Debug("using message id", field("message_id", *appError.optLocalizedMsgID))
			msg, tag, notFound, err = loc.Localize(*appError.optLocalizedMsgID, languages)
		} else {
			notLocalized = true
		}
		if err != nil {
			return nil, config.Logger.ErrorReturn("localization provider failed for message", field("error", err))
		}
		if notFound {
			config.Logger.Debug("localization provider didn't find message")
			notLocalized = true
		}

		if !notLocalized {
			config.Logger.Debug("assigning localized information to rendered error",
				field("message", msg),
				field("language", tag.String()))

			// assign localized object to rendered error
			rendered.Localized = &RenderedErrorLocalized{
				UserMessage:      msg,
				UserMessageShort: short,
				Locale:           tag,
			}
		}
	}

	return rendered, nil
}
