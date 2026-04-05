package twirperr

import (
	"context"
	"errors"
	"log/slog"

	"github.com/harwoeck/apperr"
	"github.com/harwoeck/apperr/errdetails"
	"github.com/harwoeck/apperr/resolve"
	"github.com/twitchtv/twirp"
	"golang.org/x/text/language"
)

// InterceptorOption configures the Twirp interceptor.
type InterceptorOption func(*interceptor)

// WithLocalizationProvider sets the localization provider used to translate
// error messages.
func WithLocalizationProvider(adapter errdetails.LocalizationProvider) InterceptorOption {
	return func(i *interceptor) {
		i.adapter = adapter
	}
}

// WithGetLogFunc sets the function used to obtain a logger from the request
// context. When not set, a noop logger is used.
func WithGetLogFunc(f GetLogFunc) InterceptorOption {
	return func(i *interceptor) {
		i.getLog = f
	}
}

// WithGetClientLanguagesFunc sets the function used to determine the
// preferred language tags for the current request.
func WithGetClientLanguagesFunc(f GetClientLanguagesFunc) InterceptorOption {
	return func(i *interceptor) {
		i.getLangs = f
	}
}

// EnableDebugInfo opts in to forwarding [errdetails.DebugInfo] to clients.
// By default, debug information (stack traces, internal details) is stripped
// from responses before conversion — secure by default. Only enable this in
// development or internal environments where exposing server-side debugging
// information is acceptable.
func EnableDebugInfo() InterceptorOption {
	return func(i *interceptor) {
		i.debugInfo = true
	}
}

// DisableLocalizationBestEffort makes the interceptor treat localization
// failures as hard errors instead of falling back to the unlocalized
// resolved error. By default localization is best-effort.
func DisableLocalizationBestEffort() InterceptorOption {
	return func(i *interceptor) {
		i.localizationBestEffort = false
	}
}

// GetLogFunc returns a logger derived from the request context.
type GetLogFunc func(ctx context.Context) *slog.Logger

// GetClientLanguagesFunc returns the preferred language tags for the
// current request, typically parsed from the Accept-Language header.
type GetClientLanguagesFunc func(ctx context.Context) []language.Tag

type interceptor struct {
	adapter                errdetails.LocalizationProvider
	getLog                 GetLogFunc
	getLangs               GetClientLanguagesFunc
	localizationBestEffort bool
	debugInfo              bool
}

// NewInterceptor creates a twirp.Interceptor that catches *apperr.AppError
// values, resolves them through the localization pipeline, and converts them
// into twirp.Error responses. Errors that are already twirp.Error are passed
// through unchanged. Any other error type is wrapped as an internal error.
//
// All parameters are optional. Without a LocalizationProvider, errors are
// resolved without translation. Without a GetLogFunc, a noop logger is used.
// Without a GetClientLanguagesFunc, no language preference is applied.
func NewInterceptor(opts ...InterceptorOption) twirp.Interceptor {
	noopLog := errdetails.NewNoopSlogLogger()
	i := &interceptor{
		getLog:                 func(_ context.Context) *slog.Logger { return noopLog },
		getLangs:               func(_ context.Context) []language.Tag { return nil },
		localizationBestEffort: true,
	}
	for _, opt := range opts {
		opt(i)
	}

	return func(next twirp.Method) twirp.Method {
		return func(ctx context.Context, request any) (any, error) {
			result, err := next(ctx, request)
			if err == nil {
				return result, nil
			}

			return nil, i.handleError(ctx, err)
		}
	}
}

func (i *interceptor) handleError(ctx context.Context, err error) error {
	log := i.getLog(ctx)

	var te twirp.Error
	if errors.As(err, &te) {
		return te
	}

	var e *apperr.AppError
	if !errors.As(err, &e) {
		log.Warn("unknown error type arrived at twirperr.Interceptor. Using an internal twirp error without any infos attached",
			slog.Any("error", err))
		e = apperr.Internal("")
	}

	resolveOpts := []resolve.Option{resolve.WithLogger(log)}
	if i.adapter != nil {
		resolveOpts = append(resolveOpts,
			resolve.WithLocalizationProvider(i.adapter),
			resolve.WithLanguages(i.getLangs(ctx)),
		)
	}

	resolved, localizationFailed, resolveErr := resolve.Error(e, resolveOpts...)
	if resolveErr != nil {
		if !localizationFailed || !i.localizationBestEffort {
			log.Warn("failed to resolve error, using internal twirp error without any infos attached, as failsafe",
				slog.Any("error", resolveErr))
			return twirp.InternalError("")
		}
		log.Warn("localization failed, continuing with resolved error without translation (best-effort)",
			slog.Any("error", resolveErr))
	}

	te, convertErr := Convert(resolved, i.debugInfo)
	if convertErr != nil {
		log.Warn("failed to convert resolved error to twirp error. using internal twirp error, as failsafe",
			slog.Any("error", convertErr))
		return twirp.InternalError("")
	}

	// Forward resolved headers to the HTTP response via Twirp's
	// context-based header API.
	for key, values := range resolved.Headers {
		for _, v := range values {
			if hdErr := twirp.AddHTTPResponseHeader(ctx, key, v); hdErr != nil {
				log.Warn("failed to set HTTP response header from resolved error",
					slog.String("header", key),
					slog.Any("error", hdErr))
			}
		}
	}

	return te
}
