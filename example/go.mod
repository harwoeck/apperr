module github.com/harwoeck/apperr/example

go 1.16

require (
	github.com/harwoeck/apperr/adapter/i18n v0.0.0
	github.com/harwoeck/apperr/apperr v0.0.0
	github.com/harwoeck/apperr/httperr v0.0.0
	github.com/harwoeck/liblog/contract v1.1.2
	github.com/nicksnyder/go-i18n/v2 v2.1.2
	github.com/pelletier/go-toml v1.9.3
	golang.org/x/text v0.3.6
)

replace (
	github.com/harwoeck/apperr/adapter/i18n => ../adapter/i18n
	github.com/harwoeck/apperr/apperr => ../apperr
	github.com/harwoeck/apperr/httperr => ../httperr
)
