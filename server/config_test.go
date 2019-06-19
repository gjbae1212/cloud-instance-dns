package server

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseConfig(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		input        map[interface{}]interface{}
		commonConfig *CommonConfig
		awsConfig    *AwsConfig
		gcpConfig    *GcpConfig
		err          bool
	}{
		"empty":       {input: nil, err: true},
		"emptyDomain": {input: make(map[interface{}]interface{}), err: true},
		"success": {input: map[interface{}]interface{}{
			"domain": "localhost"}, err: false, commonConfig: &CommonConfig{domain: "localhost."}},
	}

	for _, t := range tests {
		co, ac, gc, err := ParseConfig(t.input)
		if co != nil {
			assert.Equal(t.commonConfig.domain, co.domain)
		}
		assert.Equal(t.awsConfig, ac)
		assert.Equal(t.gcpConfig, gc)
		if t.err {
			assert.Error(err)
		} else {
			assert.NoError(err)
		}
	}

	yamlPath := os.Getenv("TEST_YAML_PATH")
	if yamlPath != "" {
		config := make(map[interface{}]interface{})
		bys, err := ioutil.ReadFile(yamlPath)
		assert.NoError(err)
		err = yaml.Unmarshal(bys, &config)
		assert.NoError(err)

		// both enable
		co, ac, gc, err := ParseConfig(config)
		assert.NoError(err)
		assert.NotEqual("", co.domain)
		assert.True(strings.HasSuffix(co.domain, "."))
		assert.NotEqual("", co.nameserver)
		assert.NotEqual("", co.port)
		assert.NotEqual("", co.rname)
		assert.Equal(config["private"], co.private)
		assert.NotEmpty(ac)
		assert.NotEmpty(gc)
		assert.NotEqual(0, len(ac.clients))
		assert.NotEmpty(gc.projectId)
		assert.NotEmpty(gc.zones)
		assert.NotEmpty(gc.client)

		// gcp enable off
		config["gcp"].(map[interface{}]interface{})["enable"] = "false"
		co, ac, gc, err = ParseConfig(config)
		assert.NoError(err)
		assert.NotEqual("", co.domain)
		assert.NotEmpty(ac)
		assert.Empty(gc)

		// aws enable off
		config["gcp"].(map[interface{}]interface{})["enable"] = "true"
		config["aws"].(map[interface{}]interface{})["enable"] = "false"
		co, ac, gc, err = ParseConfig(config)
		assert.NoError(err)
		assert.NotEqual("", co.domain)
		assert.Empty(ac)
		assert.NotEmpty(gc)
	}
}
