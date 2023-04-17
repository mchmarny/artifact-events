package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	ca "google.golang.org/api/containeranalysis/v1"
)

func TestJiraSender(t *testing.T) {
	secretProvider = func() ([]byte, error) {
		return []byte(fmt.Sprintf(`{
			"username": "%s",
			"token": "%s",
			"URL": "https://mchmarny.atlassian.net/"
		}`, os.Getenv("JIRA_USERNAME"),
			os.Getenv("JIRA_TOKEN"))), nil
	}

	b, err := os.ReadFile("../../../test.json")
	if err != nil {
		t.Fatalf("Failed to read occurrence: %v", err)
	}

	var occ ca.Occurrence
	if err := json.Unmarshal(b, &occ); err != nil {
		t.Fatalf("Failed to unmarshal occurrence: %v", err)
	}

	if err := Sender(context.TODO(), &occ); err != nil {
		t.Fatalf("Failed to send occurrence: %v", err)
	}
}
