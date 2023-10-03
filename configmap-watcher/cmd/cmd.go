// +gocover:ignore:file - main function
package cmd

import (
	"errors"
	"fmt"
	lgr "go.goms.io/aks/configmap-watcher/logger"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var rootCmd = &cobra.Command{
	Use:   "configmap-watcher",
	Short: "This binary will watch a configmap and load the values in a pod volume",
	Run: func(cmd *cobra.Command, args []string) {
		run()
	},
}

var (
	// ExitSignal 143=128+SIGTERM, https://tldp.org/LDP/abs/html/exitcodes.html
	ExitSignal         = 143
	kubeconfigFile     string
	settingsVolume     string
	configmapNamespace string
	configmapName      string
	mutex              *sync.Mutex
)

func init() {
	rootCmd.Flags().StringVar(&kubeconfigFile, "kubeconfig-file", "", "Path to the kubeconfig")
	rootCmd.Flags().StringVar(&configmapNamespace, "configmap-namespace", "kube-system", "The configmap namespace")
	rootCmd.Flags().StringVar(&configmapName, "configmap-name", "", "The configmap name")
	rootCmd.Flags().StringVar(&settingsVolume, "settings-volume", "", "Directory where the settings files are stored")
}

func run() {
	logger := lgr.SetupLogger(os.Stdout, "ConfigMapWatcher")
	defer logger.Sync() //nolint:errcheck
	mutex = &sync.Mutex{}

	if err := validateParameters(); err != nil {
		logger.Panic("Invalid parameter.", zap.Error(err))
	}

	// TODO: Find a way to get the version, commit and date from the build
	userAgent := fmt.Sprintf("configmap-watcher/%s %s/%s", "Version", "Commit", "Date")
	overlayClient, err := createOverlayKubeClient(userAgent)
	if err != nil {
		logger.Panic("failed to create overlay clientset", zap.Error(err))
	}

	WatchForChanges(overlayClient, logger, configmapNamespace, configmapName, settingsVolume)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		logger.Info("os interrupt SIGTERM, exiting...")
		os.Exit(ExitSignal)
	}()
}

// Execute executes the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// createOverlayKubeClient constructs a kube client instance for current overlay cluster.
func createOverlayKubeClient(userAgent string) (*kubernetes.Clientset, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigFile)
	if err != nil {
		return nil, fmt.Errorf("error build overlay kubeconfig: %w", err)
	}
	cfg.UserAgent = userAgent
	return kubernetes.NewForConfig(cfg)
}

func validateParameters() error {
	if kubeconfigFile == "" {
		return errors.New("--kubeconfig-file is required")
	}

	if settingsVolume == "" {
		return errors.New("--settings-volume is required")
	}

	if configmapName == "" {
		return errors.New("--configmap-name is required")
	}

	if configmapNamespace == "" {
		return errors.New("--configmap-namespace is required")
	}

	return nil
}
