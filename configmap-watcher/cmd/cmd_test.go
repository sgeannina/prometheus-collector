package cmd_test

import (
	"go.goms.io/aks/configmap-watcher/cmd"
	"k8s.io/client-go/kubernetes"
	testclient "k8s.io/client-go/kubernetes/fake"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSuccessCommand(t *testing.T) {
	var cli cmd.KubeClient = &KubectlMock{}
	rootCmd := cmd.NewKubeCommand(&cli)
	rootCmd.SetArgs([]string{"--kubeconfig-file=/config/fake/kubeconfig", "--settings-volume=/etc/config/settings", "--configmap-name=ama-metrics-settings-configmap", "--configmap-namespace=kube-system"})
	err := rootCmd.Execute()
	assert.Nil(t, err)
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
