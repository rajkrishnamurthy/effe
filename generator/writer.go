package generator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

type writer struct {
	buf bytes.Buffer // Accumulated output.
}

func (w *writer) Printf(format string, args ...interface{}) {
	fmt.Fprintf(&w.buf, format, args...)
}

func (w *writer) format() []byte {
	src, err := format.Source(w.buf.Bytes())
	if err != nil {
		return w.buf.Bytes()
	}
	return src
}

func (w *writer) writeNode(fset *token.FileSet, node interface{}) error {
	err := format.Node(&w.buf, fset, node)
	if err != nil {
		return err
	}
	w.Printf("\n")
	return nil
}

func (w *writer) writeType(fset *token.FileSet, typeSpec *ast.TypeSpec) error {
	node := &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			typeSpec,
		},
	}
	err := format.Node(&w.buf, fset, node)
	if err != nil {
		return err
	}
	w.Printf("\n")
	return nil
}

func (w *writer) reset() {
	w.buf.Reset()
}

func writeDiagrams(pkg *packages.Package, outputDir string, res []drawFlowRes) ([]string, error) {
	pkgDir, err := detectOutputDir(pkg.GoFiles)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(filepath.Join(pkgDir, outputDir), 0755)
	if err != nil {
		return nil, err
	}

	outputFiles := make([]string, len(res))
	buffer := new(bytes.Buffer)
	for index, flowRes := range res {
		outputName := filepath.Join(pkgDir, outputDir, fmt.Sprintf("%s.plantuml", flowRes.name))
		buffer.WriteString("@startuml\n")
		buffer.WriteString(fmt.Sprintf("right footer - %s\n", flowRes.name))
		buffer.WriteString("scale 1.2\n\nskinparam monochrome true\nskinparam SequenceBoxBackgroundColor #FAFAFA\nskinparam SequenceBoxBorderColor #F0F0F0\nhide footbox\n")
		buffer.WriteString(flowRes.graph)
		buffer.WriteString("\n@enduml")
		err = ioutil.WriteFile(outputName, buffer.Bytes(), 0600)
		if err != nil {
			return nil, err
		}
		buffer.Reset()
		outputFiles[index] = outputName
	}
	return outputFiles, nil
}

func writeGeneratedCode(pkg *packages.Package, p *pkgGen) (string, error) {
	w := &writer{}
	outDir, err := detectOutputDir(pkg.GoFiles)
	if err != nil {
		return "", err
	}
	outputName := filepath.Join(outDir, "effe_gen.go")

	w.Printf("// Code generated by Effe. DO NOT EDIT.\n")
	w.Printf("\n")
	w.Printf("//+build !effeinject\n")
	w.Printf("\n")
	w.Printf("package %s\n", pkg.Name)

	if len(p.imports) > 0 {
		w.Printf("import (\n")
		sort.Strings(p.imports)
		for _, im := range p.imports {
			if im != "" {
				w.Printf("\"%s\"\n", im)
			}
		}
		w.Printf(")\n")
	}

	w.Printf("\n")

	firstFuncDecls := append(p.flowFuncDecls, p.depInitializerFuncDecls...)

	for _, firstFuncDecl := range firstFuncDecls {
		err = w.writeNode(pkg.Fset, firstFuncDecl)
		if err != nil {
			return "", err
		}
	}

	for _, t := range p.typeSpecs {
		err = w.writeType(pkg.Fset, t)
		if err != nil {
			return "", err
		}
	}

	for _, f := range p.implFuncDecls {
		err = w.writeNode(pkg.Fset, f)
		if err != nil {
			return "", err
		}
	}

	return outputName, ioutil.WriteFile(outputName, w.format(), 0600)
}

func detectOutputDir(paths []string) (string, error) {
	if len(paths) == 0 {
		return "", errors.WithStack(errors.New("no files to derive output directory from"))
	}
	dir := filepath.Dir(paths[0])
	for _, p := range paths[1:] {
		if dir2 := filepath.Dir(p); dir2 != dir {
			return "", errors.Errorf("found conflicting directories %q and %q", dir, dir2)
		}
	}
	return dir, nil
}
