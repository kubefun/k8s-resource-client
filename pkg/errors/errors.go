package errors

import (
	"fmt"
	"sync"
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
	return "NilRESTConfig - cannot create client, use WithRESTConfig option"
}

type K8SNewForConfig struct {
	Err error
}

func (e *K8SNewForConfig) Error() string {
	return fmt.Sprintf("K8SNewForConfig - %s", e.Err)
}

type NamespaceDiscoveryError struct {
	Err error
}

func (e *NamespaceDiscoveryError) Error() string {
	return fmt.Sprintf("NamespaceDiscoveryError - %s", e.Err)
}

type ResourceDiscoveryError struct {
	Err []error
	mu  sync.Mutex
}

func (e *ResourceDiscoveryError) Error() string {
	return fmt.Sprintf("ResourceDiscoveryError - %v", e.Err)
}

func (e *ResourceDiscoveryError) Add(err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Err = append(e.Err, err)
}
