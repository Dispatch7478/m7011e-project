package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	// tmp config file
	content := `
services:
  - name: "user-service"
    url: "http://localhost:8081"
    proxy:
      path: "/api/users"
      rewrite: "/users"
`
	tmpfile, err := os.CreateTemp("", "config_*.yaml")
	require.NoError(t, err)
	// clean up
	defer os.Remove(tmpfile.Name()) 

	_, err = tmpfile.Write([]byte(content))
	require.NoError(t, err)
	tmpfile.Close()

	// test loading the config
	config, err := LoadConfig(tmpfile.Name())

	require.NoError(t, err)
	require.NotNil(t, config)
	assert.Equal(t, 1, len(config.Services))
	assert.Equal(t, "user-service", config.Services[0].Name)
	assert.Equal(t, "/api/users", config.Services[0].Proxy.Path)
}

func TestLoadConfig_NotFound(t *testing.T) {
	_, err := LoadConfig("non-existent-file.yaml")
	assert.Error(t, err)
}