package errors

import (
	"context"
	"fmt"
)

type FailedSubjectAccessCheck struct {
	Resource string
	Verb     string
}

func (e *FailedSubjectAccessCheck) Error() string {
	return fmt.Sprintf("FailedSubjectAccessCheck - resource:%v, verb:%v", e.Resource, e.Verb)
}

type ResourceNotSynced struct {
	Reason string
}

func (e *ResourceNotSynced) Error() string {
	return fmt.Sprintf("ResourceNotSynced - reason:%v", e.Reason)
}

type NilRESTConfig struct {
}

func (e *NilRESTConfig) Error() string {
	return "NilRESTConfig - cannot create client"
}

type K8SNewForConfig struct {
	Err error
}

func (e *K8SNewForConfig) Error() string {
	return fmt.Sprintf("K8SNewForConfig - %s", e.Err)
}

func NewK8SNewForConfig(_ context.Context, err error) error {
	if err != nil {
		return &K8SNewForConfig{Err: err}
	}
	return nil
}
