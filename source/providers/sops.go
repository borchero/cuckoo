package providers

import (
	"strings"

	"go.mozilla.org/sops/v3/decrypt"
)

// Sops decrypts files with Mozilla's sops tool.
type Sops struct {
}

// NewSops returns a new Sops instance to decrypt files.
func NewSops() *Sops {
	return &Sops{}
}

// Decrypt decrypts the file at the given directory and returns its contents.
func (sops *Sops) Decrypt(file string) ([]byte, error) {
	// 1) Get extension
	splits := strings.Split(file, ".")
	extension := splits[len(splits)-1]

	// // 2) Decrypt
	if extension == "env" {
		return decrypt.File(file, "dotenv")
	} else if extension == "json" || extension == "yaml" {
		return decrypt.File(file, extension)
	} else {
		return decrypt.File(file, "binary")
	}
}
