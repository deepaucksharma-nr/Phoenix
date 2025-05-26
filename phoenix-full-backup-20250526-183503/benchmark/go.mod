module github.com/phoenix/platform/projects/benchmark

go 1.23.0

toolchain go1.24.3

require (
	github.com/prometheus/client_golang v1.19.0
	github.com/prometheus/common v0.48.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_model v0.6.0 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/oauth2 v0.27.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
)

replace github.com/phoenix/platform/pkg => ../../pkg

replace github.com/phoenix/platform/pkg/common => ../../pkg/common

replace github.com/phoenix/platform/pkg/contracts => ../../pkg/contracts
