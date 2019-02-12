package cgroupfs

import (
	"io"
	"io/ioutil"
	"strconv"
	"strings"
)

func readInt(data io.Reader) (int64, error) {
	b, err := ioutil.ReadAll(data)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64)
}

func readUint(data io.Reader) (uint64, error) {
	b, err := ioutil.ReadAll(data)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(strings.TrimSpace(string(b)), 10, 64)
}
