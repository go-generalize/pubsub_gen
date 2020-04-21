package misc

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
)

// FindTopicConst finds topic from pubsub generated code
func FindTopicConst(dir string) (string, error) {
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, dir, nil, parser.AllErrors)

	if err != nil {
		return "", err
	}

	for name, v := range pkgs {
		if strings.HasSuffix(name, "_test") {
			continue
		}

		return findTopicConstInPackage(v)
	}

	return "", nil
}

func findTopicConstInPackage(pkg *ast.Package) (string, error) {
	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			if genDecl.Tok != token.CONST {
				continue
			}

			for _, spec := range genDecl.Specs {
				// 型定義
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}

				if len(valueSpec.Names) != 1 || valueSpec.Names[0].Name != "topic" {
					continue
				}

				lit, ok := valueSpec.Values[0].(*ast.BasicLit)

				if !ok || lit.Kind != token.STRING {
					continue
				}

				return strconv.Unquote(lit.Value)
			}
		}
	}

	return "", fmt.Errorf("no topic const")
}
