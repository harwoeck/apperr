package terminalerr

import (
	"testing"

	"golang.org/x/text/language"

	"github.com/harwoeck/apperr/apperr"
	"github.com/harwoeck/apperr/apperr/code"
)

func TestConvert(t *testing.T) {
	type args struct {
		rendered *apperr.RenderedError
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"simple", args{&apperr.RenderedError{
			Code:    code.DataLoss,
			Message: "data corrupted",
		}}, "DataLoss: data corrupted"},
		{"localized", args{&apperr.RenderedError{
			Code:    code.DataLoss,
			Message: "data corrupted",
			Localized: &apperr.RenderedErrorLocalized{
				UserMessage:      "Daten sind verloren oder nicht mehr lesbar",
				UserMessageShort: "Daten verloren",
				Locale:           language.German,
			},
		}}, "DataLoss: data corrupted ([de] Daten verloren: Daten sind verloren oder nicht mehr lesbar)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Convert(tt.args.rendered); got != tt.want {
				t.Errorf("Convert() = %v, want %v", got, tt.want)
			}
		})
	}
}
