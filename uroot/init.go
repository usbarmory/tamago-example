package bbtamago

import "log"
import bbmain "bb.u-root.com/bb/pkg/bbmain"
import "os"

func exit(i int) {
	log.Printf("exit %d", i)
}

func init() {
	log.Printf("os.Stdin %v %T", os.Stdin, os.Stdin)
	onexit = func() {
		bbmain.Exit = exit
		log.Printf("os.Stdin %v %T", os.Stdin, os.Stdin)
		var buf [2]byte
		n, err := os.Stdin.Read(buf[:])
		log.Printf("%v %d %v", buf, n, err)
			
		log.Printf("let's call echo")
		os.Args = []string{"echo", "hi", "there"}
		err = bbmain.Run("echo")
		log.Printf("it returned with %v", err)
		log.Printf("let's call forth")
		err = bbmain.Run("forth")
		log.Printf("it returned with %v", err)
	}
}
