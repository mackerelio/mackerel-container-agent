package cgroupfs

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type fsFile struct {
	Path    string
	Mode    os.FileMode
	Content string
}

type fsDirectory struct {
	Path string
	Mode os.FileMode
}

type fsSymLink struct {
	Path string
	To   string
}

func mockFilesystem(items []interface{}, dir, prefix string) (string, error) {
	root, _ := ioutil.TempDir(dir, prefix)

	for _, item := range items {
		switch v := item.(type) {
		case fsFile:
			ioutil.WriteFile(filepath.Join(root, v.Path), []byte(v.Content), v.Mode)
		case fsDirectory:
			os.MkdirAll(filepath.Join(root, v.Path), v.Mode)
		case fsSymLink:
			os.Symlink(v.To, filepath.Join(root, v.Path))
		}
	}

	return root, nil
}
