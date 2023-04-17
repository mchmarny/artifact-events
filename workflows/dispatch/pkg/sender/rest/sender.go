package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/mchmarny/artifact-events/workflows/dispatch/pkg/secret"
	"github.com/pkg/errors"
	ca "google.golang.org/api/containeranalysis/v1"
)

var (
	secretProvider provider = secret.GetSecret
)

type provider func() ([]byte, error)

type config struct {
	Token string `json:"token"`
	URL   string `json:"url"`
}

// Sender sends an occurrence to custom REST endpoint.
func Sender(ctx context.Context, occ *ca.Occurrence) error {
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

	d, err := json.Marshal(occ)
	if err != nil {
		return errors.Wrap(err, "failed to marshal occurrence")
	}

	req, err := http.NewRequest(http.MethodPost, conf.URL, bytes.NewBuffer(d))
	if err != nil {
		return errors.Wrap(err, "error creating request")
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", conf.Token))

	c := &http.Client{
		Timeout: 60 * time.Second,
	}
	r, err := c.Do(req)
	if err != nil {
		return errors.Wrap(err, "error posting vulnerability")
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		rb, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("failed to read response body: %v", err)
		}
		log.Println("response headers: ", r.Header)
		log.Println("response body:", string(rb))
		return errors.Errorf("unexpected status code: %d", r.StatusCode)
	}

	return nil
}
