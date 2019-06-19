package server

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewStore(t *testing.T) {
	assert := assert.New(t)

	yamlPath := os.Getenv("TEST_YAML_PATH")
	if yamlPath != "" {
		config := make(map[interface{}]interface{})
		bys, err := ioutil.ReadFile(yamlPath)
		assert.NoError(err)
		yaml.Unmarshal(bys, &config)
		assert.NoError(err)
		_, awsconfig, gcpconfig, err := ParseConfig(config)
		assert.NoError(err)

		store, err := NewStore(awsconfig, gcpconfig)
		assert.NoError(err)
		_ = store
	}

}

func TestStore_Lookup(t *testing.T) {
	assert := assert.New(t)

	yamlPath := os.Getenv("TEST_YAML_PATH")
	if yamlPath != "" {
		config := make(map[interface{}]interface{})
		bys, err := ioutil.ReadFile(yamlPath)
		assert.NoError(err)
		yaml.Unmarshal(bys, &config)
		assert.NoError(err)
		_, awsconfig, gcpconfig, err := ParseConfig(config)
		assert.NoError(err)

		store, err := NewStore(awsconfig, gcpconfig)
		assert.NoError(err)

		// empty
		records, err := store.Lookup("empty")
		assert.NoError(err)
		assert.Len(records, 0)

		if os.Getenv("TEST_AWS_1") != "" {
			records, err := store.Lookup(os.Getenv("TEST_AWS_1"))
			assert.NoError(err)
			for _, record := range records {
				assert.NotEmpty(record.PublicIP.String())
				assert.NotEmpty(record.PrivateIP.String())
				assert.NotEmpty(record.ZoneOrRegion)
				assert.Equal(AWS, record.Vendor)
			}
		}

		if os.Getenv("TEST_GCP_1") != "" {
			records, err := store.Lookup(os.Getenv("TEST_GCP_1"))
			assert.NoError(err)
			for _, record := range records {
				assert.NotEmpty(record.PublicIP.String())
				assert.NotEmpty(record.PrivateIP.String())
				assert.NotEmpty(record.ZoneOrRegion)
				assert.Equal(GCP, record.Vendor)
			}
		}
	}

}

func TestRecord_TTL(t *testing.T) {
	assert := assert.New(t)

	yamlPath := os.Getenv("TEST_YAML_PATH")
	if yamlPath != "" {
		config := make(map[interface{}]interface{})
		bys, err := ioutil.ReadFile(yamlPath)
		assert.NoError(err)
		yaml.Unmarshal(bys, &config)
		assert.NoError(err)
		_, awsconfig, gcpconfig, err := ParseConfig(config)
		assert.NoError(err)

		store, err := NewStore(awsconfig, gcpconfig)
		assert.NoError(err)

		table, ok := store.cache.Load(CacheName)
		assert.True(ok)
		for _, v := range table.(LookupTable) {
			for _, re := range v {
				assert.True(re.TTL() > (5 * time.Second))
			}
		}
	}
}

//go test github.com/gjbae1212/cloud-instance-dns/server -bench=.
func BenchmarkStore_Lookup(b *testing.B) {
	yamlPath := os.Getenv("TEST_YAML_PATH")
	if yamlPath != "" {
		config := make(map[interface{}]interface{})
		bys, _ := ioutil.ReadFile(yamlPath)
		yaml.Unmarshal(bys, &config)
		_, awsconfig, gcpconfig, _ := ParseConfig(config)
		store, _ := NewStore(awsconfig, gcpconfig)
		for i := 0; i < b.N; i++ {
			store.Lookup(os.Getenv("TEST_AWS_1"))
		}
	}
}
