package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"

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

	if err := r6eClient.AutoDiscoverResources(ctx, client, r6eCache.Namespaces...); err != nil {
		panic(err)
	}

	for _, namespace := range r6eCache.Namespaces {
		resources := r6eCache.Resources.GetResources(namespace)
		fmt.Printf("namespace: %s, resource_count: %d\n", namespace, len(resources))
	}
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
