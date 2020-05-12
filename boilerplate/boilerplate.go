package boilerplate

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/go-generalize/pubsub_gen/misc"

	"github.com/iancoleman/strcase"
)

func Main() {
	var (
		packageName   = flag.String("p", "", "Package name for generated code(default: all chars in lower case)")
		directoryName = flag.String("d", "", "Directory name for generated code(default: struct name in snake case)")
		force         = flag.Bool("f", false, "Overwrite existing files")
	)

	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatalf("struct name is not set")
	}

	structName := flag.Arg(0)

	pkg := strings.ToLower(strcase.ToLowerCamel(structName))

	if *packageName != "" {
		pkg = *packageName
	}

	dir := strcase.ToSnake(structName)

	if *directoryName != "" {
		dir = *directoryName
	}

	arg := &argument{
		PackageName:     pkg,
		LowerStructName: strcase.ToLowerCamel(structName),
		StructName:      structName,
		DirectoryName:   dir,
	}

	runBoilerplate(arg, *force)
}

type argument struct {
	PackageName     string
	LowerStructName string
	StructName      string
	DirectoryName   string
}

var (
	pubsubBoilerplateTmpl = template.Must(template.New("tmpl").Parse(pubsubBoilerplate))

	functionBoilerplateTmpl = template.Must(template.New("tmpl").Parse(functionBoilerplate))
)

func runBoilerplate(arg *argument, force bool) {
	goRoot := filepath.Dir(misc.GetGoModPath())

	if goRoot == "" {
		log.Fatalf("failed to get go module root path")
	}

	pubsubDir := filepath.Join(goRoot, "infra/pubsub", arg.DirectoryName)
	functionsDir := filepath.Join(goRoot, "infra/functions", arg.DirectoryName)

	// 既に存在するかどうか確認
	if _, err := os.Stat(pubsubDir); err == nil && !force {
		log.Fatalf("pubsub dir(%s) already exists(set -f to overwrite existing files)", pubsubDir)
	}
	if _, err := os.Stat(functionsDir); err == nil && !force {
		log.Fatalf("functions dir(%s) already exists(set -f to overwrite existing files)", functionsDir)
	}

	if err := os.MkdirAll(pubsubDir, 0700); err != nil {
		log.Fatalf("failed to create directory(%s): %+v", pubsubDir, err)
	}
	if err := os.MkdirAll(functionsDir, 0700); err != nil {
		log.Fatalf("failed to create directory(%s): %+v", functionsDir, err)
	}

	pubsubFile := filepath.Join(pubsubDir, arg.LowerStructName+".go")
	functionFile := filepath.Join(functionsDir, arg.LowerStructName+".go")

	generatePubSub(arg, pubsubFile)
	generateFunction(arg, functionFile)

	cmd := exec.Command("go", "generate", "./...")
	cmd.Dir = pubsubDir
	b, err := cmd.CombinedOutput()

	if err != nil {
		log.Fatalf("go generate failed: %+v %s", err, string(b))
	}
}

func generatePubSub(arg *argument, pubsubFile string) {
	pubsubGo, err := os.Create(pubsubFile)

	if err != nil {
		log.Fatalf("failed to create %s: %+v", pubsubFile, err)
	}

	if err := pubsubBoilerplateTmpl.Execute(pubsubGo, arg); err != nil {
		log.Fatalf("failed to execute template for pubsub: %+v", err)
	}
}

func generateFunction(arg *argument, functionFile string) {
	functionGo, err := os.Create(functionFile)

	if err != nil {
		log.Fatalf("failed to create %s: %+v", functionFile, err)
	}

	if err := functionBoilerplateTmpl.Execute(functionGo, arg); err != nil {
		log.Fatalf("failed to execute template for Cloud Functions: %+v", err)
	}
}

// language=go
const pubsubBoilerplate = `// Code generated by pubsub generator DO NOT EDIT.
package {{ .PackageName }}

//go:generate pubsub_generator gogen {{ .StructName }}

type {{ .StructName }} struct {
	
}
`

// language=go
const functionBoilerplate = `package
import "cloud.google.com/go/pubsub" {{ .PackageName }}

import (
	"context"

	"github.com/go-generalize/pubsub_gen/infra"
	"ggithub.com/go-generalize/pubsub_gen/infra/{{ .DirectoryName }}"
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
`
