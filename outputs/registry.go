package outputs

import (
	"sync"
)

type OutputConfig any

type Output interface {
	Publish([]string, *sync.Map)
}

type OutputFactory func(OutputConfig) Output

var outputRegistry = make(map[string]OutputFactory)

func Register(name string, output OutputFactory) {
	outputRegistry[name] = output
}

func GetOutput(name string) OutputFactory {
	if output, ok := outputRegistry[name]; ok {
		return output
	}
	return nil
}

func ListOutputs() []string {
	var names = []string{}
	for name := range outputRegistry {
		names = append(names, name)
	}
	return names
}
