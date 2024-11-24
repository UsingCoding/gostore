//go:build !windows

package storage

import (
	"fmt"
	"os"
	"syscall"
)

func castToSysStat(info os.FileInfo) (sysStat, error) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return sysStat{}, fmt.Errorf("failed to get raw syscall.Stat_t data for '%s'", info.Name())
	}

	return sysStat{
		uid: stat.Uid,
		gid: stat.Gid,
	}, nil
}
