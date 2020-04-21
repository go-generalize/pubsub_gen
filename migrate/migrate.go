package migrate

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"

	"github.com/ORG_NAME/REPO_NAME/server/infra/pubsub"

	"github.com/ORG_NAME/REPO_NAME/server/tools/pubsub_generator/misc"
)

func Main() {
	var (
		dryRun    = flag.Bool("dry-run", false, "Don't create actual resources(print the diff and exit)")
		projectID = flag.String("project-id", "", "Project ID for GCP")
	)
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	if err := runMigrate(ctx, *dryRun, *projectID); err != nil {
		log.Fatal(err.Error())
	}

	return
}

func runMigrate(ctx context.Context, dryRun bool, projectID string) error {
	goRoot := filepath.Dir(misc.GetGoModPath())

	if goRoot == "" {
		return fmt.Errorf("go package directory is not found")
	}

	pubsubRoot := filepath.Join(goRoot, "infra/pubsub")

	fifos, err := ioutil.ReadDir(pubsubRoot)

	if err != nil {
		return err
	}

	desiredTopics := map[string]struct{}{}
	for i := range fifos {
		if !fifos[i].IsDir() {
			continue
		}

		//nolint:govet
		topic, err := misc.FindTopicConst(filepath.Join(pubsubRoot, fifos[i].Name()))

		if err != nil {
			log.Printf("error occurred in %s: %+v", fifos[i].Name(), err)
		}

		desiredTopics[topic] = struct{}{}
	}

	client, err := pubsub.NewClient(ctx, projectID)

	if err != nil {
		return err
	}

	topics, err := client.ListTopics(ctx)

	if err != nil {
		return err
	}

	for i := range topics {
		delete(desiredTopics, topics[i])
	}

	for topic := range desiredTopics {
		log.Printf("Creating a new topic: %s\n", topic)

		if !dryRun {
			if err := client.CreateTopic(ctx, topic); err != nil {
				return err
			}
		} else {
			log.Println("skipping")
		}

		log.Printf("Done: %s\n", topic)
	}

	return nil
}
