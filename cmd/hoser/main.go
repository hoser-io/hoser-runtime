package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hoser-io/hoser-runtime/cmd/hoser/initcmd"
	"github.com/hoser-io/hoser-runtime/cmd/hoser/runcmd"
)

var (
	help = flag.Bool("h", false, "Show help info")
)

func main() {
	flag.Usage = showHelp
	flag.Parse()
	if flag.NArg() < 1 {
		showHelp()
		os.Exit(1)
	}

	cmd := flag.Arg(0)
	subargs := os.Args[2:]
	switch cmd {
	case "run":
		os.Exit(runcmd.Run(subargs))
	case "init":
		os.Exit(initcmd.Run(subargs))
	default:
		fmt.Fprintf(os.Stderr, "error: unrecognized command %s, run hoser -h for commands\n", cmd)
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Fprint(os.Stderr, "Hoser is a tool for running and managing hoser data pipelines\n")
	fmt.Fprintf(os.Stderr, `
Usage:
	hoser <command> [arguments]

Commands are:

    run       run a hoser program (.hos file)
    init      create a new hoser workspace
`)
}
