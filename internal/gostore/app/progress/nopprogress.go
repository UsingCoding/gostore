package progress

func NopProgress() Progress {
	return nopProgress{}
}

type nopProgress struct{}

func (np nopProgress) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (np nopProgress) Add(int64) {
	return
}

func (np nopProgress) Inc() {
	return
}

func (np nopProgress) Alter(...Option) Progress {
	return np
}
