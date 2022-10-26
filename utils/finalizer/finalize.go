package finalizer

import (
	"fmt"

	"github.com/harwoeck/apperr/utils/dto"

	"github.com/harwoeck/apperr"
	"github.com/harwoeck/liblog"
	"golang.org/x/text/language"
)

type renderConfig struct {
	Logger    liblog.Logger
	Provider  LocalizationProvider
	Languages []language.Tag
}

type RenderOption func(config *renderConfig)

func WithLogger(logger liblog.Logger) RenderOption {
	return func(config *renderConfig) {
		config.Logger = logger
	}
}

func WithLocalizationProvider(provider LocalizationProvider) RenderOption {
	return func(config *renderConfig) {
		config.Provider = provider
	}
}

func WithLanguages(languages []language.Tag) RenderOption {
	return func(config *renderConfig) {
		config.Languages = languages
	}
}

func Render(error *apperr.AppError, opts ...RenderOption) (*Error, error) {
	// init render config for this cycle
	config := &renderConfig{
		Logger: liblog.MustNewStd(liblog.DisableLogWrites()),
	}
	for _, opt := range opts {
		opt(config)
	}

	log := config.Logger

	// prepare final error struct
	final := &Error{
		Code:    error.Code(),
		Message: error.Message(),
	}

	// apply accumulated AppError options on our final error
	for _, opt := range error.Opts() {
		if err := opt(final); err != nil {
			return nil, log.ErrorReturn("failed to apply option to *finalize.Error", field("error", err))
		}
	}

	// if we have a LocalizationProvider, we translate specific fields
	if config.Provider != nil {
		err := final.localize(config)
		if err != nil {
			return nil, err
		}
	}

	return final, nil
}

func (e *Error) localize(config *renderConfig) error {
	if e.RawLocalized != nil {
		err := e.localizeTitleAndText(config)
		if err != nil {
			return err
		}
	}

	if len(e.RawFieldViolations) > 0 {
		e.FieldViolations = make([]*FieldViolation, 0)

		for _, fv := range e.RawFieldViolations {
			err := e.localizeFieldViolation(config, fv)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *Error) localizeTitleAndText(config *renderConfig) error {
	log := config.Logger

	// translate title
	titleID := fmt.Sprintf("CODE_%s", e.Code.String())
	title, tag, nf, err := config.Provider.Localize(titleID, config.Languages)
	if err != nil {
		return log.ErrorReturn("localization provider failed", field("error", err))
	} else if nf {
		log.Warn("localization provider didn't find message for title", field("id", titleID))
	}

	// translate text
	var text string
	if e.RawLocalized.TextID != nil {
		text, tag, nf, err = config.Provider.Localize(*e.RawLocalized.TextID, config.Languages)
	} else if e.RawLocalized.Any != nil {
		text, tag, nf, err = config.Provider.LocalizeAny(e.RawLocalized.Any, config.Languages)
	} else {
		log.Warn("neither TextID nor Any are set. Unable to provide translated text")
		return nil
	}
	if err != nil {
		return log.ErrorReturn("localization provider failed", field("error", err))
	} else if nf {
		log.Warn("localization provider didn't find message")
		return nil
	}

	// set translated strings
	e.Localized = &Localized{
		Locale: tag,
		Title:  title,
		Text:   text,
	}

	return nil
}

func (e *Error) localizeFieldViolation(config *renderConfig, fv *dto.FieldViolation) error {
	log := config.Logger

	// description is provided directly -> add unchanged
	if fv.Description != nil {
		e.FieldViolations = append(e.FieldViolations, &FieldViolation{
			Field:       fv.Field,
			Description: *fv.Description,
		})
		return nil
	}

	var (
		text string
		tag  language.Tag
		nf   bool
		err  error
	)

	// use provider to translate description of field violation
	if fv.DescriptionID != nil {
		text, tag, nf, err = config.Provider.Localize(*fv.DescriptionID, config.Languages)
	} else if fv.DescriptionAny != nil {
		text, tag, nf, err = config.Provider.LocalizeAny(fv.DescriptionAny, config.Languages)
	} else {
		log.Warn("neither DescriptionID nor DescriptionAny are set. Unable to provide translated description for field violation")
		return nil
	}

	// check for errors during translation
	if err != nil {
		return log.ErrorReturn("localization provider failed", field("error", err))
	} else if nf {
		log.Warn("localization provider didn't find message")
		return nil
	}

	e.FieldViolations = append(e.FieldViolations, &FieldViolation{
		Locale:      &tag,
		Field:       fv.Field,
		Description: text,
	})

	return nil
}
