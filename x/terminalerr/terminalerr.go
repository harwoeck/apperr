package terminalerr

import (
	"fmt"
	"log/slog"

	"github.com/harwoeck/apperr/errdetails"
)

// Log logs a resolved apperr error using the provided slog.Logger. The
// error details are structured using slog groups to mirror the JSON layout
// of ResolvedError. The log level is slog.LevelError.
//
// The keepDebugInfo parameter controls whether [errdetails.DebugInfo] is
// included in the log output. Pass false to omit server-side debugging
// information from logs.
func Log(log *slog.Logger, resolved *errdetails.ResolvedError, keepDebugInfo bool) {
	attrs := []slog.Attr{
		slog.String("code", resolved.Code.String()),
		slog.String("message", resolved.Message),
	}

	if resolved.Localized != nil {
		attrs = append(attrs, slog.Group("localized",
			slog.String("locale", resolved.Localized.Locale.String()),
			slog.String("text", resolved.Localized.Text),
		))
	}

	if resolved.ErrorInfo != nil {
		metaAttrs := make([]any, 0, len(resolved.ErrorInfo.Metadata)*2+4)
		metaAttrs = append(metaAttrs,
			slog.String("reason", resolved.ErrorInfo.Reason),
			slog.String("domain", resolved.ErrorInfo.Domain),
		)
		for k, v := range resolved.ErrorInfo.Metadata {
			metaAttrs = append(metaAttrs, slog.String(k, v))
		}
		attrs = append(attrs, slog.Group("errorInfo", metaAttrs...))
	}

	if resolved.RequestInfo != nil {
		riAttrs := []any{
			slog.String("requestId", resolved.RequestInfo.RequestID),
			slog.String("servingData", resolved.RequestInfo.ServingData),
		}
		attrs = append(attrs, slog.Group("requestInfo", riAttrs...))
	}

	if resolved.ResourceInfo != nil {
		attrs = append(attrs, slog.Group("resourceInfo",
			slog.String("type", resolved.ResourceInfo.Type),
			slog.String("name", resolved.ResourceInfo.Name),
			slog.String("owner", resolved.ResourceInfo.Owner),
			slog.String("description", resolved.ResourceInfo.Description),
		))
	}

	if len(resolved.FieldViolations) > 0 {
		fvAttrs := make([]any, 0, len(resolved.FieldViolations))
		for i, fv := range resolved.FieldViolations {
			group := []any{
				slog.String("field", fv.Field),
				slog.String("description", fv.Description),
			}
			if fv.Locale != nil {
				group = append(group, slog.String("locale", fv.Locale.String()))
			}
			fvAttrs = append(fvAttrs, slog.Group(fmt.Sprintf("%d", i), group...))
		}
		attrs = append(attrs, slog.Group("fieldViolations", fvAttrs...))
	}

	if len(resolved.PreconditionViolations) > 0 {
		pvAttrs := make([]any, 0, len(resolved.PreconditionViolations))
		for i, pv := range resolved.PreconditionViolations {
			pvAttrs = append(pvAttrs, slog.Group(fmt.Sprintf("%d", i),
				slog.String("type", pv.Type),
				slog.String("subject", pv.Subject),
				slog.String("description", pv.Description),
			))
		}
		attrs = append(attrs, slog.Group("preconditionViolations", pvAttrs...))
	}

	if len(resolved.QuotaViolations) > 0 {
		qvAttrs := make([]any, 0, len(resolved.QuotaViolations))
		for i, qv := range resolved.QuotaViolations {
			qvAttrs = append(qvAttrs, slog.Group(fmt.Sprintf("%d", i),
				slog.String("subject", qv.Subject),
				slog.String("description", qv.Description),
			))
		}
		attrs = append(attrs, slog.Group("quotaViolations", qvAttrs...))
	}

	if len(resolved.HelpLinks) > 0 {
		hlAttrs := make([]any, 0, len(resolved.HelpLinks))
		for i, link := range resolved.HelpLinks {
			hlAttrs = append(hlAttrs, slog.Group(fmt.Sprintf("%d", i),
				slog.String("url", link.URL),
				slog.String("description", link.Description),
			))
		}
		attrs = append(attrs, slog.Group("helpLinks", hlAttrs...))
	}

	if resolved.RetryInfo != nil {
		attrs = append(attrs, slog.Group("retryInfo",
			slog.Int("delay_ms", resolved.RetryInfo.Delay),
		))
	}

	if keepDebugInfo && resolved.DebugInfo != nil {
		diAttrs := []any{
			slog.String("detail", resolved.DebugInfo.Detail),
		}
		for i, entry := range resolved.DebugInfo.StackEntries {
			diAttrs = append(diAttrs, slog.String(fmt.Sprintf("stack.%d", i), entry))
		}
		attrs = append(attrs, slog.Group("debugInfo", diAttrs...))
	}

	args := make([]any, len(attrs))
	for i, a := range attrs {
		args[i] = a
	}
	log.Error("resolved error", args...)
}
