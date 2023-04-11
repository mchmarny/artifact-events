package secret

import (
	"os"

	"github.com/pkg/errors"
)

const (
	secretPathDefault = "/secrets/dispatcher"
	secretPathEnvVar  = "SECRET_PATH"
)

func GetSecret() ([]byte, error) {
	secretPath := os.Getenv("secretPathEnvVar")
	if secretPath == "" {
		secretPath = secretPathDefault
	}

	b, err := os.ReadFile(secretPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read secret file from: %s", secretPath)
	}

	return b, nil
}
