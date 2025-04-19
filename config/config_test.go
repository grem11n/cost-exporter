package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyConfig(t *testing.T) {
	emptyCfg := &Config{}
	err := emptyCfg.populateDefaults()
	assert.ErrorContains(t, err, "client configuration is required. Only AWS is supported")
}
