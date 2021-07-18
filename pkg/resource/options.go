package resource

import (
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ResourceAccessOption func(*resourceAccess)

func WithLogger(logger *zap.Logger) ResourceAccessOption {
	return func(r *resourceAccess) {
		r.logger = logger
	}
}

func WithMinimumRBAC(verbs metav1.Verbs) ResourceAccessOption {
	return func(r *resourceAccess) {
		r.minimumVerbs = verbs
	}
}
