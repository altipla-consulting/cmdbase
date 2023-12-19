package cmdbase

import (
	"log/slog"

	"github.com/kyokomi/emoji/v2"
)

func Successf(format string, args ...interface{}) {
	slog.Info(emoji.Sprintf(":white_check_mark: "+format, args...))
}

func Neutralf(format string, args ...interface{}) {
	slog.Info(emoji.Sprintf(":heavy_check_mark: "+format, args...))
}
