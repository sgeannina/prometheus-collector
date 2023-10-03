package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"sync"

	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

const (
	fileMode                = 0640
	watcherNotificationFile = "inotifysettingscreated"
)

// WatchForChanges watches a configmap for changes and updates the settings files
func WatchForChanges(clientSet *kubernetes.Clientset, logger *zap.Logger, namespace, configmapName, settingsVolume string) error {
	if exists, err := configMapExists(clientSet, namespace, configmapName); !exists {
		if err != nil {
			return fmt.Errorf("unable to read configmap %s. Error: %w", configmapName, err)
		}

		logger.Info("Configmap does not exist. Creating inotifysettingscreated file.")
		_, err := os.Create(path.Join(settingsVolume, watcherNotificationFile))
		if err != nil {
			return fmt.Errorf("unable to create inotifysettingscreated file. Error: %w", err)
		}
	}

	mutex = &sync.Mutex{}
	for {
		logger.Info("Watch for changes in configmap...")
		watcher, err := clientSet.CoreV1().ConfigMaps(namespace).Watch(context.TODO(),
			metav1.SingleObject(metav1.ObjectMeta{Name: configmapName, Namespace: namespace}))
		if err != nil {
			logger.Error(fmt.Sprintf("Unable to create watcher. Error: %s", err.Error()))
			return err
		}

		err = handleConfigmapUpdate(watcher.ResultChan(), logger, mutex)
		if err != nil {
			logger.Error(fmt.Sprintf("Error while processing configmap update. Error: %s", err.Error()))
			return err
		}
	}
}

func handleConfigmapUpdate(eventChannel <-chan watch.Event, logger *zap.Logger, mutex *sync.Mutex) error {
	for {
		event, open := <-eventChannel
		if open {
			mutex.Lock()
			switch event.Type {
			case watch.Added:
				logger.Info("Adding configmap")
				err := updateSettingsFiles(settingsVolume, event, logger)
				if err != nil {
					logger.Error(fmt.Sprintf("Unable to create settings files. Error: %s", err.Error()))
					return err
				}
			case watch.Modified:
				logger.Info("Updating configmap")
				err := updateSettingsFiles(settingsVolume, event, logger)
				if err != nil {
					logger.Error(fmt.Sprintf("Unable to update settings files. Error: %s", err.Error()))
					return err
				}
			case watch.Deleted:
				logger.Info("Deleting configmap")
				err := deleteSettingsFiles(settingsVolume, event, logger)
				if err != nil {
					logger.Error(fmt.Sprintf("Unable to delete settings files. Error: %s", err.Error()))
					return err
				}
			default:
				// Do nothing
				logger.Error(fmt.Sprintf("Unsupported event type '%s'", event.Type))
			}
			mutex.Unlock()
		} else {
			// If eventChannel is closed, it means the server has closed the connection
			logger.Info("Channel closed. Server has closed the connection.")
			return nil
		}
	}
}

func updateSettingsFiles(volumePath string, event watch.Event, logger *zap.Logger) error {
	err := removeFileIfExists(path.Join(volumePath, watcherNotificationFile))
	if err != nil {
		return err
	}

	if updatedConfigMap, ok := event.Object.(*corev1.ConfigMap); ok {
		for settingKey, settingValue := range updatedConfigMap.Data {
			logger.Info(fmt.Sprintf("Creating/updating settings file: %s ", settingKey))
			filePath := path.Join(volumePath, settingKey)
			err = os.WriteFile(filePath, []byte(settingValue), fileMode)
			if err != nil {
				return fmt.Errorf("unable to create/update file '%s'. Error: %w", filePath, err)
			}
		}
	}

	_, err = os.Create(path.Join(volumePath, watcherNotificationFile))
	if err != nil {
		return fmt.Errorf("unable to create inotifysettingscreated file. Error: %w", err)
	}

	return nil
}

func deleteSettingsFiles(volumePath string, event watch.Event, logger *zap.Logger) error {
	err := removeFileIfExists(path.Join(volumePath, watcherNotificationFile))
	if err != nil {
		return err
	}

	if updatedConfigMap, ok := event.Object.(*corev1.ConfigMap); ok {
		for settingKey := range updatedConfigMap.Data {
			logger.Info("Deleting settings file: " + settingKey)
			filePath := path.Join(volumePath, settingKey)
			err = os.Remove(filePath)
			if err != nil {
				return fmt.Errorf("unable to delete file '%s'. Error: %w", filePath, err)
			}
		}
	}

	_, err = os.Create(path.Join(volumePath, watcherNotificationFile))
	if err != nil {
		return fmt.Errorf("unable to create inotifysettingscreated file. Error: %w", err)
	}

	return nil
}

func configMapExists(clientSet *kubernetes.Clientset, namespace, configMapName string) (bool, error) {
	_, err := clientSet.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func removeFileIfExists(filePath string) error {
	if err := os.Remove(filePath); err != nil {
		// Check if the error is due to the file not existing, otherwise panic
		if !os.IsNotExist(err) {
			return fmt.Errorf("error deleting file '%s'. Error: %w", filePath, err)
		}
	}

	return nil
}
