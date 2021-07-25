package cache

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	kcache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/wwitzel3/k8s-resource-client/pkg/resource"
)

var (
	DefaultResyncDuration = time.Second * 180
	ResourceWatches       = &sync.Map{} // sync.Map{"resourceKey": [sync.Map{"namespace.resourceKey":"watchDetail"}]}
)

// WatchDetail holds the details of an Informer and Lister for a specific resource.
// An optionally configured event queue.
type WatchDetail struct {
	Informer kcache.SharedInformer
	StopCh   chan struct{}
	Resource resource.Resource
	Logger   *zap.Logger

	queueEvents bool
	Queue       *workqueue.Type

	namespace string
	informer  informers.GenericInformer
}

var _ ResourceLister = (*WatchDetail)(nil)

func (w *WatchDetail) Key() string {
	return fmt.Sprintf("%s.%s", w.namespace, w.Resource.Key())
}

func (w *WatchDetail) Namespace() string {
	return w.namespace
}

func (w *WatchDetail) List(selector labels.Selector) ([]runtime.Object, error) {
	if w.namespace == metav1.NamespaceAll {
		return w.informer.Lister().List(selector)
	}
	return w.informer.Lister().ByNamespace(w.namespace).List(selector)
}

func (w *WatchDetail) Get(name string) (runtime.Object, error) {
	if w.namespace == metav1.NamespaceAll {
		return w.informer.Lister().Get(name)
	}
	return w.informer.Lister().ByNamespace(w.namespace).Get(name)
}

// IsRunning returns true if the Informer loop for the WatchDetail is running.
func (w *WatchDetail) IsRunning() int {
	select {
	case <-w.StopCh:
		return 0
	default:
		return 1
	}
}

// Stop closes the StopCh shutting down the Drain and Informer loops.
func (w *WatchDetail) Stop() {
	if w.IsRunning() == 1 {
		close(w.StopCh)
	}
}

// Drain will get events off of the WatchDetail.Queue and send them to the provided channel.
func (w *WatchDetail) Drain(ch chan<- interface{}, stopCh chan struct{}) {
	go func() {
		for {
			select {
			// Main stopCh for errors/controllers
			case <-w.StopCh:
				w.Logger.Debug("main stopCh closed")
				return
			// Local stopCh for callers
			case <-stopCh:
				w.Logger.Debug("local stopCh closed")
				return
			default:
				i, shutdown := w.Queue.Get()
				if shutdown {
					w.Logger.Debug("processing queue and shutting down")
					ch <- i
					return
				}
				w.Logger.Debug("processing queue")
				ch <- i
			}
		}
	}()
}
