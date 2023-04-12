package aa

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	t "google.golang.org/api/transport/http"
)

const (
	minRegParts = 2
)

// Occurrence is the payload of a vulnerability occurrence.
type Occurrence struct {
	Name          string `json:"name"`
	ResourceUri   string `json:"resourceUri"`
	Project       string `json:"project"`
	Registry      string `json:"registry"`
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

// GetOccurrence gets an occurrence by name.
func GetOccurrence(ctx context.Context, name string) (*Occurrence, error) {
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

	occ.Project, occ.Registry, err = parseResourceURI(occ.ResourceUri)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse resource uri")
	}

	return &occ, nil
}

// parseResourceURI parses the resource uri and returns project and registry.
// https://us-west1-docker.pkg.dev/cloudy-demos/test/$IMAGE@sha256:5ffd302f1b1b2f2
// https://gcr.io/my-project/busybox
func parseResourceURI(uri string) (string, string, error) {
	if uri == "" {
		return "", "", errors.New("empty resource uri")
	}

	u, err := url.Parse(uri)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to parse resource uri")
	}

	parts := strings.Split(u.Path, "/")

	if len(parts) < minRegParts {
		return "", "", errors.Errorf("invalid resource uri: %s", uri)
	}

	if len(parts) == minRegParts {
		// /my-project/busybox
		return parts[1], "", nil
	}

	// /cloudy-demos/test/image@sha256:5ffd3
	return parts[1], parts[2], nil
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
	c, _, err := t.NewClient(ctx, ops...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http client")
	}
	client = c

	return client, nil
}
