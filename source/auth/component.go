package auth

import (
	"fmt"
	"os/user"
)

// Component describes a service that requires authentication.
type Component interface {

	// Name returns the full name of the component to print information about it.
	Name() string

	// EnsureAccess checks for required environment variables, possibly performs authentication
	// and returns an error if any operation fails.
	EnsureAccess() error
}

func home() string {
	usr, err := user.Current()
	if err != nil {
		panic(fmt.Sprintf("Cannot get home directory: %s", err))
	}
	return usr.HomeDir
}
