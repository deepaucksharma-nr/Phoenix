module github.com/phoenix/platform/projects/loadsim-operator

go 1.21

require (
	k8s.io/apimachinery v0.29.0
	k8s.io/client-go v0.29.0
	sigs.k8s.io/controller-runtime v0.16.3
)

replace github.com/phoenix/platform/packages/go-common => ../../packages/go-common