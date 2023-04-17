package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	ca "google.golang.org/api/containeranalysis/v1"
)

func TestRESTSender(t *testing.T) {
	b, err := os.ReadFile("../../../test.json")
	if err != nil {
		t.Fatalf("Failed to read occurrence: %v", err)
	}

	var occIn ca.Occurrence
	if err := json.Unmarshal(b, &occIn); err != nil {
		t.Fatalf("Failed to unmarshal occurrence: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		var occOut ca.Occurrence
		if err := json.NewDecoder(req.Body).Decode(&occOut); err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		if occOut.ResourceUri != occIn.ResourceUri {
			http.Error(rw, "invalid occurrence", http.StatusBadRequest)
			return
		}
		rw.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	secretProvider = func() ([]byte, error) {
		return []byte(fmt.Sprintf(`{
			"token": "test-token",
			"URL": "%s"
		}`, server.URL)), nil
	}

	if err := Sender(context.TODO(), &occIn); err != nil {
		t.Fatalf("Failed to send occurrence: %v", err)
	}
}
