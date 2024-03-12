package cfpackages

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/altipla-consulting/box"
	"github.com/altipla-consulting/errors"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"

	"github.com/altipla-consulting/cmdbase"
	"github.com/altipla-consulting/cmdbase/internal/root"
)

//go:embed bash.sh
var bashInstallScript string

// WithVersion configures install and upgrade commands through our internal
// Cloudflare R2 packages repository to keep the app updated.
func WithVersion(version string) cmdbase.RootOption {
	return func(settings *root.Settings) error {
		var cmdInstallScript = &cobra.Command{
			Use:       "install-script",
			Example:   settings.CmdRoot.Use + " install-script bash",
			Short:     "Emit an install script from the packages repository.",
			Args:      cobra.ExactArgs(1),
			ValidArgs: []string{"bash"},
			Hidden:    true,
		}
		cmdInstallScript.RunE = func(cmd *cobra.Command, args []string) error {
			var script string
			switch args[0] {
			case "bash":
				script = bashInstallScript
			default:
				return errors.Errorf("invalid platform: %s", args[0])
			}

			script = strings.TrimSpace(script)
			tmpl, err := template.New("script").Parse(script)
			if err != nil {
				return errors.Trace(err)
			}
			data := struct {
				App string
			}{
				App: settings.CmdRoot.Use,
			}
			if err := tmpl.Execute(cmd.OutOrStdout(), data); err != nil {
				return errors.Trace(err)
			}

			return nil
		}
		settings.CmdRoot.AddCommand(cmdInstallScript)

		var cmdUpgrade = &cobra.Command{
			Use:     "upgrade",
			Example: settings.CmdRoot.Use + " upgrade",
			Short:   "Upgrade to the latest version.",
			Args:    cobra.NoArgs,
		}
		cmdUpgrade.RunE = func(cmd *cobra.Command, args []string) error {
			slog.Debug("Checking latest version")
			endpoint := "https://packages.altipla.consulting/" + settings.CmdRoot.Use + "/stable.txt"
			req, err := http.NewRequestWithContext(cmd.Context(), http.MethodGet, endpoint, nil)
			if err != nil {
				return errors.Trace(err)
			}
			reply, err := http.DefaultClient.Do(req)
			if err != nil {
				return errors.Trace(err)
			}
			defer reply.Body.Close()
			if reply.StatusCode != http.StatusOK {
				return errors.Errorf("invalid status checking for the latest version: %s", reply.Status)
			}
			remote, err := io.ReadAll(reply.Body)
			if err != nil {
				return errors.Trace(err)
			}
			if strings.TrimSpace(string(remote)) == version {
				cmdbase.Successf("You are already using the latest version %s", version)
				return nil
			}

			run := exec.CommandContext(cmd.Context(), "/bin/bash", "-c", "curl 'https://packages.altipla.consulting/"+settings.CmdRoot.Use+"/install.sh' | bash")
			run.Stdout = os.Stdout
			run.Stderr = os.Stderr
			run.Stdin = os.Stdin
			if err := run.Run(); err != nil {
				return errors.Trace(err)
			}

			return nil
		}
		settings.CmdRoot.AddCommand(cmdUpgrade)

		var cmdVersion = &cobra.Command{
			Use:     "version",
			Example: settings.CmdRoot.Use + " version",
			Short:   "Show the current version.",
			Args:    cobra.NoArgs,
		}
		cmdVersion.RunE = func(cmd *cobra.Command, args []string) error {
			fmt.Println(version)
			return nil
		}
		settings.CmdRoot.AddCommand(cmdVersion)

		prerun := settings.CmdRoot.PersistentPreRunE
		settings.CmdRoot.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			if prerun != nil {
				if err := prerun(cmd, args); err != nil {
					return errors.Trace(err)
				}
			}

			if cmd.Use == "install-script" || cmd.Use == "upgrade" {
				return nil
			}
			if cmd.Parent() != nil && cmd.Parent().Use == "completion" {
				return nil
			}

			if version == "dev" || os.Getenv("CI") != "" {
				slog.Debug("Skip version check against production.")
				return nil
			}

			cachedir, err := os.UserCacheDir()
			if err != nil {
				return errors.Trace(err)
			}
			cacheFilename := filepath.Join(cachedir, settings.CmdRoot.Use, "last-version-check")

			content, err := os.ReadFile(cacheFilename)
			if err != nil && !os.IsNotExist(err) {
				return errors.Trace(err)
			}
			var lastCheck time.Time
			if content != nil {
				lastCheck, err = time.Parse(time.RFC3339, string(content))
				if err != nil {
					return errors.Trace(err)
				}
			}
			if time.Since(lastCheck) < time.Hour {
				slog.
					With("last-check", lastCheck.Format(time.RFC3339)).
					Debug("Skip version check because it was checked less than an hour ago.")
				return nil
			}

			slog.Debug("Checking latest version")
			ctxtimeout, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
			defer cancel()
			endpoint := "https://packages.altipla.consulting/" + settings.CmdRoot.Use + "/stable.txt"
			req, err := http.NewRequestWithContext(ctxtimeout, http.MethodGet, endpoint, nil)
			if err != nil {
				return errors.Trace(err)
			}
			reply, err := http.DefaultClient.Do(req)
			if err != nil {
				slog.Debug("Error checking for latest version", slog.String("error", err.Error()))
				return nil
			}
			defer reply.Body.Close()
			if reply.StatusCode != http.StatusOK {
				slog.Debug("Error cheking for the latest version", slog.String("status", reply.Status), slog.String("endpoint", endpoint))
				return nil
			}
			remote, err := io.ReadAll(reply.Body)
			if err != nil {
				return errors.Trace(err)
			}
			if remoteVersion := strings.TrimSpace(string(remote)); remoteVersion != version {
				var o box.Box
				o.AddLine("Update available. ", aurora.Red(version), " -> ", aurora.Green(remoteVersion))
				o.AddLine("Run ", aurora.Blue(settings.CmdRoot.Use+" upgrade"), " to update.")
				o.Render()
			} else {
				slog.Debug("Already using the latest version", slog.String("local", version), slog.String("remote", remoteVersion))
				if err := os.MkdirAll(filepath.Dir(cacheFilename), 0700); err != nil {
					return errors.Trace(err)
				}
				if err := os.WriteFile(cacheFilename, []byte(time.Now().Format(time.RFC3339)), 0600); err != nil {
					return errors.Trace(err)
				}
			}

			return nil
		}

		return nil
	}
}

func max(nums ...int) int {
	if len(nums) == 0 {
		panic("max() called with no arguments")
	}
	max := nums[0]
	for _, num := range nums {
		if num > max {
			max = num
		}
	}
	return max
}
