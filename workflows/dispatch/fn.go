package slack

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/mchmarny/artifact-events/workflows/dispatch/pkg/aa"
	"github.com/mchmarny/artifact-events/workflows/dispatch/pkg/pubsub"
	"github.com/mchmarny/artifact-events/workflows/dispatch/pkg/sender/stdout"
	"github.com/pkg/errors"
	ca "google.golang.org/api/containeranalysis/v1"
)

const (
	messageTypeVulnerability = "VULNERABILITY"
)

var (
	// default occurrence sender
	sender OccurrenceSender = stdout.Sender

	// uncomment the one you prefer
	// sender OccurrenceSender = slack.Sender
	// sender OccurrenceSender = jira.Sender
	// sender OccurrenceSender = rest.Sender
)

// OccurrenceSender is a function that sends an occurrence to a specific.
type OccurrenceSender func(ctx context.Context, occ *ca.Occurrence) error

// Execute is the entry point for the Cloud Function.
func Execute(ctx context.Context, m pubsub.PubSubMessage) error {
	log.SetOutput(os.Stdout)

	if m.Data == nil {
		log.Printf("no data")
		return nil
	}

	log.Printf("processing message: %s", string(m.Data))

	var v aa.VulnerabilityMessage
	if err := json.Unmarshal(m.Data, &v); err != nil {
		return errors.Wrap(err, "failed to unmarshal message")
	}

	if v.Kind != messageTypeVulnerability {
		log.Printf("ignoring message of kind: %s", v.Kind)
		return nil
	}

	occ, err := aa.GetOccurrence(ctx, v.Name)
	if err != nil {
		return errors.Wrapf(err, "failed to get occurrence for: %s", v.Name)
	}

	if err := sender(ctx, occ); err != nil {
		log.Printf("error: %v", err)
		return errors.Wrapf(err, "failed to send occurrence for: %s", v.Name)
	}

	return nil
}
