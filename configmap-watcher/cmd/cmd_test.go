package cmd_test

import (
	"context"
	"errors"
	"go.goms.io/aks/configmap-watcher/cmd"
	"k8s.io/client-go/kubernetes"
	testclient "k8s.io/client-go/kubernetes/fake"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSuccessCommandNoConfigmap(t *testing.T) {
	var cli cmd.KubeClient = &KubectlMock{}
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

type KubectlMock struct {
	kubeconfig string
	userAgent  string
}

func (cli *KubectlMock) CreateClientSet(kubeconfigFile, userAgent string) (kubernetes.Interface, error) {
	return testclient.NewSimpleClientset(), nil
}
