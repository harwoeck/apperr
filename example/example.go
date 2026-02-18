package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/harwoeck/apperr"
	"github.com/harwoeck/apperr/x/httperr"
	i18n "github.com/harwoeck/apperr/x/i18n"
	nicksnyderI18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
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

	adapter := i18n.NewAdapter(b)

	mw := httperr.NewMiddleware(
		httperr.WithLocalizationProvider(adapter),
		httperr.WithGetClientLanguagesFunc(func(_ context.Context) []language.Tag {
			return []language.Tag{language.English}
		}),
	)

	http.Handle("/unauthenticated", mw(handlerUnauthenticated))
	http.Handle("/internal", mw(handlerUnknownError))

	return http.ListenAndServe("localhost:8080", nil)
}

func handlerUnauthenticated(w http.ResponseWriter, r *http.Request) error {
	return apperr.Unauthenticated("x", apperr.Localize("x"))
}

func handlerUnknownError(w http.ResponseWriter, r *http.Request) error {
	return fmt.Errorf("unknown error")
}
