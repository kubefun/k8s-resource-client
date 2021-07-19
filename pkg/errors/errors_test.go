package errors_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wwitzel3/k8s-resource-client/pkg/errors"
)

func TestFailedSubjectAccessCheckError(t *testing.T) {
	err := &errors.FailedSubjectAccessCheck{
		Resource: "Pods",
		Verb:     "Watch",
	}

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "FailedSubjectAccessCheck - resource:Pods, verb:Watch")
}

func TestResourceNotSyncedError(t *testing.T) {
	err := &errors.ResourceNotSynced{
		Reason: "explicit mode set and resource is not listed",
	}

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "ResourceNotSynced - reason:explicit mode set and resource is not listed")
}

func TestNilRESTConfigError(t *testing.T) {
	err := &errors.NilRESTConfig{}

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "NilRESTConfig - cannot create client")
}

func TestK8SNewForConfig(t *testing.T) {
	err := &errors.K8SNewForConfig{Err: fmt.Errorf("test")}

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "K8SNewForConfig - test")
}

func TestNamespaceDiscoveryError(t *testing.T) {
	err := &errors.NamespaceDiscoveryError{Err: fmt.Errorf("test")}

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "NamespaceDiscoveryError - test")
}

func TestResourceDiscoveryError(t *testing.T) {
	err := &errors.ResourceDiscoveryError{Err: []error{fmt.Errorf("test"), fmt.Errorf("test2")}}

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "ResourceDiscoveryError - [test test2]")

	rdErr := &errors.ResourceDiscoveryError{Err: []error{fmt.Errorf("test"), fmt.Errorf("test2")}}
	rdErr.Add(fmt.Errorf("test3"))
	assert.Equal(t, rdErr.Error(), "ResourceDiscoveryError - [test test2 test3]")
}
