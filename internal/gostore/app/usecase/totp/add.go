package totp

import (
	"context"

	"github.com/pkg/errors"
	"github.com/pquerna/otp"

	"github.com/UsingCoding/gostore/internal/common/mapper"
	"github.com/UsingCoding/gostore/internal/common/maybe"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
)

type Algorithm string

const (
	AlgorithmSHA1   Algorithm = "SHA1"
	AlgorithmSHA256 Algorithm = "SHA256"
	AlgorithmSHA512 Algorithm = "SHA512"
	AlgorithmMD5    Algorithm = "MD5"
)

var (
	alg = mapper.New(map[Algorithm]otp.Algorithm{
		AlgorithmSHA1:   otp.AlgorithmSHA1,
		AlgorithmSHA256: otp.AlgorithmSHA256,
		AlgorithmSHA512: otp.AlgorithmSHA512,
		AlgorithmMD5:    otp.AlgorithmMD5,
	})
)

type AddParams struct {
	Name      string
	Secret    []byte
	Algorithm Algorithm
}

func (s service) AddIssuer(ctx context.Context, params AddParams) error {
	_, ok := alg.L()[params.Algorithm]
	if !ok {
		return errors.Errorf("unknown algorithm: %s", params.Algorithm)
	}

	// store secret
	err := s.service.Add(ctx, store.AddParams{
		SecretIndex: makeTOTPIndex(store.SecretIndex{
			Path: params.Name,
			Key:  maybe.NewJust(secretKey),
		}),
		Data: params.Secret,
	})
	if err != nil {
		return errors.Wrap(err, "failed to add totp secret")
	}

	// store alg
	err = s.service.Add(ctx, store.AddParams{
		SecretIndex: makeTOTPIndex(store.SecretIndex{
			Path: params.Name,
			Key:  maybe.NewJust(algKey),
		}),
		Data: []byte(params.Algorithm),
	})
	if err != nil {
		return errors.Wrap(err, "failed to add totp algorithm")
	}

	return nil
}
