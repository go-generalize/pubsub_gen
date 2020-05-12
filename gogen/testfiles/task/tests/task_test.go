package tests

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	taslItem "github.com/go-generalize/pubsub_gen/gogen/testfiles/task"
	"github.com/go-generalize/pubsub_gen/infra"
	"github.com/go-generalize/pubsub_gen/infra/task"
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

	taskItem := &taslItem.Task{
		Desc:    "Hello, world!",
		Created: time.Unix(time.Now().Unix(), 0),
		Done:    false,
		ID:      10,
	}

	if err := publisher.PublishTask(ctx, taskItem); err != nil {
		t.Fatalf("publish failed: %+v", err)
	}

	if diff := cmp.Diff(taskItem, receivedMessage); diff != "" {
		t.Errorf("received message differed(diff: %s)", diff)
	}
}
