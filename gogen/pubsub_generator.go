package gogen

import (
	"io"
	"log"

	"github.com/go-generalize/pubsub_gen/misc"
)

type generator struct {
	PackageName       string
	GeneratedFileName string
	FileName          string
	StructName        string
	TopicName         string
}

func (g *generator) generate(writer io.Writer) {
	t, err := misc.GetTemplate("/gogen/pubsub_generator.go.tpl")
	if err != nil {
		log.Fatalf("function_boilerplat template create failed: %v", err)
	}

	err = t.Execute(writer, g)

	if err != nil {
		log.Printf("failed to execute template: %+v", err)
	}
}
