package converters

import "sync"

type Conveter interface {
	Convert(*sync.Map)
}

type ConveterFactory func() Conveter

var converterRegistry = make(map[string]ConveterFactory)

func Register(name string, conveter ConveterFactory) {
	converterRegistry[name] = conveter
}

func GetConverter(name string) ConveterFactory {
	if converter, ok := converterRegistry[name]; ok {
		return converter
	}
	return nil
}
