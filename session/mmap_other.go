//go:build !unix

package session

import (
	"errors"
	"os"
)

func mmapFile(f *os.File, size int) ([]byte, error) {
	return nil, errors.New("mmap not supported on this platform")
}

func munmapFile(data []byte) error {
	return nil
}
