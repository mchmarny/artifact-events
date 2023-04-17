package stdout

import (
	"context"
	"encoding/json"
	"os"

	ca "google.golang.org/api/containeranalysis/v1"

	"github.com/pkg/errors"
)

// Sender marshals the occurrence to stdout.
func Sender(ctx context.Context, occ *ca.Occurrence) error {
	if occ == nil {
		return errors.New("occurrence is nil")
	}

	if err := json.NewEncoder(os.Stdout).Encode(occ); err != nil {
		return errors.Wrap(err, "failed to encode occurrence")
	}

	return nil
}
