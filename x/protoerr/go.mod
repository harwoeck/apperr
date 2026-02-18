module github.com/harwoeck/apperr/x/protoerr

go 1.25.0

require (
	github.com/harwoeck/apperr v0.0.0
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260217215200-42d3e9bedb6d
	google.golang.org/protobuf v1.36.11
)

require golang.org/x/text v0.34.0 // indirect

replace github.com/harwoeck/apperr v0.0.0 => ../../
