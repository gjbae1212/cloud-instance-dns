package server

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		input string
		err   bool
	}{
		"empty": {input: "", err: true},
	}

	for _, t := range tests {
		_, err := NewServer(t.input)
		if t.err {
			assert.NotNil(err)
		} else {
			assert.Nil(err)
		}
	}

	yamlPath := os.Getenv("TEST_YAML_PATH")
	if yamlPath != "" {
		_, err := NewServer(yamlPath)
		assert.NoError(err)
	}
}
