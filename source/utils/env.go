package utils

import (
	"fmt"
	"os"
)

// MapEnvs sets the environment variables defined by the keys to the environment variables given by
// the values.
func MapEnvs(mapping map[string]string) {
	for source, target := range mapping {
		if os.Getenv(source) == "" {
			continue
		}
		if err := os.Setenv(target, os.Getenv(source)); err != nil {
			panic(fmt.Sprintf("Unknown error: %s", err))
		}
	}
}
