package aa

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResourceURLParsing(t *testing.T) {
	valid := []string{
		"https://gcr.io/project/image@sha256:123",
		"https://gcr.io/project/image:v1.2.3",
		"https://gcr.io/project/image",
		"https://us-west1-docker.pkg.dev/project/repo/image@sha256:5ffd302f1b1b2f2",
		"https://us-west1-docker.pkg.dev/project/repo/image:v1.2.3",
		"https://us-west1-docker.pkg.dev/project/repo/image",
		"https://gcr.io/project/image@sha256:123",
		"https://gcr.io/project/image:v1.2.3",
		"https://gcr.io/project/image",
		"https://us-west1-docker.pkg.dev/project/repo/image@sha256:5ffd302f1b1b2f2",
		"https://us-west1-docker.pkg.dev/project/repo/image:v1.2.3",
		"https://us-west1-docker.pkg.dev/project/repo/image",
	}

	for i, v := range valid {
		p, r, err := parseResourceURI(v)
		assert.NoError(t, err, "uri[%d]: %s", i, v)
		assert.Equal(t, "project", p, "uri[%d]: %s", i, v)
		if !strings.HasPrefix(v, "https://gcr.io") {
			assert.Equal(t, "repo", r, "uri[%d]: %s", i, v)
		}
	}
}
