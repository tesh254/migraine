package constants

import (
	"fmt"
	"runtime"

	"github.com/tesh254/migraine/internal/version"
	"github.com/tesh254/migraine/pkg/utils"
)

const MIGRAINE_ASCII = `

               /$$                              /$$
              |__/                             |__/
 /$$$$$$/$$$$  /$$  /$$$$$$   /$$$$$$  /$$$$$$  /$$ /$$$$$$$   /$$$$$$
| $$_  $$_  $$| $$ /$$__  $$ /$$__  $$|____  $$| $$| $$__  $$ /$$__  $$
| $$ \ $$ \ $$| $$| $$  \ $$| $$  \__/ /$$$$$$$| $$| $$  \ $$| $$$$$$$$
| $$ | $$ | $$| $$| $$  | $$| $$      /$$__  $$| $$| $$  | $$| $$_____/
| $$ | $$ | $$| $$|  $$$$$$$| $$     |  $$$$$$$| $$| $$  | $$|  $$$$$$$
|__/ |__/ |__/|__/ \____  $$|__/      \_______/|__/|__/  |__/ \_______/
                   /$$  \ $$
                  |  $$$$$$/
                   \______/

`

const MIGRAINE_ASCII_V2 = `
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
@@@@@@@@@@@%#%@@##@@@@@@%###*+- %@@@@@@@
@@@@@@@=  :   :    %@%-   .   -+%@@@@@@@
@@@@@@@=  -%. :%-  *@:  :@@@.  *@@@@@@@@
@@@@@@@=  +@: -@=  *@=   *%*   *@@@@@@@@
@@@@@@@=  +@: -@=  *@@+  . .:=#@@@@@@@@@
@@@@@@@=  +@: -@=  *@@   =+++*#@@@@@@@@@
@@@@@@@*::*@=:+@*::#@@#-::::.  .%@@@@@@@
@@@@@@@@@@@@@@@@@@@@#::-%@@@#   +@@@@@@@
@@@@@@@@@@@@@@@@@@@@@=.       :+@@@@@@@@
@@@@@@@@@@@@@@@@@@@@@@@@%%%%@@@@@@@@@@@@
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
`

// VERSION returns the current version from build info
func VERSION() string {
	return version.GetVersion()
}

// VERSION_WITH_COMMIT returns version with commit information
func VERSION_WITH_COMMIT() string {
	return version.GetVersionWithCommit()
}

// SHORT_VERSION returns a concise version string
func SHORT_VERSION() string {
	return version.GetShortVersion()
}

// DETAILED_VERSION returns comprehensive version information
func DETAILED_VERSION() string {
	return version.GetDetailedVersion()
}

// BUILD_SUMMARY returns a one-line build summary
func BUILD_SUMMARY() string {
	return version.GetBuildSummary()
}

// CurrentOSWithVersion returns OS info with version
func CurrentOSWithVersion() string {
	operatingSystem := runtime.GOOS
	versionInfo := VERSION_WITH_COMMIT()

	buildType := "release"
	if version.IsDevelopment() {
		buildType = "development"
	}

	return fmt.Sprintf(`%smigraine%s %s (%s build) running on %s%s%s`,
		utils.BOLD, utils.RESET, versionInfo, buildType, utils.BOLD, operatingSystem, utils.RESET)
}

// GetReleaseInfo returns release-specific information
func GetReleaseInfo() string {
	if version.IsRelease() {
		return fmt.Sprintf("Release build %s", VERSION())
	}
	return fmt.Sprintf("Development build %s", BUILD_SUMMARY())
}
