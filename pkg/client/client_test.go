package client_test

import (
	"fmt"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"k8s.io/client-go/rest"

	"github.com/stretchr/testify/assert"

	"github.com/wwitzel3/k8s-resource-client/pkg/client"
	"github.com/wwitzel3/k8s-resource-client/pkg/errors"
)

func TestNewClientNoRestConfig(t *testing.T) {
	_, err := client.NewClient()
	assert.ErrorIs(t, err, &errors.NilRESTConfig{})
}

func TestNewClientRestConfigWarnings(t *testing.T) {
	burstWarning := false
	qpsWarning := false

	config := &rest.Config{QPS: 200, Burst: 400}
	logger := zaptest.NewLogger(t, zaptest.WrapOptions(zap.Hooks(func(e zapcore.Entry) error {
		fmt.Println(e.Message, e.Level)
		if strings.Contains(e.Message, "QPS") && e.Level == zap.WarnLevel {
			qpsWarning = true
		} else if strings.Contains(e.Message, "Burst") && e.Level == zap.WarnLevel {
			burstWarning = true
		}
		return nil
	})))

	c, err := client.NewClient(client.WithRestClientConfig(config), client.WithLogger(logger))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	assert.NotNil(t, c)
	assert.Equal(t, c.NamespaceMode, client.Auto)
	assert.Equal(t, c.ResourceMode, client.Auto)
	assert.Equal(t, c.SkipSubjectAccessChecks, false)
	assert.True(t, qpsWarning)
	assert.True(t, burstWarning)
}
func TestNewClient(t *testing.T) {
	config := &rest.Config{QPS: 400, Burst: 800}
	c, err := client.NewClient(client.WithRestClientConfig(config))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	assert.NotNil(t, c)
	assert.Equal(t, c.NamespaceMode, client.Auto)
	assert.Equal(t, c.ResourceMode, client.Auto)
	assert.Equal(t, c.SkipSubjectAccessChecks, false)
}

func TestNewClientOptions(t *testing.T) {
	config := &rest.Config{QPS: 400, Burst: 800}
	c, err := client.NewClient(
		client.WithNamespaceMode(client.Explicit),
		client.WithResourceMode(client.Explicit),
		client.WithSkipSubjectAccessChecks(true),
		client.WithRestClientConfig(config),
	)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	assert.NotNil(t, c)
	assert.Equal(t, c.NamespaceMode, client.Explicit)
	assert.Equal(t, c.ResourceMode, client.Explicit)
	assert.Equal(t, c.SkipSubjectAccessChecks, true)
}
