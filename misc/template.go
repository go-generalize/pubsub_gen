package misc

import (
	"github.com/rakyll/statik/fs"
	"golang.org/x/xerrors"
	"io/ioutil"
	"text/template"
)

func GetTemplate(templatePath string) (*template.Template, error) {
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
