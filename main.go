package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"github.com/fatih/color"

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

	const padding = 2 // Define padding for columns
	maxNameLength := 0
	maxResourceLength := 1 // Set a maximum expected length for resource quantities

	// Find the longest namespace name
	for _, ns := range namespaces.Items {
		if len(ns.Name) > maxNameLength {
			maxNameLength = len(ns.Name)
		}
	}

	nameColumnWidth := maxNameLength + padding
	resourceColumnWidth := maxResourceLength + padding

	colorBlue := color.New(color.FgBlue).SprintFunc()
	fmt.Println("Namespace usage:")

	for _, ns := range namespaces.Items {
		// Colorize the "Namespace:" label and the namespace name
		namespaceLabel := "Namespace:"
		coloredName := colorBlue(ns.Name)
		nameColumn := fmt.Sprintf("%-"+strconv.Itoa(nameColumnWidth)+"s", ns.Name)
		fmt.Printf("%s %s\n", namespaceLabel, coloredName)
		fmt.Println(strings.Repeat(" ", len(nameColumn)+resourceColumnWidth*4+3))
		pods, err := clientset.CoreV1().Pods(ns.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("| Error fetching pods: %v |\n", err)
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

		// Prepare the resource columns with fixed width
		cpuReqColumn := fmt.Sprintf("%-"+strconv.Itoa(resourceColumnWidth)+"s", totalCPURequests.String())
		cpuLimColumn := fmt.Sprintf("%-"+strconv.Itoa(resourceColumnWidth)+"s", totalCPULimits.String())
		memReqColumn := fmt.Sprintf("%-"+strconv.Itoa(resourceColumnWidth)+"s", totalMemoryRequests.String())
		memLimColumn := fmt.Sprintf("%-"+strconv.Itoa(resourceColumnWidth)+"s", totalMemoryLimits.String())

		// Print the resource usage
		fmt.Printf(" CPU Requests: %s\n CPU Limits: %s\n Memory Requests: %s\n Memory Limits: %s\n", cpuReqColumn, cpuLimColumn, memReqColumn, memLimColumn)
		fmt.Println()
		fmt.Println(strings.Repeat("▄", len(nameColumn)+resourceColumnWidth*4+3))
		fmt.Println()
		fmt.Println()
	}
}
