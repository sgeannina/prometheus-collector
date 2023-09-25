package cmd

import (
	"context"
	"os"
	"path"
	"strconv"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

var (
	settings = []string{
		"default-scrape-settings-enabled",
		"default-targets-metrics-keep-list",
		"default-targets-scrape-interval-settings",
	}
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

	println("Watch for changes in configmap...")
	for {
		watcher, err := clientSet.CoreV1().ConfigMaps(namespace).Watch(context.TODO(),
			metav1.SingleObject(metav1.ObjectMeta{Name: configmapName, Namespace: namespace}))
		if err != nil {
			panic("Unable to create watcher")
		}
		handleConfimapUpdate(watcher.ResultChan(), mutex)
		println("Retrying watch...")
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
				deleteSettingsFiles(settingsVolume)
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
	err := os.Remove(path.Join(volumePath, "inotifysettingscreated"))
	if err != nil {
		panic("Unable to delete inotifysettingscreated file. Error: " + err.Error())
	}

	if updatedConfigMap, ok := event.Object.(*corev1.ConfigMap); ok {
		println("Added configmap: " + updatedConfigMap.ObjectMeta.Name + " ok: " + strconv.FormatBool(ok))
		for _, settingKey := range settings {
			filePath := path.Join(volumePath, settingKey)
			err = os.WriteFile(filePath, []byte(updatedConfigMap.Data[settingKey]), fileMode)
			if err != nil {
				panic("Unable to create/update file: " + filePath + " Error: " + err.Error())
			}
		}
	}

	_, err = os.Create(path.Join(volumePath, "inotifysettingscreated"))
	if err != nil {
		panic("Unable to create inotifysettingscreated file. Error: " + err.Error())
	}
}

func deleteSettingsFiles(volumePath string) {
	for _, settingKey := range settings {
		filePath := path.Join(volumePath, settingKey)
		err := os.Remove(filePath)
		if err != nil {
			panic("Unable to delete file: " + filePath + " Error: " + err.Error())
		}
		println("Removed file: " + filePath)
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
		println("Error getting configmap: " + err.Error())
		return false
	}
	return true
}
