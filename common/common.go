package common

// Must panics if err is not nil.
func Must(err error) {
	if err != nil {
		panic(err)
	}
}
