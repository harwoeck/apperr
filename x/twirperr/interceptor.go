package twirperr

import (
	"context"
	"time"

	"github.com/harwoeck/apperr/utils/dto"

	"github.com/harwoeck/liblog"
	"github.com/twitchtv/twirp"
	"golang.org/x/text/language"

	"github.com/harwoeck/apperr"
	"github.com/harwoeck/apperr/utils/finalizer"
)

type GetLogFunc func(ctx context.Context) liblog.Logger

type GetClientLanguagesOrDefaultFunc func(ctx context.Context) []language.Tag

func Interceptor(adapter finalizer.LocalizationProvider, getLogFunc GetLogFunc, getClientLanguagesOrDefaultFunc GetClientLanguagesOrDefaultFunc) twirp.Interceptor {
	return func(next twirp.Method) twirp.Method {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			var (
				log       = getLogFunc(ctx)
				languages = getClientLanguagesOrDefaultFunc(ctx)
			)

			serverStart := time.Now()

			result, err := next(ctx, request)
			if err == nil {
				return result, nil
			}
			
			serverEnd := time.Now()

			var e *apperr.AppError
			switch t := err.(type) {
			case twirp.Error:
				return nil, t
			case *apperr.AppError:
				e = t
			default:
				log.Debug("unknown error type arrived at twirperr.Interceptor. Using an internal twirp error without any infos attached",
					liblog.NewField("error", err))
				e = apperr.Internal("")
			}

			rendered, err := finalizer.Render(e,
				finalizer.WithLogger(log),
				finalizer.WithLocalizationProvider(adapter),
				finalizer.WithLanguages(languages),
			)
			if err != nil {
				log.Warn("failed to render localized error. trying again without localization provider",
					liblog.NewField("error", err))

				rendered, err = finalizer.Render(e, finalizer.WithLogger(log))
				if err != nil {
					log.Warn("failed to render error. using twirp.InternalError without any infos attached, as failsafe")
					return nil, twirp.InternalError("")
				}
			}

			if rendered.RequestInfo == nil {
				rendered.RequestInfo = &dto.RequestInfo{}
			}
			dur := serverEnd.Sub(serverStart)
			rendered.RequestInfo.RequestDuration = &dur

			if rendered.ErrorInfo != nil && len(rendered.ErrorInfo.Domain) == 0 {
				rendered.ErrorInfo.Domain = "localhost:8080"
			}

			return nil, Convert(rendered)
		}
	}
}
