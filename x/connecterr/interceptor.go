package connecterr

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/harwoeck/apperr"
	"github.com/harwoeck/apperr/errdetails"
	"github.com/harwoeck/apperr/resolve"
	"golang.org/x/text/language"
)

// InterceptorOption configures the connect interceptor.
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

// WithEnrichFunc sets the function used to enrich errors with additional
// request-scoped context before they are returned to the client.
func WithEnrichFunc(f EnrichFunc) InterceptorOption {
	return func(i *interceptor) {
		i.enrich = f
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

// EnrichFunc is called before an error is returned to the client, allowing
// additional context from the request (e.g. request ID) to be attached to the
// error via err.AppendOptions.
type EnrichFunc func(ctx context.Context, err *apperr.AppError)

type interceptor struct {
	adapter                errdetails.LocalizationProvider
	getLog                 GetLogFunc
	getLangs               GetClientLanguagesFunc
	localizationBestEffort bool
	debugInfo              bool
	enrich                 EnrichFunc
}

// NewInterceptor creates a connect.Interceptor that catches *apperr.AppError
// values, resolves them through the localization pipeline, and converts them
// into *connect.Error responses. Errors that are already *connect.Error are
// passed through unchanged. Any other error type is wrapped as an internal
// error.
//
// All parameters are optional. Without a LocalizationProvider, errors are
// resolved without translation. Without a GetLogFunc, a noop logger is used.
// Without a GetClientLanguagesFunc, no language preference is applied.
func NewInterceptor(opts ...InterceptorOption) connect.Interceptor {
	noopLog := errdetails.NewNoopSlogLogger()
	i := &interceptor{
		getLog:                 func(_ context.Context) *slog.Logger { return noopLog },
		getLangs:               func(_ context.Context) []language.Tag { return nil },
		localizationBestEffort: true,
	}
	for _, opt := range opts {
		opt(i)
	}
	return i
}

func (i *interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		resp, err := next(ctx, req)
		if err == nil {
			return resp, nil
		}

		return nil, i.handleError(ctx, err)
	}
}

func (i *interceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	// Pass through – this interceptor is server-side only.
	return next
}

func (i *interceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		err := next(ctx, conn)
		if err == nil {
			return nil
		}

		return i.handleError(ctx, err)
	}
}

func (i *interceptor) handleError(ctx context.Context, err error) error {
	log := i.getLog(ctx)

	var ce *connect.Error
	if errors.As(err, &ce) {
		return ce
	}

	var e *apperr.AppError
	if !errors.As(err, &e) {
		log.Warn("unknown error type arrived at connecterr.Interceptor. Using an internal connect error without any infos attached",
			slog.Any("error", err))
		e = apperr.Internal("")
	}

	if i.enrich != nil {
		i.enrich(ctx, e)
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
			log.Warn("failed to resolve error, using internal connect error without any infos attached, as failsafe",
				slog.Any("error", resolveErr))
			return connect.NewError(connect.CodeInternal, nil)
		}
		log.Warn("localization failed, continuing with resolved error without translation (best-effort)",
			slog.Any("error", resolveErr))
	}

	ce, convertErr := Convert(resolved, i.debugInfo)
	if convertErr != nil {
		log.Warn("failed to convert resolved error to connect error. using internal connect error, as failsafe",
			slog.Any("error", convertErr))
		return connect.NewError(connect.CodeInternal, nil)
	}

	return ce
}
