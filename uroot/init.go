package bbtamago

import "log"
import bbmain "bb.u-root.com/bb/pkg/bbmain"
import "os"

func init() {
	onexit = func() {
		log.Printf("let's call echo")
		os.Args = []string{"echo", "hi", "there"}
		err := bbmain.Run("echo")
		log.Printf("it returned with %v", err)
		log.Printf("let's call forth")
		err = bbmain.Run("forth")
		log.Printf("it returned with %v", err)
	}
}
