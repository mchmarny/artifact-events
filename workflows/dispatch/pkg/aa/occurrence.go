package aa

import (
	"context"
	"log"

	"github.com/pkg/errors"
	ca "google.golang.org/api/containeranalysis/v1"
)

// GetOccurrence gets an occurrence by name.
func GetOccurrence(ctx context.Context, name string) (*ca.Occurrence, error) {
	log.Printf("getting occurrence: %s", name)

	c, err := ca.NewService(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create client")
	}

	o, err := c.Projects.Occurrences.Get(name).Context(ctx).Do()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get occurrence")
	}

	return o, nil
}
