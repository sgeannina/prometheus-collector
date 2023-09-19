// +gocover:ignore:file - main function
package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.goms.io/aks/rp/toolkit/log"
)

var rootCmd = &cobra.Command{
	Use:   "configmap-watcher",
	Short: "This binary will watch a configmap and load the values in a pod volume",
	Run: func(cmd *cobra.Command, args []string) {
		run()
	},
}

var (
	kubeconfigFile     string
	configmapNamespace string
	configmapName      string
	mutex              *sync.Mutex
)

func init() {
	rootCmd.Flags().StringVar(&kubeconfigFile, "kubeconfig-file", "", "Path to the kubeconfig")
	rootCmd.Flags().StringVar(&configmapNamespace, "configmap-namespace", "kube-system", "The configmap namespace")
	rootCmd.Flags().StringVar(&configmapName, "configmap-name", "", "The configmap name")
}

func run() {
	logger := newConfigmapWatcherLogger()
	ctx := log.WithLogger(context.Background(), logger)
	mutex = &sync.Mutex{}

	if kubeconfigFile == "" {
		logger.Fatal(ctx, "--kubeconfig-file is required")
	}

	// TODO: Logging is probably all wrong
	userAgent := fmt.Sprintf("remediator/%s %s/%s", "Version", "Commit", "Date")
	overlayClient, err := createOverlayKubeClient(userAgent)
	if err != nil {
		logger.Fatalf(ctx, "failed to create overlay clientset: %s", err)
	}

	//underlayClient, err := createCXUnderlayKubeClient(userAgent)
	//if err != nil {
	//	logger.Fatalf(ctx, "failed to create cx-underlay clientset: %s", err)
	//}

	WatchForChanges(overlayClient, configmapNamespace, configmapName, mutex)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		logger.Info(ctx, "os interrupt SIGTERM, exiting...")
		// 143=128+SIGTERM, https://tldp.org/LDP/abs/html/exitcodes.html
		os.Exit(143)
	}()
}

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

// createCXUnderlayKubeClient constructs a kube client instance for current cx underlay.
func createCXUnderlayKubeClient(userAgent string) (kubernetes.Interface, error) {
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("error read in cluster config: %w", err)
	}
	restConfig.UserAgent = userAgent
	return kubernetes.NewForConfig(restConfig)
}
