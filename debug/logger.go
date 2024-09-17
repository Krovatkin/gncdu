package debug

import (
	"fmt"
	"github.com/bastengao/gncdu/config"
	"log"
	"os"
)

var logger *log.Logger

func init() {
	file, err := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		fmt.Println(err)
	}
	logger = log.New(file, "logger: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
}

func Info(s ...interface{}) {
	if config.EnableLog {
		logger.Println(s...)
	}
}
