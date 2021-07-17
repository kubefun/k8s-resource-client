package errors

import "fmt"

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
