package terminalerr

import (
	"testing"

	"golang.org/x/text/language"

	"github.com/harwoeck/apperr/utils/code"
	"github.com/harwoeck/apperr/utils/finalizer"
)

func TestConvert(t *testing.T) {
	type args struct {
		rendered *finalizer.Error
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"simple", args{&finalizer.Error{
			Code:    code.DataLoss,
			Message: "data corrupted",
		}}, "DataLoss: data corrupted"},
		{"localized", args{&finalizer.Error{
			Code:    code.DataLoss,
			Message: "data corrupted",
			Localized: &finalizer.Localized{
				Locale: language.German,
				Title:  "Daten verloren",
				Text:   "Daten sind verloren oder nicht mehr lesbar",
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
