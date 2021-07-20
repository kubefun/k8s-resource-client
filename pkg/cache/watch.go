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

type WatchDetails struct {
	Informer    kcache.SharedInformer
	Lister      kcache.GenericLister
	StopCh      chan struct{}
	Resource    resource.Resource
	Key         string
	QueueEvents bool
	Queue       *workqueue.Type
	Logger      *zap.Logger
}

func (w *WatchDetails) IsRunning() bool {
	select {
	case <-w.StopCh:
		return false
	default:
		return true
	}
}

func (w *WatchDetails) Stop() {
	close(w.StopCh)
}

func (w *WatchDetails) NopDrain() {
	for {
		select {
		case <-w.StopCh:
			return
		default:
			_, shutdown := w.Queue.Get()
			w.Logger.Info("processing queue")
			if shutdown {
				return
			}
		}
	}
}

type Watcher struct {
	informerFactory dynamicinformer.DynamicSharedInformerFactory
	logger          *zap.Logger
}

func NewWatcher(ctx context.Context, logger *zap.Logger, client dynamic.Interface) *Watcher {
	return &Watcher{
		informerFactory: dynamicinformer.NewDynamicSharedInformerFactory(client, DefaultResyncDuration),
		logger:          logger,
	}
}

func (w *Watcher) Watch(ctx context.Context, res resource.Resource, queueEvents bool) *WatchDetails {
	resourceInformer := w.informerFactory.ForResource(res.GroupVersionResource())

	lister := resourceInformer.Lister()
	informer := resourceInformer.Informer()

	details := &WatchDetails{
		Key:         res.Key(),
		Resource:    res,
		Informer:    informer,
		Lister:      lister,
		QueueEvents: queueEvents,
		Queue:       workqueue.NewNamed(res.Key()),
		StopCh:      make(chan struct{}),
		Logger:      w.logger,
	}

	// boardcast function that will publish changes to a channel for clients
	if details.QueueEvents {
		informer.AddEventHandler(kcache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				w.logger.Info("watch add",
					zap.String("obj", fmt.Sprintf("%v", obj)),
				)
				details.Queue.Add(obj)
			},
			DeleteFunc: func(obj interface{}) {
				w.logger.Info("watch delete",
					zap.String("obj", fmt.Sprintf("%v", obj)),
				)
				details.Queue.Done(obj)
			},
			UpdateFunc: func(new, old interface{}) {
				w.logger.Info("watch update",
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

func WatcherList(onlyRunning bool) []*WatchDetails {
	watches := []*WatchDetails{}
	Watches.Range(func(k, v interface{}) bool {
		value, ok := v.(*WatchDetails)
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

func WatchCount(onlyRunning bool) int {
	count := 0
	Watches.Range(func(k, v interface{}) bool {
		value, ok := v.(*WatchDetails)
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

func WatchErrorHandlerFactory(logger *zap.Logger, key string, stopCh chan<- struct{}) func(r *kcache.Reflector, err error) {
	return func(r *kcache.Reflector, err error) {
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
