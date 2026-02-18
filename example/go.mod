module github.com/harwoeck/apperr/example

go 1.25.0

require (
	github.com/BurntSushi/toml v1.6.0
	github.com/harwoeck/apperr v0.0.0
	github.com/harwoeck/apperr/x/httperr v0.0.0
	github.com/harwoeck/apperr/x/i18n v0.0.0
	github.com/nicksnyder/go-i18n/v2 v2.6.1
	golang.org/x/text v0.34.0
)

replace (
	github.com/harwoeck/apperr => ../
	github.com/harwoeck/apperr/x/httperr => ../x/httperr
	github.com/harwoeck/apperr/x/i18n => ../x/i18n
)
