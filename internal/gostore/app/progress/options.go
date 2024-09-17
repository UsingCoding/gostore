package progress

import "github.com/schollz/progressbar/v3"

type Option interface {
	Opt() progressbar.Option
}

type OptionFunc func() progressbar.Option

func (o OptionFunc) Opt() progressbar.Option {
	return o()
}

// hack to be able pass max value to schollz/progressbar
type maxOption struct {
	max int64
	OptionFunc
}

func WithMax(m int64) Option {
	return maxOption{
		max:        m,
		OptionFunc: nil,
	}
}

func WithWidth(w int) Option {
	return OptionFunc(func() progressbar.Option {
		return progressbar.OptionSetWidth(w)
	})
}

func WithCleanOnFinish() Option {
	return OptionFunc(func() progressbar.Option {
		return progressbar.OptionClearOnFinish()
	})
}

func WithOnComplete(f func()) Option {
	return OptionFunc(func() progressbar.Option {
		return progressbar.OptionOnCompletion(f)
	})
}

func WithDescription(desc string) Option {
	return OptionFunc(func() progressbar.Option {
		return progressbar.OptionSetDescription(desc)
	})
}

func WithIts() Option {
	return OptionFunc(func() progressbar.Option {
		return progressbar.OptionShowIts()
	})
}

func WithBytes(enabled bool) Option {
	return OptionFunc(func() progressbar.Option {
		return progressbar.OptionShowBytes(enabled)
	})
}

func WithSpinnerType(spinnerType int) Option {
	return OptionFunc(func() progressbar.Option {
		return progressbar.OptionSpinnerType(spinnerType)
	})
}

func WithShowElapsedTime() Option {
	return OptionFunc(func() progressbar.Option {
		return progressbar.OptionShowElapsedTimeOnFinish()
	})
}

func WithPredictTime(predict bool) Option {
	return OptionFunc(func() progressbar.Option {
		return progressbar.OptionSetPredictTime(predict)
	})
}

func WithElapsedTime(elapsed bool) Option {
	return OptionFunc(func() progressbar.Option {
		progressbar.OptionShowIts()
		return progressbar.OptionSetElapsedTime(elapsed)
	})
}

type Theme struct {
	Saucer        string
	AltSaucerHead string
	SaucerHead    string
	SaucerPadding string
	BarStart      string
	BarEnd        string
}

func WithTheme(t Theme) Option {
	return OptionFunc(func() progressbar.Option {
		return progressbar.OptionSetTheme(progressbar.Theme(t))
	})
}
