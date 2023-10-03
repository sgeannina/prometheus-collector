package cmd

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"os"
	"path"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

const FileMode = 0640
const WatcherNotificationFile = "inotifysettingscreated"

// WatchForChanges watches a configmap for changes and updates the settings files
func WatchForChanges(clientSet *kubernetes.Clientset, logger *zap.Logger, namespace, configmapName, settingsVolume string) error {
	if exists, err := configMapExists(clientSet, namespace, configmapName); !exists {
		if err != nil {
			panic("Unable to read configmap. Error: " + err.Error())
		}

		logger.Info("Configmap does not exist. Creating inotifysettingscreated file.")
		_, err := os.Create(path.Join(settingsVolume, WatcherNotificationFile))
		if err != nil {
			return fmt.Errorf("unable to create inotifysettingscreated file. Error: %w", err)
		}
	}

	mutex = &sync.Mutex{}
	for {
		println("Watch for changes in configmap...")
		watcher, err := clientSet.CoreV1().ConfigMaps(namespace).Watch(context.TODO(),
			metav1.SingleObject(metav1.ObjectMeta{Name: configmapName, Namespace: namespace}))
		if err != nil {
			panic("Unable to create watcher")
		}

		handleConfimapUpdate(watcher.ResultChan(), logger, mutex)
	}
}

func handleConfimapUpdate(eventChannel <-chan watch.Event, logger *zap.Logger, mutex *sync.Mutex) {
	for {
		event, open := <-eventChannel
		if open {
			mutex.Lock()
			switch event.Type {
			case watch.Added:
				logger.Info("Added configmap")
				updateSettingsFiles(settingsVolume, event)
			case watch.Modified:
				logger.Info("Updated configmap")
				updateSettingsFiles(settingsVolume, event)
			case watch.Deleted:
				logger.Info("Deleted configmap")
				deleteSettingsFiles(settingsVolume, event)
			default:
				// Do nothing
				logger.Error(fmt.Sprintf("Unsupported event type '%s'", event.Type))
			}
			mutex.Unlock()
		} else {
			// If eventChannel is closed, it means the server has closed the connection
			logger.Info("Channel closed. Server has closed the connection.")
			return
		}
	}
}

func updateSettingsFiles(volumePath string, event watch.Event) {
	removeFileIfExists(path.Join(volumePath, WatcherNotificationFile))

	if updatedConfigMap, ok := event.Object.(*corev1.ConfigMap); ok {
		for settingKey, settingValue := range updatedConfigMap.Data {
			println("Creating/updating settings file: " + settingKey)
			filePath := path.Join(volumePath, settingKey)
			err := os.WriteFile(filePath, []byte(settingValue), FileMode)
			if err != nil {
				panic("Unable to create/update file: " + filePath + ". Error: " + err.Error())
			}
		}
	}

	_, err := os.Create(path.Join(volumePath, WatcherNotificationFile))
	if err != nil {
		panic("Unable to create inotifysettingscreated file. Error: " + err.Error())
	}
}

func deleteSettingsFiles(volumePath string, event watch.Event) {
	removeFileIfExists(path.Join(volumePath, WatcherNotificationFile))

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

	_, err := os.Create(path.Join(volumePath, WatcherNotificationFile))
	if err != nil {
		panic("Unable to create inotifysettingscreated file. Error: " + err.Error())
	}
}

func configMapExists(clientset *kubernetes.Clientset, namespace, configMapName string) (bool, error) {
	_, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func removeFileIfExists(filePath string) {
	if err := os.Remove(filePath); err != nil {
		// Check if the error is due to the file not existing, otherwise panic
		if !os.IsNotExist(err) {
			panic("Error removing file: " + filePath + ". Error: " + err.Error())
		}
	}
}
