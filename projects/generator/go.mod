module github.com/phoenix/platform/projects/generator

go 1.21

require (
	github.com/phoenix/platform/pkg v0.0.0-00010101000000-000000000000
	github.com/go-chi/chi/v5 v5.1.0
	google.golang.org/grpc v1.68.1
	go.uber.org/zap v1.27.0
	google.golang.org/protobuf v1.35.1
)

require (
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/sys v0.24.0 // indirect
	golang.org/x/text v0.17.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240814211410-ddb44dafa142 // indirect
)

replace github.com/phoenix/platform/pkg => ../../pkg