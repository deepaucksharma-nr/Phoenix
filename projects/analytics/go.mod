module github.com/phoenix/platform/projects/analytics

go 1.24.0

toolchain go1.24.3

require (
	github.com/phoenix/platform/pkg v0.0.0
	github.com/stretchr/testify v1.10.0
	go.uber.org/zap v1.27.0
	google.golang.org/grpc v1.68.1
	google.golang.org/protobuf v1.36.5
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241223144023-3abc09e42ca8 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/phoenix/platform/pkg => ../../pkg
