package client_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wwitzel3/k8s-resource-client/pkg/client"
)

func TestNewClient(t *testing.T) {
	c := client.NewClient()
	assert.NotNil(t, c)
	assert.Equal(t, c.NamespaceMode, client.Auto)
	assert.Equal(t, c.ResourceMode, client.Auto)
	assert.Equal(t, c.SkipSubjectAccessChecks, false)
}

func TestNewClientOptions(t *testing.T) {
	c := client.NewClient(
		client.WithNamespaceMode(client.Explicit),
		client.WithResourceMode(client.Explicit),
		client.WithSkipSubjectAccessChecks(true),
	)
	assert.NotNil(t, c)
	assert.Equal(t, c.NamespaceMode, client.Explicit)
	assert.Equal(t, c.ResourceMode, client.Explicit)
	assert.Equal(t, c.SkipSubjectAccessChecks, true)
}
