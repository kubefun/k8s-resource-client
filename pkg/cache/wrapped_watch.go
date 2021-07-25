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
	Listers []ResourceLister
}

var _ ResourceLister = (*WrappedWatchDetails)(nil)

func (w *WrappedWatchDetails) Key() string {
	keys := []string{}
	for _, detail := range w.Listers {
		keys = append(keys, detail.Key())
	}
	return strings.Join(uniqueStringSlice(keys), ",")
}

func (w *WrappedWatchDetails) Namespace() string {
	namespaces := []string{}
	for _, detail := range w.Listers {
		namespaces = append(namespaces, detail.Namespace())
	}
	return strings.Join(uniqueStringSlice(namespaces), ",")
}

func (w *WrappedWatchDetails) List(selector labels.Selector) ([]runtime.Object, error) {
	errors := []string{}
	objects := []runtime.Object{}
	for _, detail := range w.Listers {
		listObjects, err := detail.List(selector)
		if err != nil {
			logging.Logger.Error("failed to list",
				zap.String("resource", detail.Key()),
				zap.String("namespace", detail.Namespace()),
				zap.Error(err),
			)
			errors = append(errors, err.Error())
			continue
		}
		objects = append(objects, listObjects...)
	}
	if len(errors) == 0 {
		return objects, nil
	}
	return objects, fmt.Errorf(strings.Join(errors, ","))
}

func (w *WrappedWatchDetails) Get(name string) (runtime.Object, error) {
	var object runtime.Object
	namespaces := []string{}
	for _, detail := range w.Listers {
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
	for _, detail := range w.Listers {
		detail.Stop()
	}
}

// Drain will get events off of the WatchDetail.Queue and send them to the provided channel.
func (w *WrappedWatchDetails) Drain(ch chan<- interface{}, stopCh chan struct{}) {
	for _, detail := range w.Listers {
		detail.Drain(ch, stopCh)
	}
}

func (w *WrappedWatchDetails) IsRunning() int {
	count := 0
	for _, detail := range w.Listers {
		count += detail.IsRunning()
	}
	return count
}
func uniqueStringSlice(nsSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range nsSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
