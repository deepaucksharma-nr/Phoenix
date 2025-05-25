module github.com/phoenix/platform/services/loadsim-operator

go 1.21

require (
    github.com/phoenix/platform/pkg v0.0.0
    k8s.io/api v0.28.3
    k8s.io/apimachinery v0.28.3
    k8s.io/client-go v0.28.3
    sigs.k8s.io/controller-runtime v0.16.3
)

replace github.com/phoenix/platform/pkg => ../../pkg