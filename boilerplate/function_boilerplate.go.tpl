package package {{.PackageName}}

import (
    "context"

    "github.com/go-generalize/pubsub_gen/infra"
    "github.com/go-generalize/pubsub_gen/infra/{{ .DirectoryName }}"
)

var (
    {{ .LowerStructName }}Handler = {{ .PackageName }}.NewPubSubHandler(&{{ .LowerStructName }}Subscriber{})
)

// PubSubHandler handles requests in Cloud Functions
// DO NOT EDIT
func PubSubHandler(ctx context.Context, message *infra.Message) error {
    return {{ .LowerStructName }}Handler.PubSubHandler(ctx, message)
}

var _ {{ .PackageName }}.Subscriber = &{{ .LowerStructName }}Subscriber{}

type {{ .LowerStructName }}Subscriber struct {
}

func (s *{{ .LowerStructName }}Subscriber) Run(ctx context.Context, message *{{ .PackageName }}.{{ .StructName }}) error {
    panic("not implemented")
}