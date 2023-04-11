package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pkg/errors"
	s "github.com/slack-go/slack"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	htransport "google.golang.org/api/transport/http"
)

var (
	// default occurrence sender
	sender OccurrenceSender = slackSender
)

// OccurrenceSender is a function that sends an occurrence to a specific.
type OccurrenceSender func(ctx context.Context, occ *Occurrence) error

// Execute is the entry point for the Cloud Function.
func Execute(ctx context.Context, m PubSubMessage) error {
	if err := processMessage(ctx, m.Data); err != nil {
		log.Printf("error: %v", err)
		return errors.Wrap(err, "error")
	}
	return nil
}

// processMessage processes a Pub/Sub message.
func processMessage(ctx context.Context, data []byte) error {
	log.SetOutput(os.Stdout)
	log.Printf("parsing message: %s", string(data))
	var v VulnerabilityMessage
	if err := json.Unmarshal(data, &v); err != nil {
		return errors.Wrap(err, "failed to unmarshal message")
	}

	if v.Kind != "VULNERABILITY" {
		log.Printf("ignoring message of kind: %s", v.Kind)
		return nil
	}

	occ, err := getOccurrence(ctx, v.Name)
	if err != nil {
		return errors.Wrapf(err, "failed to get occurrence: %s", v.Name)
	}

	return sender(ctx, occ)
}

// getOccurrence gets an occurrence by name.
func getOccurrence(ctx context.Context, name string) (*Occurrence, error) {
	log.Printf("getting occurrence: %s", name)

	c, err := newClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create client")
	}

	u := fmt.Sprintf("https://containeranalysis.googleapis.com/v1/%s", name)

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error client creating request")
	}

	req.Header.Set("Content-Type", "application/json")

	r, err := c.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error getting projects")
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected status code: %d", r.StatusCode)
	}

	var occ Occurrence
	if err := json.NewDecoder(r.Body).Decode(&occ); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	return &occ, nil
}

// newClient creates a new http client.
func newClient(ctx context.Context) (*http.Client, error) {
	var ops []option.ClientOption
	var client *http.Client

	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create default credentials")
	}

	ops = append(ops, option.WithCredentials(creds))
	c, _, err := htransport.NewClient(ctx, ops...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http client")
	}
	client = c

	return client, nil
}

// slackSender sends an occurrence to s.
func slackSender(ctx context.Context, occ *Occurrence) error {
	if occ == nil {
		return errors.New("occurrence is nil")
	}

	fields := make([]s.AttachmentField, 0)
	for _, issue := range occ.Vulnerability.PackageIssue {
		fields = append(fields, s.AttachmentField{
			Title: issue.AffectedPackage,
			Value: issue.EffectiveSeverity,
		})
	}

	secretPath := os.Getenv("SLACK_SECRET_PATH")
	if secretPath == "" {
		return errors.New("SLACK_SECRET_PATH is not set")
	}

	b, err := os.ReadFile(secretPath)
	if err != nil {
		return errors.Wrapf(err, "failed to read secret file from: %s", secretPath)
	}

	chID := os.Getenv("SLACK_CHANNEL_ID")
	if chID == "" {
		return errors.New("SLACK_CHANNEL_ID is not set")
	}

	api := s.New(string(b))
	attachment := s.Attachment{
		Pretext: "Vulnerability Alert",
		Text: fmt.Sprintf("%s - %s",
			occ.Vulnerability.ShortDescription,
			occ.Vulnerability.Severity),
		AuthorLink: occ.ResourceUri,
		Fields:     fields,
	}

	if _, _, err := api.PostMessage(
		chID,
		s.MsgOptionAttachments(attachment),
		s.MsgOptionAsUser(false),
	); err != nil {
		return errors.Wrapf(err,
			"failed to post vulnerability for '%s' to channel: '%s'",
			occ.ResourceUri, chID)
	}

	return nil
}

// PubSubMessage is the payload of a Pub/Sub event.
type PubSubMessage struct {
	Data []byte `json:"data"`
}

// VulnerabilityMessage is the payload of a vulnerability message.
type VulnerabilityMessage struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
}

// Occurrence is the payload of a vulnerability occurrence.
type Occurrence struct {
	Name          string `json:"name"`
	ResourceUri   string `json:"resourceUri"`
	CreationTime  string `json:"createTime"`
	Vulnerability struct {
		Severity     string  `json:"severity"`
		CvssScore    float64 `json:"cvssScore"`
		PackageIssue []struct {
			AffectedPackage string `json:"affectedPackage"`
			AffectedVersion struct {
				Name string `json:"name"`
			} `json:"affectedVersion"`
			FixedPackage      string `json:"fixedPackage"`
			PackageType       string `json:"packageType"`
			EffectiveSeverity string `json:"effectiveSeverity"`
		} `json:"packageIssue"`
		ShortDescription string `json:"shortDescription"`
	} `json:"vulnerability"`
}
