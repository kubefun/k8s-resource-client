package cache

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	kcache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/wwitzel3/k8s-resource-client/pkg/logging"
	"github.com/wwitzel3/k8s-resource-client/pkg/resource"
)

// Watcher holds referenecs to the Kubernetes types and a logger.
// Use NewWatcher to create instances of Watcher.
type Watcher struct {
	dclient         dynamic.Interface
	informerFactory dynamicinformer.DynamicSharedInformerFactory
	namespace       string
	logger          *zap.Logger
}

// NewWatcher creates a Watcher object. This object is used to hold the reference
// to the Kubernetes types that implement Informers and Listers.
func NewWatcher(ctx context.Context, options ...WatcherOption) (*Watcher, error) {
	w := &Watcher{}

	for _, opt := range options {
		opt(w)
	}

	if w.logger == nil {
		w.logger = zap.NewNop()
	}

	if w.dclient == nil {
		return nil, fmt.Errorf("dynamic client nil, use WithDynamicClient option")
	}

	return w, nil
}

// Watch creates a new WatchDetail and starts the watch loop for the given Resource
// If queueEvents is true, all events for the resource will be added to the WatcheDetail.Queue
// To handle the events use WatchDetail.Drain
func (w *Watcher) Watch(ctx context.Context, namespace string, res resource.Resource, queueEvents bool) (ResourceLister, error) {
	if w.namespace != "" && namespace != w.namespace {
		return nil, fmt.Errorf("unable to create watch, resource namespace:%s does not match watcher namespace:%s", namespace, w.namespace)
	}

	lister, err := WatchForResource(res, namespace)
	if err == nil {
		return lister, nil
	}

	if namespace != "" && w.informerFactory == nil {
		w.informerFactory = dynamicinformer.NewFilteredDynamicSharedInformerFactory(w.dclient, DefaultResyncDuration, namespace, nil)
	} else if w.informerFactory == nil {
		w.informerFactory = dynamicinformer.NewDynamicSharedInformerFactory(w.dclient, DefaultResyncDuration)
	}

	genericInformer := w.informerFactory.ForResource(res.GroupVersionResource())

	detail := &WatchDetail{
		namespace:   namespace,
		Resource:    res,
		informer:    genericInformer,
		queueEvents: queueEvents,
		Queue:       workqueue.NewNamed(res.Key()),
		StopCh:      make(chan struct{}),
		Logger:      w.logger,
	}

	// boardcast function that will publish changes to a channel for clients
	if detail.queueEvents {
		genericInformer.Informer().AddEventHandler(kcache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				w.logger.Debug("watch add",
					zap.String("obj", fmt.Sprintf("%v", obj)),
				)
				detail.Queue.Add(obj)
			},
			DeleteFunc: func(obj interface{}) {
				w.logger.Debug("watch delete",
					zap.String("obj", fmt.Sprintf("%v", obj)),
				)
				detail.Queue.Done(obj)
			},
			UpdateFunc: func(new, old interface{}) {
				w.logger.Debug("watch update",
					zap.String("obj", fmt.Sprintf("%v", new)),
				)
				detail.Queue.Add(new)
			},
		})
	}

	detail.informer.Informer().SetWatchErrorHandler(WatchErrorHandlerFactory(w.logger, detail.Key(), detail.StopCh))

	go func() {
		w.logger.Debug("starting informer",
			zap.String("key", detail.Key()),
		)
		detail.informer.Informer().Run(detail.StopCh)
	}()

	if err := appendResourceWatches(res.Key(), detail); err != nil {
		return detail, err
	}

	return detail, nil
}

func appendResourceWatches(key string, detail *WatchDetail) error {
	v, ok := ResourceWatches.Load(key)
	if !ok {
		detailMap := &sync.Map{}
		detailMap.Store(detail.Key(), detail)
		ResourceWatches.Store(key, detailMap)
		return nil
	}
	detailMap, ok := v.(*sync.Map)
	if !ok {
		return fmt.Errorf("append, found key: %s, unable to cast to []*WatchDetail", key)
	}
	detailMap.Store(detail.Key(), detail)
	ResourceWatches.Store(key, detailMap)
	return nil
}

// WatcherStop stops all running watchers.
func WatcherStop() {
	ResourceWatches.Range(func(k, v interface{}) bool {
		value, ok := v.(*sync.Map)
		if !ok {
			return true
		}
		value.Range(func(k, v interface{}) bool {
			detailValue, ok := v.(*WatchDetail)
			if !ok {
				return false
			}
			detailValue.Stop()
			return true
		})
		return true
	})
}

// WatchForResource returns a WatchDetail for the given Resource.
func WatchForResource(r resource.Resource, namespaces ...string) (ResourceLister, error) {
	v, ok := ResourceWatches.Load(r.Key())
	if !ok {
		return nil, fmt.Errorf("no watch found for resource: %+v", r)
	}

	detailMap, ok := v.(*sync.Map)
	if !ok {
		return nil, fmt.Errorf("watch, found key:%s, unable to cast to *sync.Map", r.Key())
	}

	mapValues := []*WatchDetail{}
	detailMap.Range(func(k, v interface{}) bool {
		if v, ok := v.(*WatchDetail); ok {
			mapValues = append(mapValues, v)
		}
		return true
	})

	wrappedWatches := []ResourceLister{}
	if len(namespaces) == 0 { // no explict namespace, use all
		logging.Logger.Info("no namespaces provided using NamespaceAll", zap.String("resource", r.Key()))
		listers := []ResourceLister{}
		for _, detail := range mapValues {
			listers = append(listers, detail)
		}
		wrappedWatches = listers
	} else if len(namespaces) == 1 && namespaces[0] == metav1.NamespaceAll { // only one namespace and it is all, use all
		logging.Logger.Info("only NamespaceAll in namespace list", zap.String("resource", r.Key()))
		listers := []ResourceLister{}
		for _, detail := range mapValues {
			listers = append(listers, detail)
		}
		wrappedWatches = listers
	} else {
		for _, ns := range namespaces {
			if ns == metav1.NamespaceAll { // encountered NamespaceAll, use all
				logging.Logger.Info("found NamespaceAll in namespace list", zap.String("resource", r.Key()), zap.String("namespace", ns))
				listers := []ResourceLister{}
				for _, detail := range mapValues {
					listers = append(listers, detail)
				}
				wrappedWatches = listers
				break
			}
			for _, wd := range mapValues {
				if wd.Namespace() == metav1.NamespaceAll {
					filterDetail := &FilteredWatchDetail{Detail: wd, namespace: ns}
					wrappedWatches = append(wrappedWatches, filterDetail)
					logging.Logger.Info("found NamespaceAll creating filtered watch detail", zap.String("resource", r.Key()), zap.String("namespace", ns))
					continue
				}

				if wd.Namespace() == ns {
					wrappedWatches = append(wrappedWatches, wd)
					logging.Logger.Info("found watcher for namespace", zap.String("resource", r.Key()), zap.String("namespace", ns))
					continue
				}
			}
		}
	}

	if len(wrappedWatches) == 0 {
		return nil, fmt.Errorf("no matching watch found for resource: %s in namespaces: %+v", r.Key(), namespaces)
	}
	return &WrappedWatchDetails{Listers: wrappedWatches}, nil
}

// WatchList returns the current list of watchers from the cache.
// If onlyRunning is true, the list will only include running watchers.
func WatchList(onlyRunning bool) []ResourceLister {
	watches := []ResourceLister{}
	ResourceWatches.Range(func(k, v interface{}) bool {
		value, ok := v.(*sync.Map)
		if !ok {
			return false
		}
		value.Range(func(k, v interface{}) bool {
			detailValue, ok := v.(*WatchDetail)
			if !ok {
				return false
			}
			if onlyRunning && detailValue.IsRunning() == 0 {
				return true
			}
			watches = append(watches, detailValue)
			return true
		})
		return true
	})
	return watches
}

// WatchCount returns the current count of watchers from the cache.
// If onlyRunning is true, the count will only include running watchers.
func WatchCount(onlyRunning bool) int {
	count := 0
	ResourceWatches.Range(func(k, v interface{}) bool {
		value, ok := v.(*sync.Map)
		if !ok {
			return false
		}
		value.Range(func(k, v interface{}) bool {
			detailValue, ok := v.(*WatchDetail)
			if !ok {
				return false
			}
			if onlyRunning && detailValue.IsRunning() == 0 {
				return true
			}
			count += 1
			return true
		})
		return true
	})
	return count
}

// WatchErrorHandlerFactory handles Reflector errors and ensures the Informer loop is shutdown when
// encountering an error.
func WatchErrorHandlerFactory(logger *zap.Logger, key string, stopCh chan<- struct{}) func(r *kcache.Reflector, err error) {
	return func(_ *kcache.Reflector, err error) {
		switch {
		case apierrors.IsResourceExpired(err) || apierrors.IsGone(err):
			logger.Error("watch closed",
				zap.String("name", key),
				zap.Error(err),
			)
			close(stopCh)
		case err == io.EOF:
			// watch closed normally
			close(stopCh)
		case err == io.ErrUnexpectedEOF:
			logger.Error("watch closed with unexpected EOF",
				zap.String("name", key),
				zap.Error(err),
			)
			close(stopCh)
		case apierrors.IsForbidden(err) || strings.Contains(err.Error(), "forbidden"):
			logger.Error("watch closed with forbidden",
				zap.String("name", key),
				zap.Error(err),
			)
			close(stopCh)
		default:
			close(stopCh)
			return
		}
	}
}
