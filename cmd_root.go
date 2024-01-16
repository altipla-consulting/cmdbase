// Package cmdbase provides base initialization and helpers for our CLI applications.
package cmdbase

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/altipla-consulting/cmdbase/internal/root"
	"github.com/altipla-consulting/errors"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"gopkg.in/natefinch/lumberjack.v2"
)

// RootOption configures the root command.
type RootOption func(settings *root.Settings) error

// WithInstall configures an install command that installs the autocomplete script
// in the user's bashrc.
func WithInstall() RootOption {
	return func(settings *root.Settings) error {
		cmdInstall := &cobra.Command{
			Use:     "install",
			Example: settings.CmdRoot.Use + " install",
			Short:   "Install autocomplete in the user's bashrc.",
			RunE: func(cmd *cobra.Command, args []string) error {
				installLine := `. <(` + settings.CmdRoot.Use + ` completion bash)`

				home, err := os.UserHomeDir()
				if err != nil {
					return errors.Trace(err)
				}
				content, err := os.ReadFile(filepath.Join(home, ".bashrc"))
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

				slog.Info("CLI autocomplete is now installed in ~/.bashrc")
				slog.Info(fmt.Sprintf("Restart the shell to have '%s' available as a command.", settings.CmdRoot.Use))

				return nil
			},
		}
		settings.CmdRoot.AddCommand(cmdInstall)

		return nil
	}
}

// WithUpdate configures an update command that installs using Go the remote repository.
func WithUpdate(pkgname string) RootOption {
	return func(settings *root.Settings) error {
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
		settings.CmdRoot.AddCommand(cmdUpdate)
		return nil
	}
}

// WithFileLogger configures logrus to emit logs to a file with rotation.
func WithFileLogger(config func() (*lumberjack.Logger, error)) RootOption {
	return func(settings *root.Settings) error {
		logger, err := config()
		if err != nil {
			return errors.Trace(err)
		}
		settings.FileLogger = logger

		return nil
	}
}

// CmdRoot creates a new root command. Can only be called once per application.
func CmdRoot(name, short string, opts ...RootOption) *cobra.Command {
	cmdRoot := &cobra.Command{
		Use:           name,
		Short:         short,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	executeMain = cmdRoot

	settings := &root.Settings{
		CmdRoot: cmdRoot,
	}
	var initErr error
	for _, opt := range opts {
		if err := opt(settings); err != nil {
			initErr = err
			break
		}
	}

	var flagDebug bool
	cmdRoot.PersistentFlags().BoolVar(&flagDebug, "debug", false, "Enable debug logging for this tool.")
	cmdRoot.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if initErr != nil {
			return errors.Trace(initErr)
		}

		level := slog.LevelInfo
		if flagDebug {
			level = slog.LevelDebug
		}
		var w io.Writer = os.Stderr
		if settings.FileLogger != nil {
			w = io.MultiWriter(os.Stderr, settings.FileLogger)
		}
		handler := slog.New(tint.NewHandler(w, &tint.Options{
			Level:   level,
			NoColor: !isatty.IsTerminal(os.Stderr.Fd()),
		}))
		slog.SetDefault(handler)

		return nil
	}

	return cmdRoot
}
