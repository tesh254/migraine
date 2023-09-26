package constants

import (
	"fmt"
	"runtime"

	"github.com/tesh254/migraine/utils"
)

const VERSION string = "v0.0.1-alpha.1"
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
const MIGRAINE_USAGE = `
migraine [options]

migraine --init
    	Initialize migraine (creates migrations folder and migrations table in your database). Defaults to use .env as your environment
    	file and database environment variable as DATABASE_URL.

migraine --init --env <environment_file_name>
	Example: migraine --init --env ".env.local"
	Initialize migraine (creates migrations folder and migrations table in your database). Defaults to use database environment 
	variable as DATABASE_URL.

migraine --init --dbVar <database_url_environment_variable>
	Example: migraine --init --dbVar "DATABASE_URL"
	Initialize migraine (creates migrations folder and migrations table in your database). Defaults to use .env as your environment
	file.

migraine --init --env <environment_file_name> --dbVar <database_url_environment_variable>
	Example: migraine --init --env ".env.local" --dbVar "DATABASE_URL"
	Initialize migraine (creates and migrations folder and migrations table in your database). Uses the values provided in the 
	option flags as your environment file and database url variable

migraine --migration --new <migration_name>
	Example: migraine --migration --new "create user table"
	Create migrations file to house your sql code to execute to the database

migraine --migration --run
	Runs all your migrations, skips already ran migrations

migraine --rollback
	Rollsback the recent migration run
	Note: use this option wisely because if there are foreign key constraints this command will fail. More features will be added
	to help with such situations example giving a number to rollback from the recent migration or nitpicking

migraine --help
	Show migraine usage

migraine --version
	Show migraine's current version
`

func CurrentOSWithVersion() string {
	operatingSystem := runtime.GOOS

	return fmt.Sprintf(`%smigraine%s %s running on %s%s%s os`, utils.BOLD, utils.RESET, VERSION, utils.BOLD, operatingSystem, utils.RESET)
}
