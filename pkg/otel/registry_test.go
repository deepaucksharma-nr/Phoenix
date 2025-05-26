package otel

import "testing"

type dummyFactory struct{}

func (dummyFactory) Type() string                                     { return "dummy" }
func (dummyFactory) Create(map[string]interface{}) (Processor, error) { return nil, nil }

func TestRegisterProcessorFactory(t *testing.T) {
	ClearProcessorFactories()
	RegisterProcessorFactory(dummyFactory{})
	if _, ok := GetProcessorFactory("dummy"); !ok {
		t.Fatalf("factory not registered")
	}
}
