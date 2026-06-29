//go:build unix

package session

import (
	"os"

	"golang.org/x/sys/unix"
)

func mmapFile(f *os.File, size int) ([]byte, error) {
	data, err := unix.Mmap(int(f.Fd()), 0, size, unix.PROT_READ, unix.MAP_PRIVATE)
	if err != nil {
		return nil, err
	}
	_ = unix.Madvise(data, unix.MADV_SEQUENTIAL)
	return data, nil
}

func munmapFile(data []byte) error {
	return unix.Munmap(data)
}
