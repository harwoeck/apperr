module github.com/harwoeck/apperr/example

go 1.18

require (
	github.com/harwoeck/apperr v0.0.0
	github.com/harwoeck/apperr/x/httperr v0.0.0
	github.com/harwoeck/apperr/x/nicksnyder-i18n v0.0.0
	github.com/harwoeck/liblog v1.2.0
	github.com/nicksnyder/go-i18n/v2 v2.1.2
	github.com/pelletier/go-toml v1.9.3
	golang.org/x/text v0.3.7
)

replace (
	github.com/harwoeck/apperr => ../
	github.com/harwoeck/apperr/x/httperr => ../x/httperr
	github.com/harwoeck/apperr/x/nicksnyder-i18n => ../x/nicksnyder-i18n
)
