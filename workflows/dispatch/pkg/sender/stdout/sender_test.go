package stdout

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	ca "google.golang.org/api/containeranalysis/v1"
)

func TestStdOutSender(t *testing.T) {
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
