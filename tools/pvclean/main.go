package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: os.Getenv("KUBECONFIG")},
		&clientcmd.ConfigOverrides{
			CurrentContext: "",
		}).RawConfig()
	if err != nil {
		log.Fatal(err)
	}
	for name := range config.Contexts {
		log.Println(name)
		cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: os.Getenv("KUBECONFIG")},
			&clientcmd.ConfigOverrides{
				CurrentContext: name,
			}).ClientConfig()
		if err != nil {
			log.Fatal(err)
		}
		clientset, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			log.Fatal(err)
		}
		names := map[string]bool{}
		pods, err := clientset.CoreV1().Pods("grafana-agent").List(context.Background(), metav1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		}
		for _, p := range pods.Items {
			names[p.GetName()] = true
		}
		pvcs, err := clientset.CoreV1().PersistentVolumeClaims("grafana-agent").List(context.Background(), metav1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		}
		for _, p := range pvcs.Items {
			n := strings.TrimPrefix(p.GetName(), "grafana-agent-helm-")
			if !names[n] {
				fmt.Printf("kubectl delete pvc grafana-agent-helm-%s -n grafana-agent --context %s\n", n, name)
			}
		}
	}
}
