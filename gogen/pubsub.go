package gogen

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
)

func Main() {
	var (
		topic = flag.String("topic", "", "PUBSUB GENERATOR: Topic ID for Cloud PubSub(default: struct name in kebab case)")
	)

	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("You have to specify the struct name of target")
		os.Exit(1)
	}

	if err := runPubSub(flag.Arg(0), *topic); err != nil {
		log.Fatal(err.Error())
	}

	return
}

func runPubSub(structName string, topic string) error {
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, ".", nil, parser.AllErrors)

	if err != nil {
		panic(err)
	}

	for name, v := range pkgs {
		if strings.HasSuffix(name, "_test") {
			continue
		}

		return traverse(v, fs, structName, topic)
	}

	return nil
}

func traverse(pkg *ast.Package, fs *token.FileSet, structName string, topic string) error {
	gen := &generator{
		PackageName: pkg.Name,
		TopicName:   topic,
	}

	for name, file := range pkg.Files {
		gen.FileName = strings.TrimSuffix(filepath.Base(name), ".go")
		gen.GeneratedFileName = gen.FileName + "_gen"

		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			if genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				// 型定義
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				name := typeSpec.Name.Name

				if name != structName {
					continue
				}

				// structの定義
				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}
				gen.StructName = name

				if len(gen.TopicName) == 0 {
					gen.TopicName = strcase.ToKebab(name)
				}

				return generate(gen, fs, structType)
			}
		}
	}

	return fmt.Errorf("no such struct: %s", structName)
}

func generate(gen *generator, fs *token.FileSet, structType *ast.StructType) error {
	fp, err := os.Create(gen.GeneratedFileName + ".go")

	if err != nil {
		return fmt.Errorf("faield ot create %s.go: %+v", gen.GeneratedFileName, err)
	}

	gen.generate(
		fp,
	)

	fp.Close()

	return nil
}
