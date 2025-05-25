module github.com/phoenix/platform/projects/platform-api

go 1.21

toolchain go1.24.3

require (
	github.com/go-chi/chi/v5 v5.0.10
	go.uber.org/zap v1.26.0
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	go.uber.org/goleak v1.2.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
)

replace github.com/phoenix/platform/pkg => ../../pkg

replace github.com/phoenix/platform/packages/go-common => ../../packages/go-common

replace github.com/phoenix/platform/packages/contracts => ../../packages/contracts
