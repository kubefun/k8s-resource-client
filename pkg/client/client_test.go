package client_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/stretchr/testify/assert"

	"github.com/wwitzel3/k8s-resource-client/pkg/client"
	"github.com/wwitzel3/k8s-resource-client/pkg/errors"
)

func TestNewClientNoRestConfig(t *testing.T) {
	_, err := client.NewClient(context.TODO())
	assert.ErrorIs(t, err, &errors.NilRESTConfig{})
}

func TestNewClientClientsetFnErr(t *testing.T) {
	cFn := func(ctx context.Context, _ *rest.Config) (kubernetes.Interface, error) {
		return nil, &errors.K8SNewForConfig{Err: fmt.Errorf("bad clientset")}
	}
	config := &rest.Config{QPS: 400, Burst: 800}

	_, err := client.NewClient(context.TODO(), client.WithClientsetFn(cFn), client.WithRESTConfig(config))
	assert.EqualError(t, err, "K8SNewForConfig - bad clientset")
}

func TestNewClientDynamicClientFnErr(t *testing.T) {
	cFn := func(ctx context.Context, _ *rest.Config) (dynamic.Interface, error) {
		return nil, &errors.K8SNewForConfig{Err: fmt.Errorf("bad dynamic")}
	}
	config := &rest.Config{QPS: 400, Burst: 800}

	_, err := client.NewClient(context.TODO(), client.WithDynamicClientFn(cFn), client.WithRESTConfig(config))
	assert.EqualError(t, err, "K8SNewForConfig - bad dynamic")
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

	c, err := client.NewClient(context.TODO(), client.WithRESTConfig(config), client.WithLogger(logger))
	if err != nil {
		t.Fatal(err)
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
	c, err := client.NewClient(context.TODO(), client.WithRESTConfig(config))
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, c)
	assert.Equal(t, c.NamespaceMode, client.Auto)
	assert.Equal(t, c.ResourceMode, client.Auto)
	assert.Equal(t, c.SkipSubjectAccessChecks, false)
}

func TestNewClientOptions(t *testing.T) {
	config := &rest.Config{QPS: 400, Burst: 800}
	c, err := client.NewClient(context.TODO(),
		client.WithNamespaceMode(client.Explicit),
		client.WithResourceMode(client.Explicit),
		client.WithSkipSubjectAccessChecks(true),
		client.WithRESTConfig(config),
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
