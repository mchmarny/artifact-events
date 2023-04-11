package slack

import (
	"context"
	"testing"

	"github.com/mchmarny/artifact-events/workflows/dispatch/pkg/pubsub"
	"github.com/stretchr/testify/assert"
)

var (
	validMsgData = []byte(`{
		"name": "projects/cloudy-demos/occurrences/356d0419-453e-41e0-a652-c30a8fda45c4",
		"kind": "VULNERABILITY",
		"notificationTime": "2023-04-11T03:21:10.010946Z"
	}`)

	invalidKindMsgData = []byte(`{
		"name": "projects/cloudy-demos/occurrences/356d0419-453e-41e0-a652-c30a8fda45c4",
		"kind": "DISCOVERY",
		"notificationTime": "2023-04-11T03:21:10.010946Z"
	}`)
)

func TestValidMessage(t *testing.T) {
	t.Parallel()
	err := Execute(context.TODO(), pubsub.PubSubMessage{
		Data: validMsgData,
	})
	assert.NoError(t, err)
}

func TestInvalidKindMessage(t *testing.T) {
	t.Parallel()
	err := Execute(context.TODO(), pubsub.PubSubMessage{
		Data: invalidKindMsgData,
	})
	assert.NoError(t, err)
}
