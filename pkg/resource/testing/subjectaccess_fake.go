package testing

import (
	"context"
	"fmt"

	v1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	authv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
)

var _ authv1.SelfSubjectAccessReviewInterface = (*SubjectAccessFake)(nil)

type SubjectAccessFake struct {
	CreateFn func(*SubjectAccessFake) (*v1.SelfSubjectAccessReview, error)
}

func (s SubjectAccessFake) Create(ctx context.Context, selfSubjectAccessReview *v1.SelfSubjectAccessReview, opts metav1.CreateOptions) (*v1.SelfSubjectAccessReview, error) {
	if s.CreateFn != nil {
		return s.CreateFn(&s)
	}
	return nil, fmt.Errorf("default fake error")
}
