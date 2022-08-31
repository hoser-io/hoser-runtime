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
	help      = flag.Bool("h", false, "Show help info")
	debug     = flag.Bool("v", false, "Print debug information to stderr")
	shellPipe = flag.String("p", "", "Execute a shell pipe command (a la Unix pipes)")
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [.hos file] [-v name=value] [-s]\n", os.Args[0])
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
	log.Logger = log.Output(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.Out = os.Stderr
	})).Level(lvl)
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
			cmdBytes, _ := cmd.MarshalJSON()
			fmt.Fprintf(os.Stderr, "\tcontext: %s %s\n", cmd.Code(), cmdBytes)
		}
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	for {
		select {
		case err = <-errch:
			// context.Canceled is sent if the pipeline is canceled through an exit command
			if err != nil && err != context.Canceled {
				log.Error().Err(err).Msg("serve failed)")
				stop()
				return 1
			}
			log.Info().Msgf("exiting")
			return 0
		case s := <-sig:
			log.Info().Msgf("signal: %v", s)
			return 1
		}
	}
}

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}
