module github.com/phoenix/platform/projects/platform-api

go 1.21

toolchain go1.24.3

require github.com/go-chi/chi/v5 v5.0.10

replace github.com/phoenix/platform/pkg => ../../pkg

replace github.com/phoenix/platform/packages/go-common => ../../packages/go-common

replace github.com/phoenix/platform/packages/contracts => ../../packages/contracts
