package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/websocket"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"

	r6eCache "github.com/wwitzel3/k8s-resource-client/pkg/cache"
	r6eClient "github.com/wwitzel3/k8s-resource-client/pkg/client"
	"github.com/wwitzel3/k8s-resource-client/pkg/logging"
	"github.com/wwitzel3/k8s-resource-client/pkg/resource"
)

type Fields struct {
	Fields []Field `json:"fields"`
}

type Field struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Action string `json:"action"`
}

var podRes = resource.Resource{
	APIResource: metav1.APIResource{
		Name:         "pods",
		SingularName: "pod",
		Verbs:        []string{"list", "watch"},
	},
	GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
}

func Echo(ws *websocket.Conn) {
	eventCh := make(chan interface{})
	stopCh := make(chan struct{})

	pods, err := r6eCache.WatchForResource(podRes, "default")
	if err != nil {
		logging.Logger.Warn("pod watcher", zap.Error(err))
	} else {
		pods.Drain(eventCh, stopCh)
	}
	deployments, err := r6eCache.WatchForResource(resource.Resource{GroupVersionKind: schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}})
	if err != nil {
		logging.Logger.Warn("deployment watcher", zap.Error(err))
	} else {
		deployments.Drain(eventCh, stopCh)
	}
	replicaSets, err := r6eCache.WatchForResource(resource.Resource{GroupVersionKind: schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "ReplicaSet"}})
	if err != nil {
		logging.Logger.Warn("replicaset watcher", zap.Error(err))
	} else {
		replicaSets.Drain(eventCh, stopCh)
	}

	// Send the inital update
	sendUpdate(ws, pods, deployments, replicaSets)

	// Send updates when one of there resources has a Create, Update, Delete event
	for range eventCh {
		sendUpdate(ws, pods, deployments, replicaSets)
	}
}

func sendUpdate(ws *websocket.Conn, pods, deployments, replicasets r6eCache.ResourceLister) {
	fields := Fields{Fields: []Field{}}

	ts := Field{Key: "timestamp", Value: time.Now().UTC().String(), Action: ""}
	fields.Fields = append(fields.Fields, ts)

	f := Field{Key: "watcher count", Value: fmt.Sprintf("%d", r6eCache.WatchCount(true)), Action: ""}
	fields.Fields = append(fields.Fields, f)

	if pods != nil {
		podObjs, _ := pods.List(labels.Everything())
		fields.Fields = append(fields.Fields, Field{Key: "pod_count", Value: fmt.Sprintf("%d", len(podObjs))})
	}

	if deployments != nil {
		depObjs, _ := deployments.List(labels.Everything())
		fields.Fields = append(fields.Fields, Field{Key: "deployment_count", Value: fmt.Sprintf("%d", len(depObjs))})
	}

	if replicasets != nil {
		rsOjs, _ := replicasets.List(labels.Everything())
		fields.Fields = append(fields.Fields, Field{Key: "replicaset_count", Value: fmt.Sprintf("%d", len(rsOjs))})
	}

	s, _ := json.CaseSensitiveJSONIterator().MarshalToString(fields)
	websocket.Message.Send(ws, s)
}

func main() {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	autoDiscoverNamespaces := flag.Bool("discover-namespaces", false, "auto discover namespaces")

	flag.Parse()

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

	nsResources := r6eCache.Resources.Get("namespace")
	fmt.Printf("namespace resource count: %d\n", len(nsResources))

	cResources := r6eCache.Resources.Get("cluster")
	fmt.Printf("cluster resource count: %d\n", len(cResources))

	// No resources provided this will init an empty access cache, all checks will be false
	if err := r6eClient.AutoDiscoverAccess(ctx, client, ""); err != nil {
		panic(err)
	}

	namespaces := []string{"default", metav1.NamespaceAll}
	// if err := r6eClient.UpdateResourceAccess(ctx, client, podRes, namespaces); err != nil {
	// 	panic(err)
	// }
	r6eClient.WatchAllResources(ctx, client, true, namespaces)
	for _, ns := range namespaces {
		println(ns)
		if r6eCache.Access.AllowedAll(ns, podRes, []string{"watch", "list"}) {
			println("watching", podRes.Key(), ns)
			r6eClient.WatchResource(ctx, client, podRes, true, []string{ns})
		}
	}

	http.Handle("/", websocket.Handler(Echo))
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
