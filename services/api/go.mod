module github.com/phoenix-vnext/platform/services/api

go 1.21

require (
	github.com/go-chi/chi/v5 v5.0.10
	github.com/google/uuid v1.6.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.20.0
	github.com/phoenix/platform/packages/go-common v0.0.0
	github.com/prometheus/client_golang v1.19.0
	go.uber.org/zap v1.26.0
	google.golang.org/grpc v1.64.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240701130421-f6361c86f094 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240701130421-f6361c86f094 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)

replace (
	github.com/phoenix/platform/packages/contracts => ../../packages/contracts
	github.com/phoenix/platform/packages/go-common => ../../packages/go-common
)
