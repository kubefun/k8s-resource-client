package cache

import (
	"go.uber.org/zap"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
)

type WatcherOption func(*Watcher)

func WithDynamicSharedInformerFactory(dsif dynamicinformer.DynamicSharedInformerFactory) WatcherOption {
	return func(w *Watcher) {
		w.informerFactory = dsif
	}
}

func WithDynamicClient(d dynamic.Interface) WatcherOption {
	return func(w *Watcher) {
		w.dclient = d
	}
}

func WithLogger(logger *zap.Logger) WatcherOption {
	return func(w *Watcher) {
		w.logger = logger
	}
}

func WithNamespace(namespace string) WatcherOption {
	return func(w *Watcher) {
		w.namespace = namespace
	}
}
