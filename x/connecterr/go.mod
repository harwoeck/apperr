module github.com/harwoeck/apperr/x/connecterr

go 1.25.0

require (
	connectrpc.com/connect v1.18.1
	github.com/harwoeck/apperr v0.0.0
	golang.org/x/text v0.34.0
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260217215200-42d3e9bedb6d // indirect
)

require github.com/harwoeck/apperr/x/protoerr v0.0.0

require google.golang.org/protobuf v1.36.11 // indirect

replace github.com/harwoeck/apperr v0.0.0 => ../../

replace github.com/harwoeck/apperr/x/protoerr v0.0.0 => ../protoerr
