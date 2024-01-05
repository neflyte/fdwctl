package util

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/neflyte/fdwctl/lib/logger"
	"github.com/neflyte/fdwctl/lib/model"
)

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
		rawData, err := os.ReadFile(secret.FromFile)
		if err != nil {
			return "", logger.ErrorfAsError(log, "error reading file %s: %s", secret.FromFile, err)
		}
		log.Trace("returning FromFile")
		return string(rawData), nil
	}
	// (4) K8s Secret
	if secret.FromK8sSecret.Namespace != "" && secret.FromK8sSecret.SecretName != "" && secret.FromK8sSecret.SecretKey != "" {
		kubectlArgs := []string{
			"-c",
			fmt.Sprintf(
				"kubectl -n %s get secret %s -o jsonpath={.data.%s}",
				secret.FromK8sSecret.Namespace,
				secret.FromK8sSecret.SecretName,
				secret.FromK8sSecret.SecretKey,
			),
		}
		var out bytes.Buffer
		k8sCmd := exec.CommandContext(ctx, "/bin/sh", kubectlArgs...)
		k8sCmd.Stdout = &out
		err := k8sCmd.Run()
		if err != nil {
			return "", logger.ErrorfAsError(log, "error spawning kubectl: %s", err)
		}
		rawBytes, err := base64.StdEncoding.DecodeString(out.String())
		if err != nil {
			return "", logger.ErrorfAsError(log, "error decoding base64: %s", err)
		}
		log.Trace("returning FromK8sSecret")
		return string(rawBytes), nil
	}
	// We didn't get the secret...
	return "", errors.New("unable to get value for secret")
}
