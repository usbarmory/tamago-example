func init() {
	panic("fuck")
	onexit = func() {
		log.Printf("let's call forth")
		err := bbmain.Run("forth")
		log.Printf("it returned with %v", err)
	}
}
