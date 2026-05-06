package converters

import "sync"

type Converter interface {
	Convert(*sync.Map, string)
}

type ConverterFactory func() Converter

var converterRegistry = make(map[string]ConverterFactory)

func Register(name string, converter ConverterFactory) {
	converterRegistry[name] = converter
}

func GetConverter(name string) ConverterFactory {
	if converter, ok := converterRegistry[name]; ok {
		return converter
	}
	return nil
}
