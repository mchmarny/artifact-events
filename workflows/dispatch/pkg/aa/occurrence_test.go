package aa

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetOccurrence(t *testing.T) {
	list := []string{
		"projects/cloudy-demos/occurrences/d61b51a7-6b67-4d8e-b048-a4bb88bd4121",
		"projects/cloudy-demos/occurrences/679b1ad9-c24f-4fe3-a5f3-8ef8abac2b87",
		"projects/cloudy-demos/occurrences/8d67732c-c722-4bde-9cf4-0535105828e8",
	}

	ctx := context.Background()

	for i, v := range list {
		o, err := GetOccurrence(ctx, v)
		assert.Nil(t, err, "test %d: %s", i, err)
		assert.NotNil(t, o, "test %d: %s", i, err)
		assert.Equal(t, v, o.Name, "test %d: %s", i, err)
	}
}
