package config

import (
	"context"
	"encoding/json"
	"os"
	"path"

	"github.com/UsingCoding/fpgo/pkg/slices"
	"github.com/pkg/errors"

	"github.com/UsingCoding/gostore/internal/common/maybe"
	appconfig "github.com/UsingCoding/gostore/internal/gostore/app/config"
	"github.com/UsingCoding/gostore/internal/gostore/app/encryption"
	"github.com/UsingCoding/gostore/internal/gostore/app/vars"
)

const (
	configName = "config.json"
)

func NewStorage(configDir string) appconfig.Storage {
	return &storage{configDir: configDir}
}

type storage struct {
	configDir string
}

func (s *storage) Load(context.Context) (appconfig.Config, error) {
	data, err := os.ReadFile(path.Join(s.configDir, configName))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return appconfig.Config{}, appconfig.ErrConfigNotFound
		}
		return appconfig.Config{}, errors.Wrap(err, "failed to read config file")
	}

	var c config
	err = json.Unmarshal(data, &c)
	if err != nil {
		return appconfig.Config{}, errors.Wrap(err, "failed to unmarshal config")
	}

	if c.Kind != string(vars.ConfigKind) {
		return appconfig.Config{}, errors.Errorf("unknown kind %s for config", c.Kind)
	}

	return appconfig.Config{
		Context: maybe.Map(c.Context, func(c string) appconfig.StoreID {
			return appconfig.StoreID(c)
		}),
		Stores: slices.Map(c.Stores, func(s store) appconfig.Store {
			return appconfig.Store{
				ID:   appconfig.StoreID(s.ID),
				Path: s.Path,
			}
		}),
		Identities: slices.Map(c.Identities, func(i identity) encryption.Identity {
			return encryption.Identity{
				Recipient:  encryption.Recipient(i.Recipient),
				PrivateKey: encryption.PrivateKey(i.PrivateKey),
			}
		}),
	}, nil
}

func (s *storage) Store(_ context.Context, c appconfig.Config) error {
	if e, err := exists(s.configDir); !e || err != nil {
		if err != nil {
			return errors.Wrapf(err, "failed to check that folder for config at %s exists", s.configDir)
		}

		err = os.MkdirAll(s.configDir, os.ModePerm)
		if err != nil {
			return errors.Wrapf(err, "failed to create folder for config at %s", s.configDir)
		}
	}

	data, err := json.Marshal(config{
		Kind: string(vars.ConfigKind),
		Context: maybe.Map(c.Context, func(c appconfig.StoreID) string {
			return string(c)
		}),
		Stores: slices.Map(c.Stores, func(s appconfig.Store) store {
			return store{
				ID:   string(s.ID),
				Path: s.Path,
			}
		}),
		Identities: slices.Map(c.Identities, func(i encryption.Identity) identity {
			return identity{
				Recipient:  string(i.Recipient),
				PrivateKey: string(i.PrivateKey),
			}
		}),
	})
	if err != nil {
		return errors.Wrap(err, "failed to marshal config")
	}

	p := path.Join(s.configDir, configName)

	err = os.WriteFile(p, data, 0644)
	return errors.Wrapf(err, "failed to save config to file %s", p)
}

type config struct {
	Kind       string              `json:"kind"`
	Context    maybe.Maybe[string] `json:"context"`
	Stores     []store             `json:"stores"`
	Identities []identity          `json:"identities"`
}

type store struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

type identity struct {
	Recipient  string `json:"recipient"`
	PrivateKey string `json:"privateKey"`
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
