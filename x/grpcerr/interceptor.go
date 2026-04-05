package grpcerr

import (
	"context"
	"errors"
	"log/slog"

	"github.com/harwoeck/apperr"
	"github.com/harwoeck/apperr/errdetails"
	"github.com/harwoeck/apperr/resolve"
	"golang.org/x/text/language"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// InterceptorOption configures the gRPC server interceptors.
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
// current request, typically parsed from the Accept-Language header or gRPC
// metadata.
type GetClientLanguagesFunc func(ctx context.Context) []language.Tag

type interceptor struct {
	adapter                errdetails.LocalizationProvider
	getLog                 GetLogFunc
	getLangs               GetClientLanguagesFunc
	localizationBestEffort bool
	debugInfo              bool
}

func newInterceptor(opts []InterceptorOption) *interceptor {
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

// NewUnaryServerInterceptor creates a grpc.UnaryServerInterceptor that
// catches *apperr.AppError values, resolves them through the localization
// pipeline, and converts them into gRPC status errors with rich error
// details attached. Errors that are already gRPC status errors are passed
// through unchanged. Any other error type is wrapped as an internal error.
//
// All parameters are optional. Without a LocalizationProvider, errors are
// resolved without translation. Without a GetLogFunc, a noop logger is used.
// Without a GetClientLanguagesFunc, no language preference is applied.
func NewUnaryServerInterceptor(opts ...InterceptorOption) grpc.UnaryServerInterceptor {
	i := newInterceptor(opts)
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)
		if err == nil {
			return resp, nil
		}

		return nil, i.handleError(ctx, err)
	}
}

// NewStreamServerInterceptor creates a grpc.StreamServerInterceptor that
// catches *apperr.AppError values, resolves them through the localization
// pipeline, and converts them into gRPC status errors with rich error
// details attached. Errors that are already gRPC status errors are passed
// through unchanged. Any other error type is wrapped as an internal error.
//
// All parameters are optional. Without a LocalizationProvider, errors are
// resolved without translation. Without a GetLogFunc, a noop logger is used.
// Without a GetClientLanguagesFunc, no language preference is applied.
func NewStreamServerInterceptor(opts ...InterceptorOption) grpc.StreamServerInterceptor {
	i := newInterceptor(opts)
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(srv, ss)
		if err == nil {
			return nil
		}

		return i.handleError(ss.Context(), err)
	}
}

func (i *interceptor) handleError(ctx context.Context, err error) error {
	log := i.getLog(ctx)

	// If the error is already a gRPC status, pass it through.
	var se interface{ GRPCStatus() *status.Status }
	if errors.As(err, &se) {
		return err
	}

	var e *apperr.AppError
	if !errors.As(err, &e) {
		log.Warn("unknown error type arrived at grpcerr interceptor. Using an internal gRPC error without any infos attached",
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
			log.Warn("failed to resolve error, using internal gRPC error without any infos attached, as failsafe",
				slog.Any("error", resolveErr))
			return status.Error(codes.Internal, "")
		}
		log.Warn("localization failed, continuing with resolved error without translation (best-effort)",
			slog.Any("error", resolveErr))
	}

	st, convertErr := Convert(resolved, i.debugInfo)
	if convertErr != nil {
		log.Warn("failed to convert resolved error to gRPC status. using internal gRPC error, as failsafe",
			slog.Any("error", convertErr))
		return status.Error(codes.Internal, "")
	}

	// Forward resolved headers as gRPC response metadata.
	if len(resolved.Headers) > 0 {
		md := metadata.MD(resolved.Headers)
		if mdErr := grpc.SetHeader(ctx, md); mdErr != nil {
			log.Warn("failed to set gRPC response headers from resolved error",
				slog.Any("error", mdErr))
		}
	}

	return st.Err()
}
