package initcmd

import (
	"embed"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// the `init` command generates new workspaces for writing hoser pipelines. The process essentially just
// writes a bunch of filled out templates into a new directory and initializing a bunch of Python magic.

var (
	initFlags = flag.NewFlagSet("init", flag.ExitOnError)
	debug     = initFlags.Bool("v", false, "Print debugging information")
)

func Usage() {
	fmt.Fprintf(os.Stderr, "usage: hoser init [flags] [path]\n")
	initFlags.PrintDefaults()
}

func Run(args []string) int {
	initFlags.Usage = Usage
	initFlags.Parse(args)

	setupLogging()

	if initFlags.NArg() == 0 {
		log.Error().Msg("missing path to workspace to init")
		return 1
	}

	if initFlags.NArg() > 1 {
		log.Error().Msg("too many arguments (only 1 path to workspace allowed)")
		return 1
	}

	python, err := findPython()
	if err != nil {
		log.Error().Err(err).Msg("python not installed")
		return 1
	}
	log.Info().Msgf("python found at: %s", python)

	workpath := initFlags.Arg(0)
	if err = copyTemplates(workpath); err != nil {
		log.Error().Err(err).Msg("cannot copy file templates")
		return 1
	}

	if err = createVenv(python, workpath); err != nil {
		log.Error().Err(err).Msg("cannot create venv")
		return 1
	}
	return 0
}

func setupLogging() {
	var lvl zerolog.Level
	if *debug {
		lvl = zerolog.DebugLevel
	} else {
		lvl = zerolog.WarnLevel
	}
	log.Logger = log.Output(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.Out = os.Stderr
	})).Level(lvl)
}

func findPython() (string, error) {
	check := []string{
		"python",
		"python3",
	}

	for _, exe := range check {
		python, err := exec.LookPath(exe)
		if err != nil {
			log.Info().Msgf("%s missing: %v", exe, err)
			continue
		}
		return python, nil
	}
	return "", fmt.Errorf("python executable file not found in PATH")
}

//go:embed template template/.hoser/*
var templates embed.FS

func copyTemplates(workpath string) error {
	return fs.WalkDir(templates, "template", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		base, _ := filepath.Rel("template", path)
		to := filepath.Join(workpath, base)
		log.Info().Str("from", base).Str("to", to).Msg("Copying template file")
		if d.IsDir() {
			return os.Mkdir(to, 0755) // ignore dirs
		}

		templateFd, err := templates.Open(path)
		if err != nil {
			return err
		}
		data, err := io.ReadAll(templateFd)
		if err != nil {
			return err
		}

		err = os.WriteFile(to, data, 0644)
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".sh") {
			log.Info().Msgf("marking file '%s as executable", to)
			err := os.Chmod(to, 0754) // read/write/execute
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func createVenv(python, workpath string) error {
	venvCmd := exec.Command(python, "-m", "venv", filepath.Join(workpath, "venv"))
	venvCmd.Stderr = os.Stderr
	if err := venvCmd.Run(); err != nil {
		return err
	}

	install := exec.Command(filepath.Join(workpath, "venv", "bin", "pip"), "install", "hoser")
	install.Stderr = os.Stderr
	if err := install.Run(); err != nil {
		return err
	}

	requirements, err := os.OpenFile(filepath.Join(workpath, "requirements.txt"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	freeze := exec.Command(filepath.Join(workpath, "venv", "bin", "pip"), "freeze")
	freeze.Stderr = os.Stderr
	freeze.Stdout = requirements
	return freeze.Run()
}
