module github.com/harwoeck/apperr/x/twirperr

go 1.25.0

require (
	github.com/harwoeck/apperr v0.0.0
	github.com/twitchtv/twirp v8.1.3+incompatible
	golang.org/x/text v0.34.0
)

require github.com/pkg/errors v0.9.1 // indirect

replace github.com/harwoeck/apperr v0.0.0 => ../../
