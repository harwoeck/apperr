module github.com/harwoeck/apperr/x/grpcerr

go 1.25.0

require (
	github.com/harwoeck/apperr v0.0.0
	golang.org/x/text v0.34.0
	google.golang.org/grpc v1.79.1
	google.golang.org/protobuf v1.36.11
)

require google.golang.org/genproto/googleapis/rpc v0.0.0-20260217215200-42d3e9bedb6d // indirect

require (
	github.com/harwoeck/apperr/x/protoerr v0.0.0
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
)

replace github.com/harwoeck/apperr v0.0.0 => ../../

replace github.com/harwoeck/apperr/x/protoerr v0.0.0 => ../protoerr
