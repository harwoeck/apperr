package httperr

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/harwoeck/apperr"
	"github.com/harwoeck/apperr/errdetails"
	"github.com/harwoeck/apperr/resolve"
	"golang.org/x/text/language"
)

// MiddlewareOption configures the HTTP error-handling middleware.
type MiddlewareOption func(*middleware)

// WithLocalizationProvider sets the localization provider used to translate
// error messages.
func WithLocalizationProvider(adapter errdetails.LocalizationProvider) MiddlewareOption {
	return func(m *middleware) {
		m.adapter = adapter
	}
}

// WithGetLogFunc sets the function used to obtain a logger from the request
// context. When not set, a noop logger is used.
func WithGetLogFunc(f GetLogFunc) MiddlewareOption {
	return func(m *middleware) {
		m.getLog = f
	}
}

// WithGetClientLanguagesFunc sets the function used to determine the
// preferred language tags for the current request, typically parsed from the
// Accept-Language header.
func WithGetClientLanguagesFunc(f GetClientLanguagesFunc) MiddlewareOption {
	return func(m *middleware) {
		m.getLangs = f
	}
}

// EnableDebugInfo opts in to forwarding [errdetails.DebugInfo] to clients.
// By default, debug information (stack traces, internal details) is stripped
// from responses before conversion — secure by default. Only enable this in
// development or internal environments where exposing server-side debugging
// information is acceptable.
func EnableDebugInfo() MiddlewareOption {
	return func(m *middleware) {
		m.debugInfo = true
	}
}

// DisableLocalizationBestEffort makes the middleware treat localization
// failures as hard errors instead of falling back to the unlocalized
// resolved error. By default localization is best-effort.
func DisableLocalizationBestEffort() MiddlewareOption {
	return func(m *middleware) {
		m.localizationBestEffort = false
	}
}

// GetLogFunc returns a logger derived from the request context.
type GetLogFunc func(ctx context.Context) *slog.Logger

// GetClientLanguagesFunc returns the preferred language tags for the
// current request, typically parsed from the Accept-Language header.
type GetClientLanguagesFunc func(ctx context.Context) []language.Tag

// HandlerFunc is like [http.HandlerFunc] but returns an error. The
// middleware catches any returned error, resolves it, and writes the
// appropriate JSON error response. A nil return means the handler has
// already written a successful response.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

type middleware struct {
	adapter                errdetails.LocalizationProvider
	getLog                 GetLogFunc
	getLangs               GetClientLanguagesFunc
	localizationBestEffort bool
	debugInfo              bool
}

// NewMiddleware creates HTTP middleware that catches [*apperr.AppError]
// values returned by a [HandlerFunc], resolves them through the
// localization pipeline, and writes a JSON error response with the correct
// HTTP status code and Content-Type header. Any other error type is wrapped
// as an internal error.
//
// All options are optional. Without a [WithLocalizationProvider], errors are
// resolved without translation. Without a [WithGetLogFunc], a noop logger
// is used. Without a [WithGetClientLanguagesFunc], no language preference is
// applied.
func NewMiddleware(opts ...MiddlewareOption) func(HandlerFunc) http.Handler {
	noopLog := errdetails.NewNoopSlogLogger()
	m := &middleware{
		getLog:                 func(_ context.Context) *slog.Logger { return noopLog },
		getLangs:               func(_ context.Context) []language.Tag { return nil },
		localizationBestEffort: true,
	}
	for _, opt := range opts {
		opt(m)
	}

	return func(handler HandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := handler(w, r)
			if err == nil {
				return
			}

			m.handleError(w, r, err)
		})
	}
}

func (m *middleware) handleError(w http.ResponseWriter, r *http.Request, err error) {
	log := m.getLog(r.Context())

	var e *apperr.AppError
	if !errors.As(err, &e) {
		log.Warn("unknown error type arrived at httperr.Middleware. Using an internal error without any infos attached",
			slog.Any("error", err))
		e = apperr.Internal("")
	}

	resolveOpts := []resolve.Option{resolve.WithLogger(log)}
	if m.adapter != nil {
		resolveOpts = append(resolveOpts,
			resolve.WithLocalizationProvider(m.adapter),
			resolve.WithLanguages(m.getLangs(r.Context())),
		)
	}

	resolved, localizationFailed, resolveErr := resolve.Error(e, resolveOpts...)
	if resolveErr != nil {
		if !localizationFailed || !m.localizationBestEffort {
			log.Warn("failed to resolve error, using plain 500 as failsafe",
				slog.Any("error", resolveErr))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		log.Warn("localization failed, continuing with resolved error without translation (best-effort)",
			slog.Any("error", resolveErr))
	}

	status, body, convertErr := Convert(resolved, m.debugInfo)
	if convertErr != nil {
		log.Warn("failed to convert resolved error to HTTP response. using plain 500 as failsafe",
			slog.Any("error", convertErr))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Forward resolved headers to the HTTP response.
	for key, values := range resolved.Headers {
		for _, v := range values {
			w.Header().Add(key, v)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
