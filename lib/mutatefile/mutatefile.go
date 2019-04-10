package mutatefile

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
)

const (
	TemporaryNewFileSuffix = ".generated.new"
)

func MutateFile(filename string, perm os.FileMode, mutator func([]byte) ([]byte, error)) error {
	if filename == "" {
		return fmt.Errorf("No filename provided")
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			data, err = mutator(nil)
			if err != nil {
				return fmt.Errorf("Error creating data for %q: %v", filename, err)
			}
			f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, perm)
			if err != nil {
				return fmt.Errorf("Error opening %q for writing: %v", filename, err)
			}
			if _, err := f.Write(data); err != nil {
				return fmt.Errorf("Error writing to %q: %v", filename, err)
			}
			if err := f.Close(); err != nil {
				return fmt.Errorf("Error closing %q after writing: %v", filename, err)
			}
			return nil
		}
		return fmt.Errorf("Error reading file %q: %v", filename, err)
	}

	data, err = mutator(data)
	if err != nil {
		return fmt.Errorf("Error updating data for %q: %v", filename, err)
	}

	newFileTemporaryPath := filename + TemporaryNewFileSuffix
	temporaryFile, err := os.OpenFile(newFileTemporaryPath, os.O_CREATE|os.O_WRONLY, perm)
	if err != nil {
		return fmt.Errorf("Error opening %q for writing: %v", newFileTemporaryPath, err)
	}
	defer func() {
		if err := os.Remove(newFileTemporaryPath); err != nil {
			if !os.IsNotExist(err) {
				logrus.Errorf("Failed to remove %q: %v", newFileTemporaryPath)
			}
		}
	}()

	if _, err := temporaryFile.Write(data); err != nil {
		return fmt.Errorf("Error writing to %q: %v", newFileTemporaryPath, err)
	}

	if err := temporaryFile.Close(); err != nil {
		return fmt.Errorf("Error closing %q: %v", newFileTemporaryPath, err)
	}

	if err := os.Rename(newFileTemporaryPath, filename); err != nil {
		return fmt.Errorf("Error renaming %q to %q: %v", newFileTemporaryPath, filename, err)
	}

	return nil
}
