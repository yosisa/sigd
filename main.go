package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"code.google.com/p/go-shlex"
)

type sigdef struct {
	name string
	sig  os.Signal
}

var sigdefs = []sigdef{
	{"hup", syscall.SIGHUP},
	{"int", syscall.SIGINT},
	{"term", syscall.SIGTERM},
}

type handler interface {
	fmt.Stringer
	Do() error
}

type cmd struct {
	s    string
	args []string
}

func newCmd(s string) *cmd {
	args, err := shlex.Split(s)
	if err != nil {
		log.Fatal(err)
	}
	return &cmd{s, args}
}

func (c *cmd) Do() error {
	cmd := exec.Command(c.args[0], c.args[1:]...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func (c *cmd) String() string {
	return c.s
}

var (
	stopOnTerm = flag.Bool("stop-on-term", false, "Stop after SIGTERM handler ran")
	stopOnInt  = flag.Bool("stop-on-int", false, "Stop after SIGINT handler ran")
)

func main() {
	opts := make(map[os.Signal]*string)
	for _, s := range sigdefs {
		var str string
		opts[s.sig] = &str
		flag.StringVar(opts[s.sig], s.name, "", "Handler for SIG"+strings.ToUpper(s.name))
	}
	flag.Parse()

	handlers := make(map[os.Signal]handler)
	ch := make(chan os.Signal, 4)
	for sig, s := range opts {
		if *s == "" {
			continue
		}
		handlers[sig] = newCmd(*s)
		signal.Notify(ch, sig)
	}

	for recv := range ch {
		cmd, _ := handlers[recv]
		if err := cmd.Do(); err == nil {
			log.Printf("Handler for %v suceeded: %v", recv, cmd)
		} else {
			log.Printf("Handler for %v failed: %v: %v", recv, cmd, err)
		}
		if *stopOnTerm && recv == syscall.SIGTERM || *stopOnInt && recv == syscall.SIGINT {
			log.Printf("Bye!")
			os.Exit(0)
		}
	}
}
