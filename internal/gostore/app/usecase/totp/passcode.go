package totp

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/common/slices"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
)

type PasscodeView struct {
	GeneratePasscode func() (string, error)
	LastCountdown    int64
	Period           int64
}

type PasscodeGenerator func() (string, error)

const (
	period = 30
)

func (s service) PasscodeView(ctx context.Context, name string) (PasscodeView, error) {
	timepoint := time.Now

	generator, err := s.generator(ctx, name, timepoint)
	if err != nil {
		return PasscodeView{}, err
	}

	// calculate estimated time
	f1 := timepoint().Unix() % period
	countdown := period - f1

	return PasscodeView{
		GeneratePasscode: generator,
		LastCountdown:    countdown,
		Period:           period,
	}, nil
}

func (s service) generator(ctx context.Context, name string, timepoint func() time.Time) (PasscodeGenerator, error) {
	secretData, err := s.service.Get(ctx, store.GetParams{
		SecretIndex: makeTOTPIndex(store.SecretIndex{
			Path: name,
		}),
	})
	if err != nil {
		return nil, err
	}

	secret, ok := maybe.JustValid(slices.Find(secretData, func(data store.SecretData) bool {
		return data.Name == secretKey
	}))
	if !ok {
		return nil, errors.New("failed to find secret for totp")
	}

	algorithm, ok := maybe.JustValid(slices.Find(secretData, func(data store.SecretData) bool {
		return data.Name == algKey
	}))
	if !ok {
		return nil, errors.New("failed to find algorithm for totp")
	}

	a, ok := alg.L()[Algorithm(algorithm.Payload)]
	if !ok {
		return nil, errors.Wrapf(err, "unknown algorithm for totp %s", algorithm.Payload)
	}

	res := PasscodeGenerator(func() (string, error) {
		return totp.GenerateCodeCustom(
			string(secret.Payload),
			timepoint(),
			totp.ValidateOpts{
				Period:    period,
				Digits:    otp.DigitsSix,
				Algorithm: a,
			},
		)
	})

	return res, nil
}
