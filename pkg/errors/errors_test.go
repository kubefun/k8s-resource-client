package errors_test

import (
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
