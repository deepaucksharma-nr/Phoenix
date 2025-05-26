module github.com/phoenix/platform/projects/analytics

go 1.23.0

toolchain go1.24.3

require (
	github.com/go-chi/chi/v5 v5.0.11
	github.com/go-chi/cors v1.2.1
	github.com/prometheus/client_golang v1.19.0
	github.com/prometheus/common v0.48.0
	github.com/sirupsen/logrus v1.9.3
	gonum.org/v1/gonum v0.14.0
	gonum.org/v1/plot v0.14.0
)

replace github.com/phoenix/platform/pkg => ../../pkg

require (
	git.sr.ht/~sbinet/gg v0.5.0 // indirect
	github.com/ajstarks/svgo v0.0.0-20211024235047-1546f124cd8b // indirect
	github.com/campoy/embedmd v1.0.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/go-fonts/liberation v0.3.1 // indirect
	github.com/go-latex/latex v0.0.0-20230307184459-12ec69307ad9 // indirect
	github.com/go-pdf/fpdf v0.8.0 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	golang.org/x/exp v0.0.0-20240719175910-8a7402abbf56 // indirect
	golang.org/x/image v0.11.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/oauth2 v0.27.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
)

replace github.com/phoenix/platform/packages/go-common => ../../packages/go-common

replace github.com/phoenix/platform/packages/contracts => ../../packages/contracts
