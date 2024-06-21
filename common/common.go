package common

// Must panics if err is not nil.
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Assert(mustTrue bool, message string) {
	if !mustTrue {
		panic(message)
	}
}
