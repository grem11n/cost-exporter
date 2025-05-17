package outputs

import (
	"sync"
)

// OutputConfig contains the config for each Output
type OutputConfig any

// Output an interface to output the collected metrics
type Output interface {
	Publish(*sync.Map, []string)
}

type OutputFactory func(OutputConfig) Output //nolint:revive

var outputRegistry = make(map[string]OutputFactory)

// Register an Output by name
func Register(name string, output OutputFactory) {
	outputRegistry[name] = output
}

// GetOutput returns an Output by its name
func GetOutput(name string) OutputFactory {
	if output, ok := outputRegistry[name]; ok {
		return output
	}
	return nil
}

// ListOutputs returns a list of all registered outputs
func ListOutputs() []string {
	var names = []string{}
	for name := range outputRegistry {
		names = append(names, name)
	}
	return names
}
