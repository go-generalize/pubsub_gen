package infra

import (
	"context"
	"time"

	"google.golang.org/api/iterator"

	"google.golang.org/api/option"

	"cloud.google.com/go/functions/metadata"

	"cloud.google.com/go/pubsub"
)

type Client interface {
	CreateTopic(ctx context.Context, id string) error
	Publish(ctx context.Context, id string, message []byte) error
	ListTopics(ctx context.Context) ([]string, error)
}

type client struct {
	pubsubClient *pubsub.Client
}

func NewClient(ctx context.Context, projectID string, opts ...option.ClientOption) (Client, error) {
	pubsubClient, err := pubsub.NewClient(ctx, projectID, opts...)

	if err != nil {
		return nil, err
	}

	return &client{
		pubsubClient: pubsubClient,
	}, nil
}

var _ Client = &client{}

func (c *client) CreateTopic(ctx context.Context, id string) error {
	_, err := c.pubsubClient.CreateTopic(ctx, id)

	return err
}

func (c *client) Publish(ctx context.Context, id string, message []byte) error {
	_, err := c.pubsubClient.Topic(id).Publish(ctx, &pubsub.Message{
		Data: message,
	}).Get(ctx)

	return err
}

func (c *client) ListTopics(ctx context.Context) ([]string, error) {
	it := c.pubsubClient.Topics(ctx)

	topics := make([]string, 0, 10)
	for {
		topic, err := it.Next()

		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, err
		}

		topics = append(topics, topic.ID())
	}

	return topics, nil
}

type fakeClient struct {
	fakeHandler func(ctx context.Context, message *Message) error

	topics []string
}

func (c *fakeClient) CreateTopic(ctx context.Context, id string) error {
	return nil
}

func (c *fakeClient) Publish(ctx context.Context, id string, message []byte) error {
	ctx = metadata.NewContext(ctx, &metadata.Metadata{
		EventID:   "event-id",
		Timestamp: time.Now(),
		EventType: "example.com/" + id,
	})

	return c.fakeHandler(ctx, &Message{Data: message})
}

func (c *fakeClient) ListTopics(ctx context.Context) ([]string, error) {
	return c.topics, nil
}

func NewFakeClient(fakeHandler func(ctx context.Context, message *Message) error, topics ...string) Client {
	return &fakeClient{fakeHandler: fakeHandler, topics: topics}
}
