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

func TestServer_Lookup(t *testing.T) {
	assert := assert.New(t)

	yamlPath := os.Getenv("TEST_YAML_PATH")
	if yamlPath != "" {
		s, err := NewServer(yamlPath)
		assert.NoError(err)
		records, err := s.(*server).Lookup("empty")
		assert.NoError(err)
		assert.Len(records, 0)

		records, err = s.(*server).Lookup("1.empty")
		assert.NoError(err)
		assert.Len(records, 0)

		records, err = s.(*server).Lookup("empty.aws")
		assert.NoError(err)
		assert.Len(records, 0)

		records, err = s.(*server).Lookup("empty.gcp")
		assert.NoError(err)
		assert.Len(records, 0)

		records, err = s.(*server).Lookup("1.empty.gcp")
		assert.NoError(err)
		assert.Len(records, 0)

		records, err = s.(*server).Lookup("1.empty.aws")
		assert.NoError(err)
		assert.Len(records, 0)

		records, err = s.(*server).Lookup("\".....aaaabb..rn**^^()#$#_!")
		assert.NoError(err)
		assert.Len(records, 0)
	}
}

func TestServer_Start(t *testing.T) {
	assert := assert.New(t)
	yamlPath := os.Getenv("TEST_YAML_PATH")
	if yamlPath != "" {
		s, err := NewServer(yamlPath)
		assert.NoError(err)
		go s.Start()
	}
}
