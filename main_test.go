package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kr/bubble/build"
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
	src, err := ioutil.ReadFile(name)
	if err != nil {
		t.Error(name, err)
		return
	}
	const magic = "\n// Output:"
	p := bytes.Index(src, []byte(magic))
	if p < 0 {
		return
	}
	want := strings.TrimSpace(
		strings.Replace(string(src[p+len(magic):]), "\n// ", "\n", -1),
	)

	tmpf, err := ioutil.TempFile("", "bubbletest")
	if err != nil {
		t.Error(name, err)
		return
	}

	err = build.Build(tmpf, name)
	if err != nil {
		t.Error(name, err)
		return
	}
	tmpf.Seek(0, 0)

	cmd := exec.Command("node")
	cmd.Stdin = tmpf
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		t.Error(name, err)
		return
	}

	got := strings.TrimSpace(string(out))
	if testing.Verbose() {
		t.Log(name + "\n" + got)
	}
	if got != want {
		t.Errorf("%s got %q want %q", name, got, want)
	}
}
