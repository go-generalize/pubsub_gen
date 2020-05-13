package boilerplate

import (
	"flag"
	"github.com/rakyll/statik/fs"
	"golang.org/x/xerrors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/go-generalize/pubsub_gen/misc"
	_ "github.com/go-generalize/pubsub_gen/statik"

	"github.com/iancoleman/strcase"
)

var (
	pubsubBoilerplateTmpl   *template.Template
	functionBoilerplateTmpl *template.Template
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

	var err error
	pubsubBoilerplateTmpl, err = getTemplate("/boilerplate/pubsub_boilerplate.go.gpl")
	if err != nil {
		log.Fatalf("pubsub_boilerplate template create failed: %v", err)
	}

	functionBoilerplateTmpl, err = getTemplate("/boilerplate/function_boilerplate.go.tpl")
	if err != nil {
		log.Fatalf("function_boilerplat template create failed: %v", err)
	}

	runBoilerplate(arg, *force)
}

func getTemplate(templatePath string) (*template.Template, error) {
	statikFs, err := fs.New()
	if err != nil {
		return nil, xerrors.Errorf("assets open failed: %w", err)
	}

	f, err := statikFs.Open(templatePath)
	if err != nil {
		return nil, xerrors.Errorf("template open failed: %w", err)
	}

	t, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, xerrors.Errorf("template read failed: %w", err)
	}

	return template.Must(template.New("tmpl").Parse(string(t))), nil
}

type argument struct {
	PackageName     string
	LowerStructName string
	StructName      string
	DirectoryName   string
}

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
