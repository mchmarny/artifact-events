package slack

import (
	"context"
	"fmt"
	"os"

	"github.com/mchmarny/artifact-events/workflows/dispatch/pkg/aa"
	"github.com/mchmarny/artifact-events/workflows/dispatch/pkg/secret"
	"github.com/pkg/errors"
	s "github.com/slack-go/slack"
)

// Sender sends an occurrence to Slack.
func Sender(ctx context.Context, occ *aa.Occurrence) error {
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

	chID := os.Getenv("SLACK_CHANNEL_ID")
	if chID == "" {
		return errors.New("SLACK_CHANNEL_ID is not set")
	}

	b, err := secret.GetSecret()
	if err != nil {
		return errors.Wrap(err, "failed to get secret")
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
