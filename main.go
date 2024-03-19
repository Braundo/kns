package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/apimachinery/pkg/api/resource"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	for _, ns := range namespaces.Items {
		fmt.Printf("____________________________________\n")
		fmt.Printf("| Namespace: %-22s |\n", ns.Name)
		fmt.Printf("------------------------------------\n")
		pods, err := clientset.CoreV1().Pods(ns.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("| Error fetching pods: %-12v |\n", err)
			continue
		}

		var totalCPURequests, totalCPULimits, totalMemoryRequests, totalMemoryLimits resource.Quantity
		for _, pod := range pods.Items {
			for _, container := range pod.Spec.Containers {
				if cpuRequest, ok := container.Resources.Requests[corev1.ResourceCPU]; ok {
					totalCPURequests.Add(cpuRequest)
				}
				if cpuLimit, ok := container.Resources.Limits[corev1.ResourceCPU]; ok {
					totalCPULimits.Add(cpuLimit)
				}
				if memoryRequest, ok := container.Resources.Requests[corev1.ResourceMemory]; ok {
					totalMemoryRequests.Add(memoryRequest)
				}
				if memoryLimit, ok := container.Resources.Limits[corev1.ResourceMemory]; ok {
					totalMemoryLimits.Add(memoryLimit)
				}
			}
		}

		fmt.Printf("| Total CPU Requests: %-14s |\n", totalCPURequests.String())
		fmt.Printf("| Total CPU Limits: %-15s |\n", totalCPULimits.String())
		fmt.Printf("| Total Memory Requests: %-10s |\n", totalMemoryRequests.String())
		fmt.Printf("| Total Memory Limits: %-11s |\n", totalMemoryLimits.String())
		fmt.Printf("____________________________________\n\n")
	}
}
