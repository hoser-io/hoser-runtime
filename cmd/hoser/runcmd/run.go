package runcmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/hoser-io/hoser-runtime/hosercmd"
	"github.com/hoser-io/hoser-runtime/interpreter"
	"github.com/hoser-io/hoser-runtime/supervisor"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	runFlags  = flag.NewFlagSet("run", flag.ExitOnError)
	debug     = runFlags.Bool("v", false, "Print debug information to stderr")
	shellPipe = runFlags.String("p", "", "Execute a shell pipe command (a la Unix pipes)")
)

func Usage() {
	fmt.Fprintf(os.Stderr, "usage: hoser run [flags] [hosfiles...]\n")
	runFlags.PrintDefaults()
}

func Run(args []string) int {
	runFlags.Usage = Usage
	if err := runFlags.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	hosfile := runFlags.Arg(0)
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
