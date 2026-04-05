package resolve

import (
	"fmt"
	"log/slog"

	"golang.org/x/text/language"

	"github.com/harwoeck/apperr"
	"github.com/harwoeck/apperr/errdetails"
)

type config struct {
	Logger    *slog.Logger
	Provider  errdetails.LocalizationProvider
	Languages []language.Tag
}

// Option configures how an AppError is resolved.
type Option func(config *config)

// WithLogger sets the logger used during resolution.
func WithLogger(logger *slog.Logger) Option {
	return func(config *config) {
		config.Logger = logger
	}
}

// WithLocalizationProvider sets the provider used to translate message IDs.
func WithLocalizationProvider(provider errdetails.LocalizationProvider) Option {
	return func(config *config) {
		config.Provider = provider
	}
}

// WithLanguages sets the preferred language tags for localization.
func WithLanguages(languages []language.Tag) Option {
	return func(config *config) {
		config.Languages = languages
	}
}

// Error resolves an *apperr.AppError into a *errdetails.ResolvedError by
// applying all accumulated options and performing localization when a
// LocalizationProvider is available.
//
// Internally, options are first applied to an [errdetails.UnresolvedError]
// (which may carry raw, pre-localization fields). After localization the
// result is projected into a [errdetails.ResolvedError] that exposes only
// the final consumer-facing fields.
//
// It returns three values:
//
//   - resolved: the resolved error, which is usable even when localization
//     fails (it simply won't contain translated messages).
//   - localizationFailed: true when the LocalizationProvider returned an
//     error. In this case resolved is non-nil but unlocalized, and err
//     contains the localization error.
//   - err: a hard error when option application fails (resolved will be nil),
//     or the localization error when localizationFailed is true.
//
// Callers that treat localization as best-effort should check
// localizationFailed and, when true, proceed with the unlocalized resolved
// error while optionally logging err. Callers that require localization
// should treat any non-nil err as a failure.
func Error(appErr *apperr.AppError, opts ...Option) (resolved *errdetails.ResolvedError, localizationFailed bool, err error) {
	if appErr == nil {
		return nil, false, fmt.Errorf("resolve: appErr must not be nil")
	}

	cfg := &config{
		Logger: errdetails.NewNoopSlogLogger(),
	}
	for _, opt := range opts {
		opt(cfg)
	}

	log := cfg.Logger

	// Accumulate options into an unresolved intermediate.
	var u errdetails.UnresolvedError
	u.Code = appErr.Code()
	u.Message = appErr.Message()

	for _, opt := range appErr.Opts() {
		if applyErr := opt.Apply(&u); applyErr != nil {
			log.Error("failed to apply option", slog.Any("error", applyErr))
			return nil, false, applyErr
		}
	}

	// Resolve localization — helpers populate u.Localized / u.FieldViolations,
	// falling back to raw values on failure.
	if cfg.Provider != nil {
		if lErr := localize(&u, cfg); lErr != nil {
			log.Warn("localization failed, returning resolved error without translation",
				slog.Any("error", lErr))
			localizationFailed = true
			err = lErr
		}
	} else {
		if u.RawLocalized != nil && u.RawLocalized.Any != nil {
			return nil, false, fmt.Errorf("resolve: LocalizeAny requires a LocalizationProvider; there is no fallback for untyped localization values")
		}
		for _, fv := range u.RawFieldViolations {
			if fv.DescriptionAny != nil {
				return nil, false, fmt.Errorf("resolve: FieldViolationLocalizeAny requires a LocalizationProvider; there is no fallback for untyped localization values")
			}
		}
		copyRawLocalizedFallback(&u)
		copyRawFieldViolationsFallback(&u)
	}

	// Project to resolved error — only resolved fields are exposed.
	resolved = &errdetails.ResolvedError{
		Code:                   u.Code,
		Message:                u.Message,
		Headers:                u.Headers,
		Localized:              u.Localized,
		RequestInfo:            u.RequestInfo,
		ResourceInfo:           u.ResourceInfo,
		ErrorInfo:              u.ErrorInfo,
		HelpLinks:              u.HelpLinks,
		FieldViolations:        u.FieldViolations,
		PreconditionViolations: u.PreconditionViolations,
		QuotaViolations:        u.QuotaViolations,
		RetryInfo:              u.RetryInfo,
		DebugInfo:              u.DebugInfo,
	}

	return resolved, localizationFailed, err
}

func localize(u *errdetails.UnresolvedError, cfg *config) error {
	var firstErr error

	if u.RawLocalized != nil {
		if err := localizeText(u, cfg); err != nil {
			copyRawLocalizedFallback(u)
			firstErr = err
		}
	}

	if len(u.RawFieldViolations) > 0 {
		u.FieldViolations = make([]*errdetails.ResolvedFieldViolation, 0, len(u.RawFieldViolations))

		for i, fv := range u.RawFieldViolations {
			err := localizeFieldViolation(u, cfg, fv)
			if err == nil {
				continue
			}

			// Localization failed for this entry — fall back to raw values
			// for this and all remaining field violations, preserving any
			// entries that were already resolved successfully.
			for _, remaining := range u.RawFieldViolations[i:] {
				rfv := &errdetails.ResolvedFieldViolation{
					Field: remaining.Field,
				}
				if remaining.Description != nil {
					rfv.Description = *remaining.Description
				} else if remaining.DescriptionID != nil {
					rfv.Description = *remaining.DescriptionID
				}
				u.FieldViolations = append(u.FieldViolations, rfv)
			}
			if firstErr == nil {
				firstErr = err
			}
			break
		}
	}

	return firstErr
}

func localizeText(u *errdetails.UnresolvedError, cfg *config) error {
	log := cfg.Logger

	var (
		text string
		tag  language.Tag
		nf   bool
		err  error
	)

	if u.RawLocalized.Any != nil {
		text, tag, nf, err = cfg.Provider.LocalizeAny(u.RawLocalized.Any, cfg.Languages)
	} else if u.RawLocalized.TextID != nil {
		text, tag, nf, err = cfg.Provider.Localize(*u.RawLocalized.TextID, cfg.Languages)
	} else {
		log.Warn("neither TextID nor Any are set. Unable to provide translated text")
		return nil
	}
	if err != nil {
		log.Error("localization provider failed", slog.Any("error", err))
		return err
	} else if nf {
		log.Warn("localization provider didn't find message")
		return nil
	}

	u.Localized = &errdetails.LocalizedMessage{
		Locale: tag,
		Text:   text,
	}

	return nil
}

func localizeFieldViolation(u *errdetails.UnresolvedError, cfg *config, fv *errdetails.RawFieldViolation) error {
	log := cfg.Logger

	// description is provided directly -> add unchanged
	if fv.Description != nil {
		u.FieldViolations = append(u.FieldViolations, &errdetails.ResolvedFieldViolation{
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
	if fv.DescriptionAny != nil {
		text, tag, nf, err = cfg.Provider.LocalizeAny(fv.DescriptionAny, cfg.Languages)
	} else if fv.DescriptionID != nil {
		text, tag, nf, err = cfg.Provider.Localize(*fv.DescriptionID, cfg.Languages)
	} else {
		log.Warn("neither DescriptionID nor DescriptionAny are set. Unable to provide translated description for field violation")
		return nil
	}

	// check for errors during translation
	if err != nil {
		log.Error("localization provider failed", slog.Any("error", err))
		return err
	} else if nf {
		log.Warn("localization provider didn't find message")
		return nil
	}

	u.FieldViolations = append(u.FieldViolations, &errdetails.ResolvedFieldViolation{
		Locale:      &tag,
		Field:       fv.Field,
		Description: text,
	})

	return nil
}

func copyRawLocalizedFallback(u *errdetails.UnresolvedError) {
	if u.RawLocalized == nil || u.Localized != nil {
		return
	}
	if u.RawLocalized.TextID != nil {
		u.Localized = &errdetails.LocalizedMessage{
			Text: *u.RawLocalized.TextID,
		}
	}
}

func copyRawFieldViolationsFallback(u *errdetails.UnresolvedError) {
	if len(u.RawFieldViolations) == 0 {
		return
	}
	u.FieldViolations = make([]*errdetails.ResolvedFieldViolation, 0, len(u.RawFieldViolations))
	for _, fv := range u.RawFieldViolations {
		rfv := &errdetails.ResolvedFieldViolation{
			Field: fv.Field,
		}
		if fv.Description != nil {
			rfv.Description = *fv.Description
		} else if fv.DescriptionID != nil {
			rfv.Description = *fv.DescriptionID
		}
		u.FieldViolations = append(u.FieldViolations, rfv)
	}
}
