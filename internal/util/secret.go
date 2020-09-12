package util

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/neflyte/fdwctl/internal/model"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

func SecretIsDefined(secret model.Secret) bool {
	return secret.Value != "" || secret.FromEnv != "" || secret.FromFile != "" ||
		(secret.FromK8sSecret.Namespace != "" && secret.FromK8sSecret.SecretName != "" && secret.FromK8sSecret.SecretKey != "")
}

func GetSecret(ctx context.Context, secret model.Secret) (string, error) {
	log := logger.Log(ctx).
		WithField("function", "GetSecret")
	// (1) Explicit value
	if secret.Value != "" {
		log.Trace("returning Value")
		return secret.Value, nil
	}
	// (2) Environment variable
	if secret.FromEnv != "" {
		secValue, ok := os.LookupEnv(secret.FromEnv)
		if ok {
			log.Trace("returning FromEnv")
			return secValue, nil
		}
		log.Trace("FromEnv FAILED")
	}
	// (3) Flat file
	if secret.FromFile != "" {
		rawData, err := ioutil.ReadFile(secret.FromFile)
		if err != nil {
			return "", logger.ErrorfAsError(log, "error reading file %s: %s", secret.FromFile, err)
		}
		log.Trace("returning FromFile")
		return fmt.Sprintf("%s", rawData), nil
	}
	// (4) K8s Secret
	if secret.FromK8sSecret.Namespace != "" && secret.FromK8sSecret.SecretName != "" && secret.FromK8sSecret.SecretKey != "" {
		kubectlArgs := []string{
			fmt.Sprintf("-n %s", secret.FromK8sSecret.Namespace),
			"get",
			"secret",
			secret.FromK8sSecret.SecretName,
			fmt.Sprintf("-o jsonpath={.data.%s}", secret.FromK8sSecret.SecretKey),
		}
		log.Tracef("command: kubectl %s", strings.Join(kubectlArgs, " "))
		k8sCmd := exec.CommandContext(ctx, "kubectl", kubectlArgs...)
		err := k8sCmd.Run()
		if err != nil {
			return "", logger.ErrorfAsError(log, "error spawning kubectl: %s", err)
		}
		b64Bytes, err := k8sCmd.Output()
		if err != nil {
			return "", logger.ErrorfAsError(log, "error reading kubectl output: %s", err)
		}
		rawBytes, err := base64.StdEncoding.DecodeString(string(b64Bytes))
		if err != nil {
			return "", logger.ErrorfAsError(log, "error decoding base64: %s", err)
		}
		log.Trace("returning FromK8sSecret")
		return string(rawBytes), nil
	}
	// We didn't get the secret...
	return "", errors.New("unable to get value for secret")
}
