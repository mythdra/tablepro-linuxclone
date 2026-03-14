package driver

import (
	"fmt"
)

var drivers = make(map[DatabaseType]func() DatabaseDriver)

func RegisterDriver(dbType DatabaseType, factory func() DatabaseDriver) {
	drivers[dbType] = factory
}

func NewDriver(dbType DatabaseType) (DatabaseDriver, error) {
	factory, ok := drivers[dbType]
	if !ok {
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
	return factory(), nil
}

func IsSupported(dbType DatabaseType) bool {
	_, ok := drivers[dbType]
	return ok
}

func SupportedDrivers() []DatabaseType {
	types := make([]DatabaseType, 0, len(drivers))
	for dbType := range drivers {
		types = append(types, dbType)
	}
	return types
}
