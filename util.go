package soonsocks

import (
	"log"
	"os"
)

var Logger = log.New(os.Stdout, "SSLogger: ", log.Ldate | log.Ltime)

