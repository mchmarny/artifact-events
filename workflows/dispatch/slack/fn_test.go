package slack

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessingMessage(t *testing.T) {
	b, err := os.ReadFile("./test.json")
	assert.NoError(t, err)
	sender = func(ctx context.Context, occ *Occurrence) error {
		assert.NotNil(t, occ)
		assert.Equal(t, "projects/cloudy-demos/occurrences/356d0419-453e-41e0-a652-c30a8fda45c4", occ.Name)
		assert.NotEmpty(t, occ.Vulnerability.Severity)
		assert.NotEmpty(t, occ.Vulnerability.PackageIssue)
		assert.NotEmpty(t, occ.Vulnerability.PackageIssue[0].AffectedVersion.Name)
		return nil
	}
	err = processMessage(context.TODO(), b)
	assert.NoError(t, err)
}
