package build

import (
	"errors"
	"go/token"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kr/bubble/ast"
	"github.com/kr/bubble/cps"
	"github.com/kr/bubble/fun"
	"github.com/kr/bubble/naivegen"
	"github.com/kr/bubble/optimizer"
	"github.com/kr/bubble/parser"
	"github.com/kr/pretty"
)

// mode flags
const (
	Debug = 1 << iota // print debug info to stderr
)

var Mode int

var (
	BUBBLEROOT string // initialized from linker flag at build time
	BUBBLEPATH string
)

type pkg struct {
	importPath string
	*ast.Package
}

func Build(w io.Writer, file ...string) error {
	pkgs, err := parseProgram(file)
	if err != nil {
		log.Fatalln(err)
	}

	pkgtab := make(map[string]fun.Tab)
	tabf := func(s string) fun.Tab {
		return pkgtab[s]
	}
	var seq []fun.Exp
	for _, p := range pkgs {
		exp, ptab := fun.Convert(p.Package, tabf)
		pkgtab[p.importPath] = ptab
		seq = append(seq, exp)
		if Mode&Debug != 0 {
			pretty.Fprintf(os.Stderr, "% #v\n", exp)
		}
	}

	cexp, r := cps.Convert(seq)
	if Mode&Debug != 0 {
		pretty.Fprintf(os.Stderr, "% #v\n", cexp)
	}

	cexp = optimizer.Optimize(cexp)
	if Mode&Debug != 0 {
		pretty.Fprintf(os.Stderr, "opt % #v\n", cexp)
	}

	js := naivegen.Gen(cexp, r)
	if Mode&Debug != 0 {
		cmd := exec.Command("js-beautify", "-f", "-")
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Fatalln(err)
		}
		err = cmd.Start()
		_, err = io.WriteString(stdin, js)
		if err != nil {
			log.Fatalln(err)
		}
		stdin.Close()
		cmd.Wait()
	}

	_, err = io.WriteString(w, js)
	if err != nil {
		log.Fatalln(err)
	}
	return nil
}

func parseProgram(files []string) ([]*pkg, error) {
	main, err := parseFiles(files)
	if err != nil {
		return nil, err
	}
	// TODO(kr): give main a valid import path
	return parseDeps(main, nil)
}

func parseDeps(p *pkg, tab []*pkg) ([]*pkg, error) {
	// TODO(kr): detect cycles
	for _, f := range p.Files {
		for _, spec := range f.Imports {
			path := spec.Path.String()
			if !containsPackage(tab, path) {
				dep, err := parsePackage(path)
				if err != nil {
					return nil, err
				}
				more, err := parseDeps(dep, tab)
				if err != nil {
					return nil, err
				}
				tab = append(tab, more...)
			}
		}
	}
	return append(tab, p), nil
}

func parsePackage(path string) (*pkg, error) {
	names, err := packageFiles(path)
	if err != nil {
		return nil, err
	}
	p, err := parseFiles(names)
	if err != nil {
		return nil, err
	}
	p.importPath = path
	return p, nil
}

func parseFiles(names []string) (*pkg, error) {
	if len(names) == 0 {
		return nil, errors.New("must supply at least one file to build")
	}
	fileSet := token.NewFileSet()
	for _, name := range names {
		src, err := ioutil.ReadFile(name)
		if err != nil {
			return nil, err
		}

		fileSet.AddFile(name, -1, len(src))
	}
	var pmode parser.Mode
	if Mode&Debug != 0 {
		pmode |= parser.Debug
	}
	ast, err := parser.Parse(fileSet, pmode)
	if err != nil {
		return nil, err
	}
	if Mode&Debug != 0 {
		pretty.Fprintf(os.Stderr, "% #v\n", ast)
	}
	return &pkg{"", ast}, nil
}

func packageFiles(path string) ([]string, error) {
	dir, err := findPackage(path)
	if err != nil {
		return nil, err
	}

	return filepath.Glob(filepath.Join(dir, "*.b"))
}

func findPackage(importPath string) (dir string, err error) {
	fpath := filepath.FromSlash(importPath)
	search := []string{BUBBLEROOT}
	search = append(search, filepath.SplitList(BUBBLEPATH)...)
	for _, base := range search {
		dir := filepath.Join(base, "src", fpath)
		if st, err := os.Stat(dir); err == nil && st.IsDir() {
			return dir, nil
		}
	}
	return "", errors.New("package not found: " + importPath)
}

func containsPackage(a []*pkg, path string) bool {
	for _, p := range a {
		if p.importPath == path {
			return true
		}
	}
	return false
}
