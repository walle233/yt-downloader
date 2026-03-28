package service

import "os"

func removeAll(path string) error {
	return os.RemoveAll(path)
}
