package progress

import (
	"io"

	"github.com/schollz/progressbar/v3"
)

type Progress interface {
	io.Writer

	Add(delta int64)
	Inc()

	// Finish fills bar to full
	Finish()
	// Exit progress and leave progress at current state
	Exit()

	Alter(opts ...Option) Progress
}

func New(w io.Writer, options ...Option) Progress {
	opts := []progressbar.Option{
		progressbar.OptionSetWriter(w),
	}

	var maxV int64 = -1

	for _, option := range options {
		m, ok := option.(maxOption)
		if ok {
			maxV = m.max
			continue
		}

		opts = append(opts, option.Opt())
	}

	p := progressbar.NewOptions64(maxV, opts...)

	return progress{
		progress: p,
		w:        w,
		options:  options,
	}
}

type progress struct {
	progress *progressbar.ProgressBar

	w       io.Writer
	options []Option
}

func (p progress) Write(data []byte) (n int, err error) {
	return p.progress.Write(data)
}

func (p progress) Add(delta int64) {
	_ = p.progress.Add64(delta)
}

func (p progress) Inc() {
	_ = p.progress.Add64(1)
}

func (p progress) Finish() {
	_ = p.progress.Finish()
}

func (p progress) Exit() {
	_ = p.progress.Exit()
}

func (p progress) Alter(opts ...Option) Progress {
	p.options = append(p.options, opts...)

	return New(p.w, p.options...)
}
