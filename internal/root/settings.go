package root

import (
	"github.com/spf13/cobra"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Settings struct {
	FileLogger *lumberjack.Logger
	CmdRoot    *cobra.Command
}
