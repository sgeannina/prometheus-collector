package cmd

import (
	"context"
	"os"
	"path"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

const fileMode = 0640

// WatchForChanges watches a configmap for changes and updates the settings files
func WatchForChanges(clientSet *kubernetes.Clientset, namespace, configmapName, settingsVolume string, mutex *sync.Mutex) {
	if !configMapExists(clientSet, namespace, configmapName) {
		println("Configmap does not exist. Creating inotifysettingscreated file.")
		_, err := os.Create(path.Join(settingsVolume, "inotifysettingscreated"))
		if err != nil {
			panic("Unable to create inotifysettingscreated file. Error: " + err.Error())
		}
	}

	for {
		println("Watch for changes in configmap...")
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
				mutex.Lock()
				println("Added configmap")
				updateSettingsFiles(settingsVolume, event)
				mutex.Unlock()
			case watch.Modified:
				mutex.Lock()
				println("Updated configmap")
				updateSettingsFiles(settingsVolume, event)
				mutex.Unlock()
			case watch.Deleted:
				mutex.Lock()
				println("Deleted configmap")
				deleteSettingsFiles(settingsVolume, event)
				mutex.Unlock()
			default:
				// Do nothing
				println("Do nothing")
			}
		} else {
			// If eventChannel is closed, it means the server has closed the connection
			println("Channel closed. Server has closed the connection.")
			return
		}
	}
}

func updateSettingsFiles(volumePath string, event watch.Event) {
	removeFileIfExists(path.Join(volumePath, "inotifysettingscreated"))

	if updatedConfigMap, ok := event.Object.(*corev1.ConfigMap); ok {
		for settingKey, settingValue := range updatedConfigMap.Data {
			println("Creating/updating settings file: " + settingKey)
			filePath := path.Join(volumePath, settingKey)
			err := os.WriteFile(filePath, []byte(settingValue), fileMode)
			if err != nil {
				panic("Unable to create/update file: " + filePath + ". Error: " + err.Error())
			}
		}
	}

	_, err := os.Create(path.Join(volumePath, "inotifysettingscreated"))
	if err != nil {
		panic("Unable to create inotifysettingscreated file. Error: " + err.Error())
	}
}

func deleteSettingsFiles(volumePath string, event watch.Event) {
	removeFileIfExists(path.Join(volumePath, "inotifysettingscreated"))

	if updatedConfigMap, ok := event.Object.(*corev1.ConfigMap); ok {
		for settingKey := range updatedConfigMap.Data {
			println("Deleting settings file: " + settingKey)
			filePath := path.Join(volumePath, settingKey)
			err := os.Remove(filePath)
			if err != nil {
				panic("Unable to delete file: " + filePath + " Error: " + err.Error())
			}
		}
	}

	_, err := os.Create(path.Join(volumePath, "inotifysettingscreated"))
	if err != nil {
		panic("Unable to create inotifysettingscreated file. Error: " + err.Error())
	}
}

func configMapExists(clientset *kubernetes.Clientset, namespace, configMapName string) bool {
	_, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return false
	}
	if err != nil {
		// TODO: Fix error handling
		println("Error getting configmap: " + err.Error())
		return false
	}
	return true
}

func removeFileIfExists(filePath string) {
	if err := os.Remove(filePath); err != nil {
		// Check if the error is due to the file not existing, otherwise panic
		if !os.IsNotExist(err) {
			panic("Error removing file: " + filePath + ". Error: " + err.Error())
		}
	}
}
