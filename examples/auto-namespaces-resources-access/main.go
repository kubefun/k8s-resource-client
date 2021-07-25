package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"k8s.io/apimachinery/pkg/labels"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/wwitzel3/k8s-resource-client/pkg/cache"
	r6eCache "github.com/wwitzel3/k8s-resource-client/pkg/cache"
	r6eClient "github.com/wwitzel3/k8s-resource-client/pkg/client"
)

func main() {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
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

	err = r6eClient.AutoDiscoverNamespaces(ctx, client)
	if err != nil {
		panic(err)
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

	// Update the access cache for the first namespaced resource and check if we can list/watch it.
	r6eClient.UpdateResourceAccess(ctx, client, nsResources[0], []string{""})
	fmt.Println(fmt.Sprintf("check list,watch access for %v: ", nsResources[0]), r6eCache.Access.AllowedAll("", nsResources[0], []string{"list", "watch"}))

	r6eClient.WatchAllResources(ctx, client, false, []string{""})

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(1)
	}()

	for {
		fmt.Println("active watcher count:", cache.WatchCount(true))
		v, ok := cache.ResourceWatches.Load("v1.Pod")
		if !ok {
			fmt.Println("not found")
		}
		watcher, ok := v.(*cache.WatchDetail)
		if !ok {
			fmt.Println("bad conversion")
		}

		fmt.Println("pod counts by namespaces")
		for _, ns := range cache.Namespaces {
			objs, err := watcher.List(labels.Everything())
			if err != nil {
				panic(err)
			}
			fmt.Printf("total pods in namespace %s: %d\n", ns, len(objs))
		}

		objs, err := watcher.List(labels.Everything())
		if err != nil {
			panic(err)
		}
		fmt.Printf("total pods in cluster: %d\n", len(objs))
		time.Sleep(5 * time.Second)
	}
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
