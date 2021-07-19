package client_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	typedAuthv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
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

func TestNewClientServerResourcesFnErr(t *testing.T) {
	srFn := func(_ context.Context, clientset kubernetes.Interface) (discovery.ServerResourcesInterface, error) {
		return nil, fmt.Errorf("server resources err")
	}
	config := &rest.Config{QPS: 400, Burst: 800}

	_, err := client.NewClient(context.TODO(), client.WithServerResourcesFn(srFn), client.WithRESTConfig(config))
	assert.EqualError(t, err, "server resources err")
}

func TestNewClientSubjectAccessFnErr(t *testing.T) {
	saFn := func(_ context.Context, clientset kubernetes.Interface) (typedAuthv1.SelfSubjectAccessReviewInterface, error) {
		return nil, fmt.Errorf("subject access error")
	}
	config := &rest.Config{QPS: 400, Burst: 800}

	_, err := client.NewClient(context.TODO(), client.WithSubjectAccessFn(saFn), client.WithRESTConfig(config))
	assert.EqualError(t, err, "subject access error")
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
	ctx := context.TODO()

	config := &rest.Config{QPS: 400, Burst: 800}
	clientset, err := client.NewClientset(ctx, config)

	clientsetFn := func(context.Context, *rest.Config) (kubernetes.Interface, error) {
		return clientset, err
	}

	srFn := func(_ context.Context, clientset kubernetes.Interface) (discovery.ServerResourcesInterface, error) {
		return clientset.Discovery(), nil
	}

	saFn := func(_ context.Context, clientset kubernetes.Interface) (typedAuthv1.SelfSubjectAccessReviewInterface, error) {
		return clientset.AuthorizationV1().SelfSubjectAccessReviews(), nil
	}

	c, err := client.NewClient(context.TODO(),
		client.WithNamespaceMode(client.Explicit),
		client.WithResourceMode(client.Explicit),
		client.WithSkipSubjectAccessChecks(true),
		client.WithRESTConfig(config),
		client.WithClientsetFn(clientsetFn),
		client.WithServerResourcesFn(srFn),
		client.WithSubjectAccessFn(saFn),
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

func TestNewClientset(t *testing.T) {
	config := &rest.Config{
		RateLimiter: nil,
		QPS:         100,
		Burst:       -10,
	}
	_, err := client.NewClientset(context.TODO(), config)
	assert.EqualError(t, err, "K8SNewForConfig - burst is required to be greater than 0 when RateLimiter is not set and QPS is set to greater than 0")
}

func TestNewDynamic(t *testing.T) {
	config := &rest.Config{
		WarningHandler: rest.NewWarningWriter(nil, rest.WarningWriterOptions{}),
		Host:           "ftp:///bad.host.org",
	}
	_, err := client.NewDynamicClient(context.TODO(), config)
	assert.EqualError(t, err, "K8SNewForConfig - host must be a URL or a host:port pair: \"ftp:///bad.host.org\"")
}

func TestNewClientFuncs(t *testing.T) {
	_, err := client.NewServerResources(context.TODO(), nil)
	assert.EqualError(t, err, "nil client.clientset")

	_, err = client.NewSubjectAccess(context.TODO(), nil)
	assert.EqualError(t, err, "nil client.clientset")
}
