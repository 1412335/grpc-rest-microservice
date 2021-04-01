package log

import (
	"os"
)

var (
	DefaultLogger Factory
)

func init() {
	DefaultLogger = NewFactory(WithLevel(os.Getenv("LOG_LEVEL")))
}
