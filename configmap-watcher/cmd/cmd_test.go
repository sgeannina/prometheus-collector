package cmd_test

import (
	"context"
	"errors"
	"github.com/spf13/cobra"
	"go.goms.io/aks/configmap-watcher/cmd"
	"io/fs"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	testclient "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	k8stesting "k8s.io/client-go/testing"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSuccessCommandNoConfigmap(t *testing.T) {
	var cli cmd.KubeClient = &KubectlMock{}
	cli.CreateClientSet("kubeconfig-file", "user-agent")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tmpDir := t.TempDir()

	rootCmd := cmd.NewKubeCommand(&cli)
	rootCmd.SetArgs([]string{"--kubeconfig-file=/config/fake/kubeconfig", "--settings-volume=" + tmpDir, "--configmap-name=ama-metrics-settings-configmap", "--configmap-namespace=kube-system"})

	go func() {
		if err := rootCmd.ExecuteContext(ctx); err != nil && !errors.Is(ctx.Err(), context.DeadlineExceeded) {
			t.Fatalf("Command execution failed: %v", err)
		}
	}()

	// Wait for the context to be done (either by command completion or timeout)
	<-ctx.Done()
}

func TestInvalidParameter(t *testing.T) {
	var cli cmd.KubeClient = &KubectlMock{}
	rootCmd := cmd.NewKubeCommand(&cli)
	rootCmd.SetArgs([]string{"test", "value"})

	err := rootCmd.Execute()
	assert.EqualError(t, err, "invalid parameter: --kubeconfig-file is required")
}

func TestSuccessCommandWhenConfigmapExists(t *testing.T) {
	t.Logf("Case 1: Watch Create")
	data := loadConfigmapFromFile(t, "../tests/settings-configmap-create.yaml")
	fakeClient, watch := createFakeClient()
	// the command runs indefinitely in a loop, therefore we need to cancel it after a while
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Simulate watch event
	_, tmpDir := executeConfigmapWatch(t, ctx, fakeClient)
	watch.Add(data)
	time.Sleep(1 * time.Second)

	// Wait for the context to be done
	<-ctx.Done()

	// Assert result
	files, _ := ioutil.ReadDir(tmpDir)
	assert.Equal(t, 9, len(files))
	// TODO: Move these to the watch_test
	assert.True(t, fileExists(files, "inotifysettingscreated"))
	assert.True(t, fileExists(files, "default-targets-scrape-interval-settings"))
	assert.True(t, fileExists(files, "pod-annotation-based-scraping"))
	assert.True(t, fileExists(files, "prometheus-collector-settings"))
	assert.True(t, fileExists(files, "schema-version"))
	assert.True(t, fileExists(files, "config-version"))
	assert.True(t, fileExists(files, "debug-mode"))
	assert.True(t, fileExists(files, "default-scrape-settings-enabled"))
	assert.True(t, fileExists(files, "default-targets-metrics-keep-list"))
}

func fileExists(files []fs.FileInfo, fileName string) bool {
	for _, file := range files {
		if file.Name() == fileName {
			return true
		}
	}

	return false
}

func loadConfigmapFromFile(t *testing.T, configmapFile string) *corev1.ConfigMap {
	// Read file content
	fileContent, err := ioutil.ReadFile(configmapFile)
	if err != nil {
		t.Fatalf("Error reading configmap test file: %s", err)
	}

	// Decode the YAML into a ConfigMap object
	decode := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer().Decode
	obj, _, err := decode(fileContent, nil, nil)
	if err != nil {
		t.Fatalf("Error decoding YAML to ConfigMap: %s", err)
	}

	configMap, ok := obj.(*corev1.ConfigMap)
	if !ok {
		t.Fatalf("Decoded object is not a ConfigMap, it is a %T", obj)
	}

	return configMap
}

func createFakeClient() (kubernetes.Interface, *watch.RaceFreeFakeWatcher) {
	fakeClient := testclient.NewSimpleClientset()
	watch := watch.NewRaceFreeFake()
	fakeClient.PrependWatchReactor("configmaps", k8stesting.DefaultWatchReactor(watch, nil))
	return fakeClient, watch
}

func executeConfigmapWatch(t *testing.T, ctx context.Context, fakeClient kubernetes.Interface) (command *cobra.Command, settingsVolume string) {
	// Implement the specific behavior for this test case
	var cli cmd.KubeClient = &KubectlMock{
		clientSet: fakeClient,
	}
	tmpDir := t.TempDir()
	rootCmd := cmd.NewKubeCommand(&cli)
	rootCmd.SetArgs([]string{"--kubeconfig-file=/config/fake/kubeconfig", "--settings-volume=" + tmpDir, "--configmap-name=ama-metrics-settings-configmap", "--configmap-namespace=kube-system"})

	go func() {
		if err := rootCmd.ExecuteContext(ctx); err != nil && !errors.Is(ctx.Err(), context.DeadlineExceeded) {
			t.Fatalf("Command execution failed: %v", err)
		}
	}()

	return rootCmd, tmpDir
}

// KubectlMock is a mock implementation of the KubeClient interface
type KubectlMock struct {
	kubeconfig string
	userAgent  string
	clientSet  kubernetes.Interface
}

func (cli *KubectlMock) CreateClientSet(kubeconfigFile, userAgent string) (kubernetes.Interface, error) {
	if cli.clientSet == nil {
		cli.clientSet = testclient.NewSimpleClientset()
	}

	return cli.clientSet, nil
}
