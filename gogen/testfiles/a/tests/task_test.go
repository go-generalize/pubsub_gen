// +build internal

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	task "github.com/gcp-kit/pubsub-gen/gogen/testfiles/a"
	"github.com/gcp-kit/pubsub-gen/infra"
)

type subscriber struct {
}

var receivedMessage *task.Task

func (*subscriber) Run(ctx context.Context, message *task.Task) error {
	receivedMessage = message

	return nil
}

func TestPubsub(t *testing.T) {
	taskFunctionsHandler := task.NewPubSubHandler(&subscriber{})

	client := infra.NewFakeClient(taskFunctionsHandler.PubSubHandler)

	publisher := task.NewPublisher(client)

	ctx := context.Background()

	task := &task.Task{
		Desc:    "Hello, world!",
		Created: time.Unix(time.Now().Unix(), 0),
		Done:    false,
		ID:      10,
	}

	if err := publisher.PublishTask(ctx, task); err != nil {
		t.Fatalf("publish failed: %+v", err)
	}

	if diff := cmp.Diff(task, receivedMessage); diff != "" {
		t.Errorf("received message differed(diff: %s)", diff)
	}
}
