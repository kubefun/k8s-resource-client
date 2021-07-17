package client

import (
	"github.com/wwitzel3/k8s-resource-client/pkg/errors"
)

func CheckResourceAccess(resource, verb string) error {
	return &errors.FailedSubjectAccessCheck{
		Resource: "Pods",
		Verb:     "Watch",
	}
}
