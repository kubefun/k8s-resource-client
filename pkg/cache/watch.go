package cache

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/wwitzel3/k8s-resource-client/pkg/resource"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	kcache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

var (
	DefaultResyncDuration = time.Second * 180
	Watches               = &sync.Map{}
)

// WatchDetail holds the details of an Informer and Lister for a specific resource.
// An optionally configured event queue.
type WatchDetail struct {
	Informer kcache.SharedInformer
	Lister   kcache.GenericLister
	StopCh   chan struct{}
	Resource resource.Resource
	Key      string
	Logger   *zap.Logger

	queueEvents bool
	Queue       *workqueue.Type
}

// IsRunning returns true if the Informer loop for the WatchDetail is running.
func (w *WatchDetail) IsRunning() bool {
	select {
	case <-w.StopCh:
		return false
	default:
		return true
	}
}

// Stop closes the StopCh shutting down the Drain and Informer loops.
func (w *WatchDetail) Stop() {
	close(w.StopCh)
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

// Watcher holds referenecs to the Kubernetes types and a logger.
// Use NewWatcher to create instances of Watcher.
type Watcher struct {
	dclient         dynamic.Interface
	informerFactory dynamicinformer.DynamicSharedInformerFactory
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

	if w.informerFactory == nil {
		w.informerFactory = dynamicinformer.NewDynamicSharedInformerFactory(w.dclient, DefaultResyncDuration)
	}

	return w, nil
}

// WatcherStop stops all running watchers.
func WatcherStop() {
	Watches.Range(func(k, v interface{}) bool {
		value, ok := v.(*WatchDetail)
		if !ok {
			return true
		}
		if value.IsRunning() {
			close(value.StopCh)
		}
		return true
	})
}

// Watch creates a new WatchDetail and starts the watch loop for the given Resource
// If queueEvents is true, all events for the resource will be added to the WatcheDetail.Queue
// To handle the events use WatchDetail.Drain
func (w *Watcher) Watch(ctx context.Context, res resource.Resource, queueEvents bool) *WatchDetail {
	resourceInformer := w.informerFactory.ForResource(res.GroupVersionResource())

	lister := resourceInformer.Lister()
	informer := resourceInformer.Informer()

	details := &WatchDetail{
		Key:         res.Key(),
		Resource:    res,
		Informer:    informer,
		Lister:      lister,
		queueEvents: queueEvents,
		Queue:       workqueue.NewNamed(res.Key()),
		StopCh:      make(chan struct{}),
		Logger:      w.logger,
	}

	// boardcast function that will publish changes to a channel for clients
	if details.queueEvents {
		informer.AddEventHandler(kcache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				w.logger.Debug("watch add",
					zap.String("obj", fmt.Sprintf("%v", obj)),
				)
				details.Queue.Add(obj)
			},
			DeleteFunc: func(obj interface{}) {
				w.logger.Debug("watch delete",
					zap.String("obj", fmt.Sprintf("%v", obj)),
				)
				details.Queue.Done(obj)
			},
			UpdateFunc: func(new, old interface{}) {
				w.logger.Debug("watch update",
					zap.String("obj", fmt.Sprintf("%v", new)),
				)
				details.Queue.Add(new)
			},
		})
	}

	informer.SetWatchErrorHandler(WatchErrorHandlerFactory(w.logger, details.Key, details.StopCh))

	go func() {
		w.logger.Debug("starting informer",
			zap.String("key", details.Key),
		)
		informer.Run(details.StopCh)
	}()

	Watches.Store(details.Key, details)
	return details
}

// WatchForResource returns a WatchDetail for the given Resource.
func WatchForResource(r resource.Resource) (*WatchDetail, bool) {
	v, ok := Watches.Load(r.Key())
	if !ok {
		return nil, ok
	}
	w, wok := v.(*WatchDetail)
	return w, wok
}

// WatchList returns the current count of watchers from the cache.
// If onlyRunning is true, the count will only include running watchers.
func WatchList(onlyRunning bool) []*WatchDetail {
	watches := []*WatchDetail{}
	Watches.Range(func(k, v interface{}) bool {
		value, ok := v.(*WatchDetail)
		if !ok {
			return false
		}
		if onlyRunning && !value.IsRunning() {
			return true
		}
		watches = append(watches, value)
		return true
	})
	return watches
}

// WatchCount returns the current count of watchers from the cache.
// If onlyRunning is true, the count will only include running watchers.
func WatchCount(onlyRunning bool) int {
	count := 0
	Watches.Range(func(k, v interface{}) bool {
		value, ok := v.(*WatchDetail)
		if !ok {
			return false
		}
		if onlyRunning && !value.IsRunning() {
			return true
		}
		count += 1
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
		default:
			return
		}
	}
}
