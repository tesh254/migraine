package constants

import (
	"fmt"
	"runtime"

	"github.com/tesh254/migraine/utils"
)

const VERSION string = "v0.0.2-alpha.5"
const CONFIG string = ".migraine.config.json"
const MIGRATION_CONTENT = `--migraine-up


--migraine-down

`
const MIGRAINE_UP_MARKER = "migraine-up"
const MIGRAINE_DOWN_MARKER = "migraine-down"
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
const MIGRAINE_USAGE = `
migraine [options]

migraine --help
	Show migraine usage

migraine --version
	Show migraine's current version
`

func CurrentOSWithVersion() string {
	operatingSystem := runtime.GOOS

	return fmt.Sprintf(`%smigraine%s %s running on %s%s%s os`, utils.BOLD, utils.RESET, VERSION, utils.BOLD, operatingSystem, utils.RESET)
}
