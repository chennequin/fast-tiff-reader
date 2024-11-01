package tiffio

import "bytes"

func readCString(buf []byte) (string, int, error) {
	n := bytes.IndexByte(buf, 0)
	if n == -1 {
		return string(buf), len(buf), ZeroNotFound
	}
	return string(buf[:n]), n, nil
}
