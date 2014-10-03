// Bubble is a toy programming language.
package main

// Much of this compiler is transliterated
// from formulae presented in the following book:
// Andrew Appel, Compiling with Continuations
// (Cambridge University Press, 1992).

import (
	"flag"
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

var (
	flagD = flag.Bool("d", false, "debug")
	flagO = flag.String("o", "", "output file")
	flagR = flag.Bool("r", true, "run program")
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	flag.Parse()
	src := read(flag.Arg(0))

	fileSet := token.NewFileSet()
	fileSet.AddFile(flag.Arg(0), -1, len(src))
	var mode parser.Mode
	if *flagD {
		mode |= parser.Debug
	}
	ast, err := parser.Parse(fileSet, mode)
	if err != nil {
		log.Fatalln(err)
	}
	if *flagD {
		pretty.Fprintf(os.Stderr, "% #v\n", ast)
	}

	funp := fun.Convert(ast)
	if *flagD {
		pretty.Fprintf(os.Stderr, "% #v\n", funp)
	}

	cexp, r := cps.Convert(funp)
	if *flagD {
		pretty.Fprintf(os.Stderr, "% #v\n", cexp)
	}

	cexp = optimizer.Optimize(cexp)
	if *flagD {
		pretty.Fprintf(os.Stderr, "opt % #v\n", cexp)
	}

	js := naivegen.Gen(cexp, r)
	if *flagD {
		cmd := exec.Command("js-beautify", "-f", "-")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		w, err := cmd.StdinPipe()
		if err != nil {
			log.Fatalln(err)
		}
		err = cmd.Start()
		_, err = io.WriteString(w, js)
		if err != nil {
			log.Fatalln(err)
		}
		w.Close()
		cmd.Wait()
	}

	var fout io.Writer
	switch {
	case *flagO != "":
		fout, err = os.Create(*flagO)
	case *flagR:
		cmd := exec.Command("node")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		w, err := cmd.StdinPipe()
		if err != nil {
			log.Fatalln(err)
		}
		err = cmd.Start()
		defer cmd.Wait()
		defer w.Close()
		fout = w
	default:
		return
	}

	_, err = io.WriteString(fout, js)
	if err != nil {
		log.Fatalln(err)
	}
}

func read(name string) []byte {
	f, err := os.Open(name)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalln(err)
	}
	return b
}
