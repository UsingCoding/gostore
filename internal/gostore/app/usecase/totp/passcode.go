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

func (s service) Passcode(ctx context.Context, name string) (string, error) {
	secretData, err := s.service.Get(ctx, store.GetParams{
		SecretIndex: makeTOTPIndex(store.SecretIndex{
			Path: name,
		}),
	})
	if err != nil {
		return "", err
	}

	secret, ok := maybe.JustValid(slices.Find(secretData, func(data store.SecretData) bool {
		return data.Name == secretKey
	}))
	if !ok {
		return "", errors.Wrapf(err, "failed to find secret for totp")
	}

	algorithm, ok := maybe.JustValid(slices.Find(secretData, func(data store.SecretData) bool {
		return data.Name == algKey
	}))
	if !ok {
		return "", errors.Wrapf(err, "failed to find algorithm for totp")
	}

	a, ok := alg.L()[Algorithm(algorithm.Payload)]
	if !ok {
		return "", errors.Wrapf(err, "unknown algorithm for totp %s", algorithm.Payload)
	}

	passcode, err := totp.GenerateCodeCustom(
		string(secret.Payload),
		time.Now(),
		totp.ValidateOpts{
			Period:    30, // valid for 30 sec
			Digits:    otp.DigitsSix,
			Algorithm: a,
		},
	)
	if err != nil {
		return "", errors.Wrapf(err, "failed to generate totp passcode for %s", name)
	}

	return passcode, nil
}
