package clients

import "sync"

type ClientConfig any

type Client interface {
	GetMetrics(*sync.Map)
}

type ClientFactory func(ClientConfig) Client

var clientRegistry = make(map[string]ClientFactory)

func Register(name string, client ClientFactory) {
	clientRegistry[name] = client
}

func GetClient(name string) ClientFactory {
	if client, ok := clientRegistry[name]; ok {
		return client
	}
	return nil
}

func ListClients() []string {
	var names = []string{}
	for name := range clientRegistry {
		names = append(names, name)
	}
	return names
}
