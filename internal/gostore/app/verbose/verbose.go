package verbose

import "github.com/pkg/errors"

type Verbose uint

const (
	Level0 = iota
	Level1
	Level2
	Level3
)

func Ensure(i uint) Verbose {
	err := Valid(i)
	if err != nil {
		panic(err)
	}

	return Verbose(i)
}

func Valid(i uint) error {
	switch i {
	case Level0, Level1, Level2, Level3:
		return nil
	default:
		return errors.Errorf("invalid verbose level %d", i)
	}
}
