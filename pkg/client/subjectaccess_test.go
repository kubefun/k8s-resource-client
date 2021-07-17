package client_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wwitzel3/k8s-resource-client/pkg/client"
	_errors "github.com/wwitzel3/k8s-resource-client/pkg/errors"
)

func TestCheckResourceAccess(t *testing.T) {
	err := client.CheckResourceAccess("pod", "list")

	var e *_errors.FailedSubjectAccessCheck
	assert.True(t, errors.As(err, &e))
}
