package cmd

import "fmt"

// updateCmd is disabled — self-update via download has been removed for security.
// To update unity-cli, build from source.
func updateCmd(args []string) error {
	_ = parseSubFlags(args)
	return fmt.Errorf("self-update is disabled for security. Build from source to update.")
}
