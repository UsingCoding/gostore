package progress

import (
	"github.com/pkg/errors"
	"os"
)

type Mode string

const (
	AutoMode Mode = "auto"
	NoMode   Mode = "no"
)

func Init(m Mode) (Progress, error) {
	switch m {
	case NoMode:
		return NopProgress(), nil
	case AutoMode:
		return initAutoMode(), nil
	default:
		return nil, errors.Errorf("invalid progress mode %s", m)
	}
}

func initAutoMode() Progress {
	w := os.Stderr

	return New(
		w,
		WithOnComplete(func() {
			_, _ = w.Write([]byte{'\n'})
		}),
		WithElapsedTime(false),
		WithPredictTime(false),
		WithTheme(Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
}
