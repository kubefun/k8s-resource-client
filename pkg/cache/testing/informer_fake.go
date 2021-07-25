package testing

import (
	"time"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

var _ dynamicinformer.DynamicSharedInformerFactory = (*FakeDynamicSharedInformerFactory)(nil)

type FakeDynamicSharedInformerFactory struct {
	GenericInformer *FakeGenericInformer
}

func NewFakeDynamicSharedInformerFactory() *FakeDynamicSharedInformerFactory {
	return &FakeDynamicSharedInformerFactory{
		GenericInformer: NewFakeGenericInformer(),
	}
}

func (d FakeDynamicSharedInformerFactory) Start(stopCh <-chan struct{}) {}
func (d FakeDynamicSharedInformerFactory) ForResource(gvr schema.GroupVersionResource) informers.GenericInformer {
	return d.GenericInformer
}
func (d FakeDynamicSharedInformerFactory) WaitForCacheSync(stopCh <-chan struct{}) map[schema.GroupVersionResource]bool {
	return map[schema.GroupVersionResource]bool{}
}

type FakeGenericInformer struct {
	SharedIndexInformer *FakeSharedIndexInformer
	GenericLister       *FakeGenericLister
}

func NewFakeGenericInformer() *FakeGenericInformer {
	return &FakeGenericInformer{
		SharedIndexInformer: NewFakeSharedIndexInformer(),
		GenericLister:       NewFakeGenericLister(),
	}
}

func (s FakeGenericInformer) Informer() cache.SharedIndexInformer {
	return s.SharedIndexInformer
}

func (s FakeGenericInformer) Lister() cache.GenericLister { return s.GenericLister }

type FakeSharedIndexInformer struct {
	Handlers []cache.ResourceEventHandler
}

func NewFakeSharedIndexInformer() *FakeSharedIndexInformer {
	return &FakeSharedIndexInformer{
		Handlers: []cache.ResourceEventHandler{},
	}
}

func (s *FakeSharedIndexInformer) AddEventHandler(handler cache.ResourceEventHandler) {
	s.Handlers = append(s.Handlers, handler)
}

func (s FakeSharedIndexInformer) AddEventHandlerWithResyncPeriod(handler cache.ResourceEventHandler, resyncPeriod time.Duration) {
}
func (s FakeSharedIndexInformer) GetStore() cache.Store           { return nil }
func (s FakeSharedIndexInformer) GetController() cache.Controller { return nil }
func (s FakeSharedIndexInformer) Run(stopCh <-chan struct{})      {}
func (s FakeSharedIndexInformer) HasSynced() bool                 { return true }
func (s FakeSharedIndexInformer) LastSyncResourceVersion() string { return "" }
func (s FakeSharedIndexInformer) SetWatchErrorHandler(handler cache.WatchErrorHandler) error {
	return nil
}
func (s FakeSharedIndexInformer) AddIndexers(indexers cache.Indexers) error { return nil }
func (s FakeSharedIndexInformer) GetIndexer() cache.Indexer                 { return nil }

type FakeGenericLister struct {
	ListErr error
	Objects []runtime.Object

	GetErr error
	Object runtime.Object

	NamespaceLister *FakeNamespaceLister
}

func NewFakeGenericLister() *FakeGenericLister {
	return &FakeGenericLister{
		NamespaceLister: NewFakeNamespaceLister(),
	}
}

// List will return all objects across namespaces
func (f FakeGenericLister) List(selector labels.Selector) (ret []runtime.Object, err error) {
	if f.ListErr != nil {
		return nil, f.ListErr
	}
	return f.Objects, nil
}

// Get will attempt to retrieve assuming that name==key
func (f FakeGenericLister) Get(name string) (runtime.Object, error) {
	if f.GetErr != nil {
		return nil, f.GetErr
	}
	return f.Object, nil
}

// ByNamespace will give you a GenericNamespaceLister for one namespace
func (f *FakeGenericLister) ByNamespace(namespace string) cache.GenericNamespaceLister {
	f.NamespaceLister.Namespace = namespace
	return f.NamespaceLister
}

// GenericNamespaceLister is a lister skin on a generic Indexer
type FakeNamespaceLister struct {
	Namespace string

	ListErr error
	Objects []runtime.Object

	GetErr error
	Object runtime.Object
}

func NewFakeNamespaceLister() *FakeNamespaceLister {
	return &FakeNamespaceLister{}
}

// List will return all objects in this namespace
func (f FakeNamespaceLister) List(selector labels.Selector) (ret []runtime.Object, err error) {
	if f.ListErr != nil {
		return nil, f.ListErr
	}
	return f.Objects, nil
}

// Get will attempt to retrieve by namespace and name
func (f FakeNamespaceLister) Get(name string) (runtime.Object, error) {
	if f.GetErr != nil {
		return nil, f.GetErr
	}
	return f.Object, nil
}
