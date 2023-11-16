package consoleoutput

type Opt interface {
	apply(o *cfg)
}

type optFunc func(o *cfg)

func (opt optFunc) apply(o *cfg) {
	opt(o)
}

func WithPrefix(prefix string) Opt {
	return optFunc(func(o *cfg) {
		o.prefix = prefix
	})
}

func WithNewline(newline bool) Opt {
	return optFunc(func(o *cfg) {
		o.newline = newline
	})
}
