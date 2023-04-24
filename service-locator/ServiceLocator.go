package service_locator

import (
	"fmt"
	"reflect"
)

type ServiceEntry struct {
	Instance interface{}
	Type     reflect.Type
}

type ServiceLocator struct {
	services map[string]*ServiceEntry
}

func NewServiceLocator() *ServiceLocator {
	return &ServiceLocator{
		services: make(map[string]*ServiceEntry),
	}
}

func (sl *ServiceLocator) Register(name string, service interface{}) {
	sl.services[name] = &ServiceEntry{
		Instance: service,
		Type:     reflect.TypeOf(service),
	}
}

func (sl *ServiceLocator) GetWithType(name string) (interface{}, reflect.Type, error) {
	entry, exists := sl.services[name]
	if !exists {
		return nil, nil, fmt.Errorf("service not found: %s", name)
	}
	return entry.Instance, entry.Type, nil
}
