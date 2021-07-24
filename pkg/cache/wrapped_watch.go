package cache

import (
	"fmt"
	"strings"

	"github.com/wwitzel3/k8s-resource-client/pkg/logging"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

type WrappedWatchDetails struct {
	listers []ResourceLister
}

var _ ResourceLister = (*WrappedWatchDetails)(nil)

func (w *WrappedWatchDetails) Key() string {
	keys := []string{}
	for _, detail := range w.listers {
		keys = append(keys, detail.Key())
	}
	return strings.Join(keys, ",")
}

func (w *WrappedWatchDetails) Namespace() string {
	namespaces := []string{}
	for _, detail := range w.listers {
		namespaces = append(namespaces, detail.Namespace())
	}
	return strings.Join(namespaces, ",")
}

func (w *WrappedWatchDetails) List(selector labels.Selector) ([]runtime.Object, error) {
	objects := []runtime.Object{}
	for _, detail := range w.listers {
		listObjects, err := detail.List(selector)
		if err != nil {
			logging.Logger.Error("failed to list",
				zap.String("resource", detail.Key()),
				zap.String("namespace", detail.Namespace()),
				zap.Error(err),
			)
			continue
		}
		objects = append(objects, listObjects...)
	}
	return objects, nil
}

func (w *WrappedWatchDetails) Get(name string) (runtime.Object, error) {
	var object runtime.Object
	namespaces := []string{}
	for _, detail := range w.listers {
		getObj, err := detail.Get(name)
		if err != nil {
			namespaces = append(namespaces, detail.Namespace())
			continue
		}
		object = getObj
	}
	if object == nil {
		return nil, fmt.Errorf("unable to find object %s in any namespace of: %+v", name, namespaces)
	}
	return object, nil
}

// Stop closes the StopCh shutting down the Drain and Informer loops.
func (w *WrappedWatchDetails) Stop() {
	for _, detail := range w.listers {
		detail.Stop()
	}
}

// Drain will get events off of the WatchDetail.Queue and send them to the provided channel.
func (w *WrappedWatchDetails) Drain(ch chan<- interface{}, stopCh chan struct{}) {
	for _, detail := range w.listers {
		detail.Drain(ch, stopCh)
	}
}
