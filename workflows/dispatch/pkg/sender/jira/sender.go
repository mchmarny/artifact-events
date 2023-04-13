package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	j "github.com/andygrunwald/go-jira/v2/cloud"
	"github.com/mchmarny/artifact-events/workflows/dispatch/pkg/aa"
	"github.com/mchmarny/artifact-events/workflows/dispatch/pkg/secret"
	"github.com/pkg/errors"
)

var (
	secretProvider provider = secret.GetSecret
)

type provider func() ([]byte, error)

type config struct {
	Username string `json:"username"`
	Token    string `json:"token"`
	URL      string `json:"url"`
}

// Sender sends an occurrence to Jira.
func Sender(ctx context.Context, occ *aa.Occurrence) error {
	if occ == nil {
		return errors.New("occurrence is nil")
	}

	b, err := secretProvider()
	if err != nil {
		return errors.Wrap(err, "failed to get secret")
	}

	var conf config
	if err := json.Unmarshal(b, &conf); err != nil {
		return errors.Wrap(err, "failed to unmarshal secret")
	}

	at := j.BasicAuthTransport{
		Username: conf.Username,
		APIToken: conf.Token,
	}

	c, err := j.NewClient(conf.URL, at.Client())
	if err != nil {
		return errors.Wrap(err, "failed to create Jira client")
	}

	i, r, err := c.Issue.Create(ctx, &j.Issue{
		Fields: &j.IssueFields{
			Reporter: &j.User{
				Name: "vulnerability-bot",
			},
			Description: fmt.Sprintf(
				"Exposure: %s, Severity: %s, Project: %s, Registry: %s",
				occ.Vulnerability.ShortDescription, occ.Vulnerability.Severity, occ.Project, occ.Registry),
			Type: j.IssueType{
				Name: "Bug",
			},
			Project: j.Project{
				Key: "VUL",
			},
			Summary: fmt.Sprintf("Vulnerability %s of severity %s (score: %f) discovered in image %s",
				occ.Vulnerability.ShortDescription, occ.Vulnerability.Severity, occ.Vulnerability.CvssScore, occ.ResourceUri),
		},
	})
	if err != nil {
		log.Printf("error response: %+v", r.Response)
		return errors.Wrap(err, "failed to create Jira issue")
	}
	if r.StatusCode != http.StatusOK && r.StatusCode != http.StatusCreated {
		return errors.Errorf("issue create resulted in an invalid status code: %d (%s)",
			r.StatusCode, r.Status)
	}

	log.Printf("jira issue created: %s - %s", i.ID, i.Key)

	return nil
}
