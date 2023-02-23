package cmdbase

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"libs.altipla.consulting/errors"
)

var cmdRoot = &cobra.Command{
	SilenceUsage: true,
}

type RootOption func(name string)

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

func CmdRoot(name, short string, opts ...RootOption) *cobra.Command {
	cmdRoot.Use = name
	cmdRoot.Short = short

	var flagDebug, flagTrace bool
	cmdRoot.PersistentFlags().BoolVar(&flagDebug, "debug", false, "Enable debug logging for this tool.")
	cmdRoot.PersistentFlags().BoolVar(&flagTrace, "trace", false, "Enable trace logging for this tool.")
	cmdRoot.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if flagDebug {
			log.SetLevel(log.DebugLevel)
		}
		if flagTrace {
			log.SetLevel(log.TraceLevel)
		}

		return nil
	}

	return cmdRoot
}
