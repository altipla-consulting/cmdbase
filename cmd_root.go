// Package cmdbase provides base initialization and helpers for our CLI applications.
package cmdbase

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/natefinch/lumberjack"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"libs.altipla.consulting/errors"
)

var cmdRoot = &cobra.Command{
	SilenceUsage: true,
}

// RootOption configures the root command.
type RootOption func(name string)

// WithInstall configures an install command that installs the autocomplete script
// in the user's bashrc.
func WithInstall() RootOption {
	return func(name string) {
		cmdInstall := &cobra.Command{
			Use: "install",
			RunE: func(cmd *cobra.Command, args []string) error {
				installLine := `. <(` + name + ` completion bash)`

				home, err := os.UserHomeDir()
				if err != nil {
					return errors.Trace(err)
				}
				content, err := ioutil.ReadFile(filepath.Join(home, ".bashrc"))
				if err != nil {
					return errors.Trace(err)
				}
				lines := strings.Split(string(content), "\n")

				for _, line := range lines {
					if line == installLine {
						return nil
					}
				}

				f, err := os.OpenFile(filepath.Join(home, ".bashrc"), os.O_APPEND|os.O_WRONLY, 0600)
				if err != nil {
					return errors.Trace(err)
				}
				defer f.Close()

				fmt.Fprintln(f)
				fmt.Fprintln(f, installLine)

				log.Info("CLI autocomplete is now installed in ~/.bashrc")
				log.Info("Restart your terminal to finish your setup")

				return nil
			},
		}
		cmdRoot.AddCommand(cmdInstall)
	}
}

// WithUpdate configures an update command that installs using Go the remote repository.
func WithUpdate(pkgname string) RootOption {
	return func(pkgname string) {
		cmdUpdate := &cobra.Command{
			Use: "update",
			RunE: func(cmd *cobra.Command, args []string) error {
				install := exec.Command("go", "install", "-v", pkgname+"@latest")
				install.Stdin = os.Stdin
				install.Stdout = os.Stdout
				install.Stderr = os.Stderr
				if err := install.Run(); err != nil {
					return errors.Trace(err)
				}

				Successf("CLI updated successfully!")

				return nil
			},
		}
		cmdRoot.AddCommand(cmdUpdate)
	}
}

type loggerHook struct {
	logger    *lumberjack.Logger
	formatter log.Formatter
}

func (hook *loggerHook) Levels() []log.Level {
	return log.AllLevels
}

func (hook *loggerHook) Fire(entry *log.Entry) error {
	f, err := hook.formatter.Format(entry)
	if err != nil {
		return err
	}
	_, err = hook.logger.Write(f)
	return err
}

// WithFileLogger configures logrus to emit logs to a file with rotation.
func WithFileLogger(config func() *lumberjack.Logger) RootOption {
	return func(pkgname string) {
		prerun := cmdRoot.PersistentPreRunE
		cmdRoot.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			file := config()
			log.AddHook(&loggerHook{
				logger:    file,
				formatter: new(log.JSONFormatter),
			})
			return errors.Trace(prerun(cmd, args))
		}
	}
}

// CmdRoot creates a new root command. Can only be called once per application.
func CmdRoot(name, short string, opts ...RootOption) *cobra.Command {
	cmdRoot.Use = name
	cmdRoot.Short = short

	var flagDebug, flagTrace bool
	cmdRoot.PersistentFlags().BoolVar(&flagDebug, "debug", false, "Enable debug logging for this tool.")
	cmdRoot.PersistentFlags().BoolVar(&flagTrace, "trace", false, "Enable trace logging for this tool.")
	cmdRoot.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		log.SetFormatter(new(log.TextFormatter))
		switch {
		case flagTrace:
			log.SetLevel(log.TraceLevel)
		case flagDebug:
			log.SetLevel(log.DebugLevel)
		default:
			log.SetLevel(log.InfoLevel)
		}

		return nil
	}

	for _, opt := range opts {
		opt(name)
	}

	return cmdRoot
}
