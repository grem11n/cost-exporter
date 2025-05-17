package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyConfig(t *testing.T) {
	emptyCfg := &Config{}
	err := emptyCfg.populateDefaults()
	assert.ErrorIs(t, err, ErrClientConfig)
}
