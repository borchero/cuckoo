package cmd

import (
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/spf13/cobra"
	"go.borchero.com/cuckoo/providers"
	"go.borchero.com/cuckoo/utils"
	"go.borchero.com/typewriter"
)

const decryptDescription = `
The decrypt command recursively scans the current directory and all of its subdirectories to
decrypt all files matching a particular pattern using Mozilla's Sops. This way, secrets can be
stored within the repository and be made available easily for the CI.
`

var decryptArgs struct {
	inputPattern string
	output       string
}

func init() {
	decryptCommand := &cobra.Command{
		Use:   "decrypt",
		Short: "Decrypt all files in the current directory and its subdirectories using Sops.",
		Long:  decryptDescription,
		Args:  cobra.ExactArgs(0),
		Run:   runDecrypt,
	}

	decryptCommand.Flags().StringVarP(
		&decryptArgs.inputPattern, "decrypt", "d", "^(.*)\\.enc\\.(.*)$",
		"The regex to apply for matching files to decrypt. May define capture groups.",
	)
	decryptCommand.Flags().StringVarP(
		&decryptArgs.output, "output", "o", "%s.%s",
		"The format string to use for writing decrypted files. Uses the capture groups from -d.",
	)

	rootCmd.AddCommand(decryptCommand)
}

func runDecrypt(cmd *cobra.Command, args []string) {
	logger := typewriter.NewCLILogger()

	// 1) Get all files matching pattern
	matches, err := utils.GetMatchingFiles(decryptArgs.inputPattern, ".")
	if err != nil {
		typewriter.Fail(logger, "Could not find any files", err)
	}

	// 2) Decrypt all files
	sops := providers.NewSops()
	for _, file := range matches {
		// 2.1) Decrypt file
		logger.Infof("Decrypting '%s'...", file)
		contents, err := sops.Decrypt(file)
		if err != nil {
			typewriter.Fail(logger, "Could not decrypt file", err)
			return
		}

		// 2.2) Write decrypted contents
		inputRegex := regexp.MustCompile(decryptArgs.inputPattern)
		matches := inputRegex.FindStringSubmatch(file)

		stringArgs := make([]interface{}, len(matches)-1)
		for i, m := range matches[1:] {
			stringArgs[i] = m
		}

		newFile := fmt.Sprintf(decryptArgs.output, stringArgs...)
		err = ioutil.WriteFile(newFile, contents, 0644)
		if err != nil {
			typewriter.Fail(logger, "Could not write decrypted file", err)
			return
		}
	}

	logger.Success("Done 🎉")
}
