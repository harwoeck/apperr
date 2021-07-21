package main

import (
	"fmt"
	"net/http"
	"os"

	logger "github.com/harwoeck/liblog/contract"
	nicksnyderI18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pelletier/go-toml"
	"golang.org/x/text/language"

	"github.com/harwoeck/apperr/adapter/i18n"
	"github.com/harwoeck/apperr/apperr"
	"github.com/harwoeck/apperr/httperr"
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
	_, err = b.LoadMessageFile("de-DE.toml")
	if err != nil {
		return nil, err
	}
	return b, nil
}

func run() error {
	http.Handle("/unauthenticated", middleware(handlerUnauthenticated))
	http.Handle("/internal", middleware(handlerUnknownError))
	return http.ListenAndServe("localhost:8080", nil)
}

func middleware(handler func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	b, err := getI18nBundle()
	if err != nil {
		panic(err)
	}
	adapter := i18n.NewI18nAdapter(b)

	return func(w http.ResponseWriter, r *http.Request) {
		var ae *apperr.AppError
		switch err := handler(w, r).(type) {
		case *apperr.AppError:
			ae = err
		default:
			ae = apperr.Internal("internal server error", apperr.Localize("INTERNAL"))
		}

		rendered, err := apperr.Render(ae,
			apperr.EnableLogging(logger.MustNewStd()),
			apperr.RenderLocalized(adapter, r.Header.Get("Accept-Language")),
		)
		if err != nil {
			panic(err)
		}

		status, body, err := httperr.Convert(rendered)
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
