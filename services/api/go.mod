module github.com/phoenix/platform/services/api

go 1.21

require (
	github.com/go-chi/chi/v5 v5.0.10
	github.com/google/uuid v1.4.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.18.0
	github.com/phoenix/platform/packages/go-common v0.0.0
	github.com/prometheus/client_golang v1.18.0
	go.uber.org/zap v1.26.0
	google.golang.org/grpc v1.60.1
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/matttproud/golang_protobuf_extensions/v2 v2.0.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.45.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20231120223509-83a465c0220f // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231212172506-995d672761c0 // indirect
	google.golang.org/protobuf v1.32.0 // indirect
)

replace (
	github.com/phoenix/platform/packages/contracts => ../../packages/contracts
	github.com/phoenix/platform/packages/go-common => ../../packages/go-common
)
