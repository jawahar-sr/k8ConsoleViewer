package main

import (
	"flag"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

var k8client *kubernetes.Clientset

func NewK8ClientForContext(context string) error {
	var kubeconfig *string
	configPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return errors.New("No config found in ~/.kube")
	}
	kubeconfig = flag.String("kubeconfig", configPath, "(optional) absolute path to the kubeconfig file")

	flag.Parse()

	config, err := buildConfigFromFlags(context, *kubeconfig)
	if err != nil {
		return err
	}

	k8client, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	return nil
}

func getPods(namespace string) Namespace {
	ns := Namespace{Name: namespace}

	pods, err := k8client.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		ns.Error = err
		return ns
	}

	ns.Name = namespace
	ns.Pods = make([]Pod, len(pods.Items))

	for i, p := range pods.Items {
		pod := Pod{
			Name:      p.Name,
			Status:    string(p.Status.Phase),
			Total:     len(p.Status.ContainerStatuses),
			Ready:     countReady(&pods.Items[i]),
			Restarts:  countRestarts(&pods.Items[i]),
			Age:       timeToAge(p.Status.StartTime.Time, time.Now()),
			Namespace: &ns,
		}

		containers := make([]Container, len(p.Spec.Containers))
		for ic, c := range p.Spec.Containers {
			cont := Container{
				Name:  c.Name,
				Image: c.Image,
				Ready: isReady(&pods.Items[i], c.Name),
				Pod:   &pod,
			}
			containers[ic] = cont
		}
		pod.Containers = containers
		ns.Pods[i] = pod
	}
	return ns
}

func isReady(p *v1.Pod, name string) bool {

	for _, v := range p.Status.ContainerStatuses {
		if name == v.Name {
			return v.Ready
		}
	}

	log.Println("Container status not found for container:", name)
	return false
}

func countReady(p *v1.Pod) int {
	count := 0
	for _, s := range p.Status.ContainerStatuses {
		if s.Ready {
			count++
		}
	}
	return count
}

func countRestarts(p *v1.Pod) int {
	count := 0
	for _, s := range p.Status.ContainerStatuses {
		count += int(s.RestartCount)
	}
	return count
}

func updateNamespaces(g *Group) ([]Namespace, error) {
	var wg sync.WaitGroup

	nsInfoCh := make(chan Namespace, len(g.Namespaces))
	wg.Add(len(g.Namespaces))

	// Get pod info in parallel
	for i := range g.Namespaces {
		go func(nsInfoCh chan<- Namespace, ctxName, nsName string) {
			nsInfoCh <- getPods(nsName)
		}(nsInfoCh, g.Context, g.Namespaces[i])
	}

	// Wait for everything to finish and collect pod info into one slice
	go func() {
		wg.Wait()
		close(nsInfoCh)
	}()

	nsInfos := make([]Namespace, 0)
	for nsInfo := range nsInfoCh {
		nsInfo.sortPods()
		nsInfos = append(nsInfos, nsInfo)
		wg.Done()
	}
	sort.Slice(nsInfos, func(i, j int) bool { return nsInfos[i].Name < nsInfos[j].Name })
	return nsInfos, nil
}

func (ns *Namespace) sortPods() {
	sort.Slice(ns.Pods, func(i, j int) bool {
		return ns.Pods[i].Name < ns.Pods[j].Name
	})
}

func buildConfigFromFlags(context, kubeconfigPath string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		}).ClientConfig()
}
