package progress

import (
	"fmt"
	"github.com/schollz/progressbar/v3"
	"io"
)

type Progress interface {
	io.Writer

	Add(delta int64)
	Inc()

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
	fmt.Println("WRITE")
	return p.progress.Write(data)
}

func (p progress) Add(delta int64) {
	_ = p.progress.Add64(delta)
}

func (p progress) Inc() {
	_ = p.progress.Add64(1)
}

func (p progress) Alter(opts ...Option) Progress {
	p.options = append(p.options, opts...)

	return New(p.w, p.options...)
}
