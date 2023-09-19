package cmd

import (
	"context"
	"sync"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

func WatchForChanges(clientSet *kubernetes.Clientset, namespace string, configmapName string, mutex *sync.Mutex) {
	for {
		watcher, err := clientSet.CoreV1().ConfigMaps(namespace).Watch(context.TODO(),
			metav1.SingleObject(metav1.ObjectMeta{Name: configmapName, Namespace: namespace}))
		if err != nil {
			panic("Unable to create watcher")
		}
		handleConfimapUpdate(watcher.ResultChan(), mutex)
	}
}

func handleConfimapUpdate(eventChannel <-chan watch.Event, mutex *sync.Mutex) {
	for {
		event, open := <-eventChannel
		if open {
			switch event.Type {
			case watch.Added:
				fallthrough
			case watch.Modified:
				mutex.Lock()
				println("Updated configmap")
				// Update files
				if updatedMap, ok := event.Object.(*corev1.ConfigMap); ok {
					if settingKey, ok := updatedMap.Data["current.target"]; ok {
						if settingValue, ok := updatedMap.Data[settingKey]; ok {
							println(settingValue)
						}
					}
				}
				mutex.Unlock()
			case watch.Deleted:
				mutex.Lock()
				println("Deleted configmap")
				mutex.Unlock()
			default:
				// Do nothing
				println("Do nothing")
			}
		} else {
			// If eventChannel is closed, it means the server has closed the connection
			println("Channel closed")
			return
		}
	}
}

func UpdateSettingsFiles(volumePath string, settingKey string, settingValue string) {

}
