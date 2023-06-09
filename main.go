package cmdbase

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/altipla-consulting/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var executeMain *cobra.Command

func Main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.SetLevel(log.InfoLevel)

	if err := executeMain.ExecuteContext(ctx); err != nil {
		log.Error(err.Error())
		log.Debug(errors.Stack(err))
		os.Exit(1)
	}
}
