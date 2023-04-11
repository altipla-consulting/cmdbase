package cmdbase

import (
	"os"

	"github.com/altipla-consulting/errors"
	log "github.com/sirupsen/logrus"
)

func Main() {
	log.SetLevel(log.InfoLevel)

	if err := cmdRoot.Execute(); err != nil {
		log.Error(err.Error())
		log.Debug(errors.Stack(err))
		os.Exit(1)
	}
}
