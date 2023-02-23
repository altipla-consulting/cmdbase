package cmdbase

import (
	"github.com/kyokomi/emoji/v2"
	log "github.com/sirupsen/logrus"
)

func Successf(format string, args ...interface{}) {
	log.Println(emoji.Sprintf(":white_check_mark: "+format, args...))
}

func Neutralf(format string, args ...interface{}) {
	log.Println(emoji.Sprintf(":heavy_check_mark: "+format, args...))
}
