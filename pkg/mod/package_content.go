package mod

import (
	"html/template"
	"io"
	"os"
)

type PackageContent struct {
	PkgName    string
	ImportPath string
	Values     []string
	Types      []string
}

func (p PackageContent) WriteFile(output string) error {
	out, err := os.Create(output)
	if err != nil {
		return err
	}
	p.write(out)
	return nil
}

var tmpl = template.Must(template.ParseFiles("register.template"))

func (p PackageContent) write(output io.Writer) error {
	return tmpl.Execute(output, p)
}
