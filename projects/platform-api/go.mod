module github.com/phoenix-vnext/platform/projects/platform-api

go 1.21

toolchain go1.24.3

require (
	github.com/go-chi/chi/v5 v5.0.10
	github.com/phoenix-vnext/platform/packages/go-common v0.0.0
	go.uber.org/zap v1.26.0
)

require go.uber.org/multierr v1.11.0 // indirect

replace github.com/phoenix-vnext/platform/pkg => ../../pkg

replace github.com/phoenix-vnext/platform/packages/go-common => ../../packages/go-common

replace github.com/phoenix-vnext/platform/packages/contracts => ../../packages/contracts
