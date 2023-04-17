package slack

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mchmarny/artifact-events/workflows/dispatch/pkg/secret"
	"github.com/pkg/errors"
	s "github.com/slack-go/slack"
	ca "google.golang.org/api/containeranalysis/v1"
)

var (
	secretProvider provider = secret.GetSecret
)

type provider func() ([]byte, error)

type config struct {
	ChannelID string `json:"channel_id"`
	Token     string `json:"token"`
}

// Sender sends an occurrence to Slack.
func Sender(ctx context.Context, occ *ca.Occurrence) error {
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

	b, err := secretProvider()
	if err != nil {
		return errors.Wrap(err, "failed to get secret")
	}

	var conf config
	if err := json.Unmarshal(b, &conf); err != nil {
		return errors.Wrap(err, "failed to unmarshal secret")
	}

	api := s.New(conf.Token)
	attachment := s.Attachment{
		Pretext: "Vulnerability Alert",
		Text: fmt.Sprintf("%s - %s",
			occ.Vulnerability.ShortDescription,
			occ.Vulnerability.Severity),
		AuthorLink: occ.ResourceUri,
		Fields:     fields,
	}

	if _, _, err := api.PostMessage(
		conf.ChannelID,
		s.MsgOptionAttachments(attachment),
		s.MsgOptionAsUser(false),
	); err != nil {
		return errors.Wrapf(err,
			"failed to post vulnerability for '%s' to channel: '%s'",
			occ.ResourceUri, conf.ChannelID)
	}

	return nil
}
