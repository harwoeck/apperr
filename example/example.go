package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/harwoeck/liblog"
	nicksnyderI18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"

	"github.com/harwoeck/apperr"
	"github.com/harwoeck/apperr/utils/finalizer"
	"github.com/harwoeck/apperr/x/httperr"
	i18n "github.com/harwoeck/apperr/x/nicksnyder-i18n"
)

func main() {
	err := run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getI18nBundle() (*nicksnyderI18n.Bundle, error) {
	t, err := language.Parse("en-US")
	if err != nil {
		return nil, err
	}

	b := nicksnyderI18n.NewBundle(t)
	b.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	_, err = b.LoadMessageFile("en-US.toml")
	if err != nil {
		return nil, err
	}

	_, err = b.LoadMessageFile("de-DE.toml")
	if err != nil {
		return nil, err
	}

	return b, nil
}

func run() error {
	b, err := getI18nBundle()
	if err != nil {
		return err
	}

	adapter := i18n.NewI18nAdapter(b)

	http.Handle("/unauthenticated", middleware(adapter, handlerUnauthenticated))
	http.Handle("/internal", middleware(adapter, handlerUnknownError))

	return http.ListenAndServe("localhost:8080", nil)
}

func middleware(adapter finalizer.LocalizationProvider, handler func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serverStart := time.Now()

		err := handler(w, r)
		if err == nil {
			return
		}

		serverEnd := time.Now()

		// convert all errors to AppError
		var ae *apperr.AppError
		var ok bool
		if ae, ok = err.(*apperr.AppError); !ok {
			ae = apperr.Internal("internal server error", apperr.Localize("INTERNAL"))
		}

		requestStart, _ := time.Parse(time.RFC3339Nano, r.Header.Get("X-Request-Start"))

		dur := serverEnd.Sub(serverStart)
		latency := serverStart.Sub(requestStart)

		// append data we want on every error returned to our clients, like the request-id
		ae.AppendOptions(
			apperr.RequestInfo("some-random-request-uuid", &dur, "", &latency),
		)

		rendered, err := finalizer.Render(ae,
			finalizer.WithLogger(liblog.MustNewStd()),
			finalizer.WithLocalizationProvider(adapter),
			finalizer.WithLanguages([]language.Tag{language.English}),
		)
		if err != nil {
			panic(err)
		}

		status, body, err := httperr.Convert(nil, rendered)
		if err != nil {
			panic(err)
		}

		w.WriteHeader(status)
		_, err = w.Write(body)
		if err != nil {
			panic(err)
		}
	}
}

func handlerUnauthenticated(w http.ResponseWriter, r *http.Request) error {
	return apperr.Unauthenticated("x", apperr.Localize("x"))
}

func handlerUnknownError(w http.ResponseWriter, r *http.Request) error {
	return fmt.Errorf("unknown error")
}
