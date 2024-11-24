package storage

import (
	"os"

	"github.com/pkg/errors"
)

func castToSysStat(info os.FileInfo) (sysStat, error) {
	return sysStat{}, errors.New("sysStat: not implemented on windows")
}
