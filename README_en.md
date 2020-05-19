## PubSub

- `make init` must be completed before performing the following steps.
- A tool that automatically generates the code group necessary to use PubSub and actually operates by hitting the API.

### How to use
#### pubsub_generator boilerplate
- Specify the struct name and form the basis of the required file.
- One struct is associated with one topic
- Assumption that "If you throw a message from GAE to PubSub, it will be processed by Cloud Functions"

```console
$ pubsub_generator boilerplate Task # go
$ cat infra/pubsub/test/test.go
// Code generated by pubsub generator DO NOT EDIT.
package test

//go:generate pubsub_generator gogen Test

type Test struct {
        
}
$ cat infra/functions/test/test.go
package test

import (
	"context"

	pubsub "github.com/ORG_NAME/REPO_NAME/server/infra/pubsub/test"
)

func init() {
	pubsub.Subscribe(&testSubscriber{})
}

var _ pubsub.Subscriber = &testSubscriber{}

type testSubscriber struct {
}

func (s *testSubscriber) Run(ctx context.Context, message *pubsub.Test) error {
	panic("not implemented")
}
```

Write the content you want to pass via PubSub in the contents of `Test struct` (it must be exported because it is JSON encoded)
Also, write the content you want to execute with Cloud Functions in the Run method. If you do not want to execute with Cloud Functions, you need to delete the ones generated inside infra / functions.

pubsub_generator boilerplate can optionally overwrite directory and package names. Once created, the struct will not be created again.
```console
$ pubsub_generator boilerplate -help   
Usage of pubsub_generator:
  -d string
        Directory name for generated code(default: struct name in snake case)
  -f    Overwrite existing files
  -p string
        Package name for generated code(default: all chars in lower case)
```

#### pubsub_generator deploy
- Command to deploy a function to Cloud Functions
- Deploy the function in "infra/functions"
- Identify the topic name by matching the folder name in "infra/functions" with "infra/pubsub"
- Running `go mod vendor` or copying files internally to use private repository in Go Module mode

```console
Usage of pubsub_generator:
  -dry-run
        Don't create actual resources(print the diff and exit)
  -options string
        Options for gcloud functions deploy
  -region string
        Region to deploy functions (default "asia-northeast1")
  -runtime string
        Runtime for Cloud Functions (default "go111")
```

#### pubsub_generator migrate
- Automatically create missing PubSub topics
- Search automatically generated files under "infra/pubsub" and generate topic
- Specifying the project-id option is required

```console
Usage of pubsub_generator:
  -dry-run
        Don't create actual resources(print the diff and exit)
  -project-id string
        Project ID for GCP
```