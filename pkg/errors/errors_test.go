package errors_test

import (
	"context"
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
	var err error
	err = &errors.K8SNewForConfig{Err: fmt.Errorf("test")}

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "K8SNewForConfig - test")

	err = errors.NewK8SNewForConfig(context.TODO(), nil)
	assert.Nil(t, err)

	err = errors.NewK8SNewForConfig(context.TODO(), fmt.Errorf("test"))
	assert.EqualError(t, err, "K8SNewForConfig - test")
}
