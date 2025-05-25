module github.com/phoenix/platform/projects/controller

go 1.21

require (
	github.com/lib/pq v1.10.9
	github.com/phoenix/platform/pkg v0.0.0
	go.uber.org/zap v1.26.0
)

replace github.com/phoenix/platform/pkg => ../../pkg

require (
	go.uber.org/multierr v1.11.0 // indirect
)
replace github.com/phoenix/platform/packages/go-common => ../../packages/go-common
replace github.com/phoenix/platform/packages/contracts => ../../packages/contracts
