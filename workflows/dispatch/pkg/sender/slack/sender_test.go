package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	ca "google.golang.org/api/containeranalysis/v1"
)

func TestSlackSender(t *testing.T) {
	secretProvider = func() ([]byte, error) {
		return []byte(fmt.Sprintf(`{
			"channel_id": "%s",
			"token": "%s"
		}`, os.Getenv("SLACK_CHANNEL"),
			os.Getenv("SLACK_TOKEN"))), nil
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
