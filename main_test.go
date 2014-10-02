package main

import (
	"bytes"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kr/bubble/cps"
	"github.com/kr/bubble/fun"
	"github.com/kr/bubble/naivegen"
	"github.com/kr/bubble/optimizer"
	"github.com/kr/bubble/parser"
)

func TestCompile(t *testing.T) {
	files, err := filepath.Glob("sample/*.b")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		testonefile(t, file)
	}
}

func testonefile(t *testing.T, name string) {
	src := read(name)
	const magic = "\n// Output:"
	p := bytes.Index(src, []byte(magic))
	if p < 0 {
		return
	}
	want := strings.TrimSpace(
		strings.Replace(string(src[p+len(magic):]), "\n// ", "\n", -1),
	)

	fileSet := token.NewFileSet()
	fileSet.AddFile(name, -1, len(src))
	ast, err := parser.Parse(fileSet, 0)
	if err != nil {
		t.Error(err)
		return
	}

	funp := fun.Convert(ast)
	cexp, r := cps.Convert(funp)
	cexp = optimizer.Optimize(cexp)
	js := naivegen.Gen(cexp, r)

	tmpf, err := ioutil.TempFile("", "bubbletest")
	if err != nil {
		t.Error(err)
		return
	}
	os.Remove(tmpf.Name())
	io.WriteString(tmpf, js)
	tmpf.Seek(0, 0)

	cmd := exec.Command("node")
	cmd.Stdin = tmpf
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		t.Error(err)
		return
	}

	got := strings.TrimSpace(string(out))
	if got != want {
		t.Errorf("%s got %q want %q", name, got, want)
	}
}
