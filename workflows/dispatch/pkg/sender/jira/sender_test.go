package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/mchmarny/artifact-events/workflows/dispatch/pkg/aa"
)

var (
	occJSON = []byte(`{
		"name": "projects/cloudy-demos/occurrences/356d0419-453e-41e0-a652-c30a8fda45c4",
		"resourceUri": "https://us-west1-docker.pkg.dev/cloudy-demos/test/image@sha256:5ffd30269c7bde2e29453bb9b8d3618814b7034e37aef299e3c071acbb565911",
		"noteName": "projects/cloudy-demos/notes/CVE-2019-7577",
		"kind": "VULNERABILITY",
		"createTime": "2023-03-30T23:09:28.443028Z",
		"updateTime": "2023-03-30T23:09:28.443028Z",
		"vulnerability": {
			"severity": "MEDIUM",
			"cvssScore": 6.8,
			"packageIssue": [
				{
					"affectedCpeUri": "cpe:/o:canonical:ubuntu_linux:18.04",
					"affectedPackage": "libsdl2",
					"affectedVersion": {
						"name": "2.0.8+dfsg1",
						"revision": "1ubuntu1.18.04.5~18.04.1",
						"kind": "NORMAL",
						"fullName": "2.0.8+dfsg1-1ubuntu1.18.04.5~18.04.1"
					},
					"fixedCpeUri": "cpe:/o:canonical:ubuntu_linux:18.04",
					"fixedPackage": "libsdl2",
					"fixedVersion": {
						"kind": "MAXIMUM"
					},
					"packageType": "OS",
					"effectiveSeverity": "LOW"
				}
			],
			"shortDescription": "CVE-2019-7577",
			"longDescription": "NIST vectors: AV:N/AC:M/Au:N/C:P/I:P/A:P",
			"relatedUrls": [
				{
					"url": "http://people.ubuntu.com/~ubuntu-security/cve/CVE-2019-7577",
					"label": "More Info"
				}
			],
			"effectiveSeverity": "LOW",
			"cvssv3": {
				"baseScore": 8.8,
				"exploitabilityScore": 2.8,
				"impactScore": 5.9,
				"attackVector": "ATTACK_VECTOR_NETWORK",
				"attackComplexity": "ATTACK_COMPLEXITY_LOW",
				"privilegesRequired": "PRIVILEGES_REQUIRED_NONE",
				"userInteraction": "USER_INTERACTION_REQUIRED",
				"scope": "SCOPE_UNCHANGED",
				"confidentialityImpact": "IMPACT_HIGH",
				"integrityImpact": "IMPACT_HIGH",
				"availabilityImpact": "IMPACT_HIGH"
			}
		}
	}`)
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

	var occ aa.Occurrence
	if err := json.Unmarshal(occJSON, &occ); err != nil {
		t.Fatalf("Failed to unmarshal occurrence: %v", err)
	}

	if err := Sender(context.TODO(), &occ); err != nil {
		t.Fatalf("Failed to send occurrence: %v", err)
	}
}
