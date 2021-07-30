package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/gorilla/websocket"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"

	r6eCache "github.com/wwitzel3/k8s-resource-client/pkg/cache"
	r6eClient "github.com/wwitzel3/k8s-resource-client/pkg/client"
	"github.com/wwitzel3/k8s-resource-client/pkg/resource"
)

var currentNamespace = metav1.NamespaceAll

func main() {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	autoDiscoverNamespaces := flag.Bool("discover-namespaces", false, "auto discover namespaces")
	namespace := flag.String("namespace", metav1.NamespaceAll, "namespace to use, defaults to all")

	flag.Parse()

	currentNamespace = *namespace
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	config.QPS = 500
	config.Burst = 1000

	ctx := context.Background()

	client, err := r6eClient.NewClient(ctx, r6eClient.WithRESTConfig(config))
	if err != nil {
		panic(err)
	}

	if *autoDiscoverNamespaces {
		err = r6eClient.AutoDiscoverNamespaces(ctx, client)
		if err != nil {
			panic(err)
		}
	}

	if err := r6eClient.AutoDiscoverResources(ctx, client); err != nil {
		panic(err)
	}

	nsResources := r6eCache.Resources.Get(r6eCache.NamespacedResources)
	fmt.Printf("namespace resource count: %d\n", len(nsResources))

	cResources := r6eCache.Resources.Get(r6eCache.ClusterScopedResources)
	fmt.Printf("cluster resource count: %d\n", len(cResources))

	if err := r6eClient.AutoDiscoverAccess(ctx, client, currentNamespace, nsResources...); err != nil {
		panic(err)
	}

	namespaces := []string{currentNamespace}
	r6eClient.WatchAllResources(ctx, client, true, namespaces)

	http.HandleFunc("/ws", handler)
	http.HandleFunc("/stats", stats)
	http.HandleFunc("/resources", resources)
	if err := http.ListenAndServe("127.0.0.1:1234", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

type Stats struct {
	Total   int `json:"total"`
	Running int `json:"running"`
	Stopped int `json:"stopped"`
}

func stats(writer http.ResponseWriter, request *http.Request) {
	total := r6eCache.WatchCount(false)
	running := r6eCache.WatchCount(true)

	stats := Stats{
		Total:   total,
		Running: running,
		Stopped: total - running,
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	encoder := json.NewEncoder(writer)
	err := encoder.Encode(stats)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
	}
}

type Resource struct {
	Group   string `json:"group"`
	Kind    string `json:"kind"`
	Version string `json:"version"`
	List    bool   `json:"list"`
	Watch   bool   `json:"watch"`
}

func resources(writer http.ResponseWriter, request *http.Request) {
	resourceScope := r6eCache.ClusterScopedResources
	scope := request.URL.Query().Get("scope")
	if scope == "namespace" {
		resourceScope = r6eCache.NamespacedResources
	}

	resources := []Resource{}
	for _, r := range r6eCache.Resources.Get(resourceScope) {
		canList := r6eCache.Access.Allowed(currentNamespace, r, "list")
		canWatch := r6eCache.Access.Allowed(currentNamespace, r, "watch")

		resources = append(resources, Resource{
			List:    canList,
			Watch:   canWatch,
			Group:   r.GroupVersionKind.Group,
			Version: r.GroupVersionKind.Version,
			Kind:    r.GroupVersionKind.Kind,
		})
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	encoder := json.NewEncoder(writer)
	err := encoder.Encode(resources)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
	}
}

type Watcher struct {
	Namespace           string `json:"namespace"`
	Resource            string `json:"resource"`
	IsRunning           int    `json:"isRunning"`
	Queue               bool   `json:"queue"`
	HandledEventCount   int    `json:"handledEventCount"`
	UnhandledEventCount int    `json:"unhandledEventCount"`
	LastEvent           string `json:"lastEvent"`
}

var (
	pongWait = 60 * time.Second
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(*http.Request) bool { return true },
	}
)

type Client struct {
	conn    *websocket.Conn
	eventCh chan interface{}
	stopCh  chan struct{}
	watcher r6eCache.ResourceLister
}

func handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	eventCh := make(chan interface{})
	stopCh := make(chan struct{})

	eventWatcher, err := r6eCache.WatchForResource(podResource, currentNamespace)
	if err != nil {
		println(err.Error())
		return
	}

	eventWatcher.Drain(eventCh, stopCh)

	client := &Client{conn: conn, eventCh: eventCh, stopCh: stopCh, watcher: eventWatcher}
	go client.writePump()
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
		close(c.stopCh)
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	c.sendUpdate(nil)
	for e := range c.eventCh {
		c.sendUpdate(e)
	}
}

func (c *Client) sendUpdate(e interface{}) {
	watchers := []Watcher{}
	for _, w := range r6eCache.WatchList(false) {
		watchers = append(watchers, Watcher{
			Namespace:           w.Namespace(),
			Resource:            w.Key(),
			IsRunning:           w.IsRunning(),
			HandledEventCount:   w.HandledEventCount(),
			UnhandledEventCount: w.UnhandledEventCount(),
			Queue:               true,
			LastEvent:           fmt.Sprintf("%+v", e),
		})
	}

	sort.Slice(watchers, func(i, j int) bool {
		if watchers[i].Namespace < watchers[j].Namespace {
			return true
		}
		if watchers[i].Namespace > watchers[j].Namespace {
			return false
		}
		return watchers[i].Resource < watchers[j].Resource
	})

	w, err := c.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		c.watcher.Stop()
		return
	}

	err = json.NewEncoder(w).Encode(watchers)
	if err != nil {
		c.watcher.Stop()
		return
	}
}

var podResource = resource.Resource{
	GroupVersionKind: schema.GroupVersionKind{
		Version: "v1",
		Kind:    "Pod",
	},
}
