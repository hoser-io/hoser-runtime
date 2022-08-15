package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/c-bata/go-prompt"
	"github.com/hoser-io/hoser-runtime/hosercmd"
	"github.com/hoser-io/hoser-runtime/interpreter"
	"github.com/hoser-io/hoser-runtime/supervisor"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	help  = flag.Bool("h", false, "Show help info")
	debug = flag.Bool("v", false, "Print debug information to stderr")
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [.hos file] [-v name=value]\n", os.Args[0])
}

func main() {
	os.Exit(run())
}

func run() int {
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "bad path: no Hoser file given\n")
		os.Exit(1)
	}
	hosfile := flag.Arg(0)
	hosfd, err := os.Open(hosfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "no Hoser file found: %v\n", err)
		os.Exit(1)
	}
	cmds, err := hosercmd.ReadFiles(hosfd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", hosfile, err)
		if err, ok := err.(*hosercmd.Error); ok {
			fmt.Fprintf(os.Stderr, "  context: %s\n", string(err.Context))
		}
		os.Exit(1)
	}

	var lvl zerolog.Level
	if *debug {
		lvl = zerolog.DebugLevel
	} else {
		lvl = zerolog.WarnLevel
	}
	log.Logger = zerolog.New(zerolog.NewConsoleWriter()).Level(lvl)
	flag.Usage = usage

	super := supervisor.New(filepath.Join(os.TempDir(), fmt.Sprintf("hoser.%d", os.Getpid())))
	defer super.Close()

	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	errch := super.ServeBackground(ctx)
	preter := interpreter.New(super)
	for _, cmd := range cmds {
		err := preter.Exec(ctx, cmd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			cmdBytes, _ := cmd.Body.MarshalJSON()
			fmt.Fprintf(os.Stderr, "\tcontext: %s %s\n", cmd.Code, cmdBytes)
		}
	}

	// for {
	// 	t := prompt.Input("> ", completer)
	// 	if t == "exit" {
	// 		stop()
	// 		break
	// 	} else {
	// 		cmd, err := hosercmd.Read([]byte(t))
	// 		if err != nil {
	// 			fmt.Fprintf(os.Stderr, "Bad command: %v", err)
	// 			continue
	// 		}
	// 		err = preter.Exec(ctx, cmd)
	// 		if err != nil {
	// 			fmt.Fprintf(os.Stderr, "Fail: %v", err)
	// 			continue
	// 		}
	// 		fmt.Fprintf(os.Stderr, "OK")
	// 	}
	// }

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	for {
		select {
		case err = <-errch:
			// context.Canceled is sent if the pipeline is canceled through an exit command
			if err != nil && err != context.Canceled {
				fmt.Fprintf(os.Stderr, "serve failed: %v\n", err)
				stop()
				return 1
			}
			fmt.Fprintf(os.Stderr, "exiting\n")
			return 0
		case s := <-sig:
			fmt.Fprintf(os.Stderr, "signal: %v", s)
			return 1
		}
	}
}

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}
