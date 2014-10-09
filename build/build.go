package build

import (
	"go/token"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

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

func BuildFile(w io.Writer, name string, mode int) error {
	src, err := ioutil.ReadFile(name)
	if err != nil {
		return err
	}

	fileSet := token.NewFileSet()
	fileSet.AddFile(name, -1, len(src))
	var pmode parser.Mode
	if mode&Debug != 0 {
		pmode |= parser.Debug
	}
	ast, err := parser.Parse(fileSet, pmode)
	if err != nil {
		log.Fatalln(err)
	}
	if mode&Debug != 0 {
		pretty.Fprintf(os.Stderr, "% #v\n", ast)
	}

	funp := fun.Convert(ast)
	if mode&Debug != 0 {
		pretty.Fprintf(os.Stderr, "% #v\n", funp)
	}

	cexp, r := cps.Convert(funp)
	if mode&Debug != 0 {
		pretty.Fprintf(os.Stderr, "% #v\n", cexp)
	}

	cexp = optimizer.Optimize(cexp)
	if mode&Debug != 0 {
		pretty.Fprintf(os.Stderr, "opt % #v\n", cexp)
	}

	js := naivegen.Gen(cexp, r)
	if mode&Debug != 0 {
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
