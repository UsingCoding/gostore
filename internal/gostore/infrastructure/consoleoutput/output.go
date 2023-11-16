package consoleoutput

import (
	"fmt"
	"io"
)

type Output interface {
	io.Writer

	Printf(format string, args ...any)
	OKf(format string, args ...any)
	Errorf(format string, args ...any)
}

func New(writer io.Writer, opts ...Opt) Output {
	c := cfg{}
	for _, opt := range opts {
		opt.apply(&c)
	}

	return &output{
		writer: writer,
		cfg:    c,
	}
}

type cfg struct {
	prefix  string
	newline bool
}

type output struct {
	writer io.Writer
	cfg    cfg
}

func (o output) Write(p []byte) (n int, err error) {
	return o.writer.Write(p)
}

func (o output) Printf(format string, args ...any) {
	fmt.Fprintf(o.writer, o.prefix()+format+o.newline(), args...)
}

func (o output) OKf(format string, args ...any) {
	fmt.Fprintf(o.writer, o.prefix()+"✅ "+format+o.newline(), args...)
}

func (o output) Errorf(format string, args ...any) {
	fmt.Fprintf(o.writer, o.prefix()+"❌ "+format+o.newline(), args...)
}

func (o output) prefix() string {
	return o.cfg.prefix
}

func (o output) newline() string {
	if o.cfg.newline {
		return "\n"
	}
	return ""
}
