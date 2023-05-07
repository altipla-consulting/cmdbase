package cmdbase

import (
	"os"

	"github.com/altipla-consulting/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var executeMain *cobra.Command

func Main() {
	log.SetLevel(log.InfoLevel)

	if err := executeMain.Execute(); err != nil {
		log.Error(err.Error())
		log.Debug(errors.Stack(err))
		os.Exit(1)
	}
}
