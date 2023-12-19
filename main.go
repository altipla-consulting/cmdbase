package cmdbase

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/altipla-consulting/errors"
	"github.com/spf13/cobra"
)

var executeMain *cobra.Command

func Main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := executeMain.ExecuteContext(ctx); err != nil {
		slog.Error(err.Error())
		slog.Debug(errors.Stack(err))
		os.Exit(1)
	}
}
