package subjectaccess_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	_errors "github.com/wwitzel3/k8s-resource-client/pkg/errors"
	"github.com/wwitzel3/k8s-resource-client/pkg/subjectaccess"
)

func TestCheckResourceAccess(t *testing.T) {
	err := subjectaccess.CheckResourceAccess("pod", "list")

	var e *_errors.FailedSubjectAccessCheck
	assert.True(t, errors.As(err, &e))
}
