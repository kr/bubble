// Bubble is a toy programming language.
package main

// Much of this compiler is transliterated
// from formulae presented in the following book:
// Andrew Appel, Compiling with Continuations
// (Cambridge University Press, 1992).

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/kr/bubble/build"
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
	if *flagD {
		build.Mode |= build.Debug
	}

	if s := os.Getenv("BUBBLEROOT"); s != "" {
		build.BUBBLEROOT = s
	}
	if s := os.Getenv("BUBBLEPATH"); s != "" {
		build.BUBBLEPATH = s
	}

	var (
		targ *os.File
		err  error
	)

	if *flagO != "" {
		targ, err = os.Create(*flagO)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		targ, err = ioutil.TempFile("", "bubble")
		if err != nil {
			log.Fatalln(err)
		}
		os.Remove(targ.Name())
	}

	err = build.Build(targ, flag.Args()...)
	if err != nil {
		log.Fatalln(err)
	}

	if *flagO == "" && *flagR {
		targ.Seek(0, 0)
		c := exec.Command("node")
		c.Stdin = targ
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		err = c.Run()
		if err != nil {
			log.Fatalln(err)
		}
	}
}
