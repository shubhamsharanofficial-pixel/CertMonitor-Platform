package assets

import (
	_ "embed"
)

//go:embed install.sh
var InstallScript string

// GetInstallScript returns the raw content of the script
func GetInstallScript() string {
	return InstallScript
}
