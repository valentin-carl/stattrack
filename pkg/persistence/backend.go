package persistence

import (
	_ "log"

	_ "github.com/valentin-carl/stattrack/pkg/measurements"
)

type Backend interface {
	//Store(measurements.Measurement) error
	Start() error
}
