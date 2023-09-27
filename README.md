![___migraine_cover___](https://github.com/tesh254/migraine/assets/31078302/4c5b4aca-bdc4-4a41-9287-69df27250c26)


# `migraine`: A CLI for managing migrations in backend projects with PostgreSQL

`migraine` is a command-line interface (CLI) tool designed to help you manage migrations in your backend project using a PostgreSQL database.

---

## Table of Contents

- [Installation](#installation)
  - [Using Homebrew](#homebrew)
- [Prerequisites](#prerequisites)
- [Commands](#commands)
  - [Initialize migraine](#initialize-migraine)
  - [Create a new migration](#create-a-new-migration)
  - [Run migrations](#run-migrations)
  - [Rollback](#rollback)
  - [Help and Version](#help-and-version)
- [Migrations](#migrations)
  - [Writing migrations](#writing-migrations)

---

### Installation
System Requirements:
- macOS(Intel/Apple Silicon) and Linux supported

#### Homebrew
> We recommend using Homebrew on macOS/linux{distro that supports it}
If you don't have Homebrew installed on your macOS/linux use the command below to install it

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

1. Clone the Homebrew tap

```bash
brew tap tesh254/migraine https://github.com/tesh254/homebrew-migraine
```
2. Install `migraine`

```bash
brew install migraine
```

To update the package to a new version use

```bash
brew update

brew upgrade migraine
```
### Prerequisites

- `migraine` currently supports only PostgreSQL databases.
- While Migraine doesn't generate SQL queries for you, a planned feature will allow AI-generated SQL queries based on a `prompt` flag.
- `migraine` currently works with PostgreSQL connection strings in your `.env` file, but a feature is in development to use more credential-specific variables as flags or fetch them from a vault.

### Commands

#### Initialize Migraine

Initialize Migraine, creating a migrations folder and a migrations table in your database. By default, it uses `.env` as your environment file and the database environment variable as `DATABASE_URL`.

```bash
migraine --init
```

Example with a custom environment file name:

```bash
migraine --init --env ".env.local"
```

Example with a custom database URL environment variable:

```bash
migraine --init --dbVar "DATABASE_URL"
```

Example with both custom environment file and database URL environment variable:

```bash
migraine --init --env ".env.local" --dbVar "DATABASE_URL"
```

#### Create a New Migration

Create a new migration file to house your SQL code for execution in the database.

```bash
migraine --migration --new "<migration_name>"
```

Example:

```bash
migraine --migration --new "create_user_table"
```

#### Run Migrations

Run all your migrations, skipping those that have already been executed.

```bash
migraine --migration --run
```

#### Rollback

Rollback the most recent migration. Use this option cautiously, especially if there are foreign key constraints, as it may fail.

```bash
migraine --rollback
```

#### Help and Version

Display Migraine's usage and current version.

```bash
migraine --help
```

```bash
migraine --version
```

### Migrations

When you run `migraine --init`, a `./migrations` folder is created, along with the `_migraine_migrations` table to store all your migrations in one place for tracking.

**Note**: After creating and running migrations with `migraine --migration --run`, it is recommended not to delete or modify any `.sql` file within the `./migrations` folder. Migraine relies on the chronology of each file to determine the execution order. Similarly, do not modify the `migrations` table created in your database.

#### Writing Migrations

After running `migraine --init` successfully, you can create a new migration file by running:

```bash
migraine --migration --new "<migration_name>"
```

You can name your migration in formats like `"create user table"` or `"create_user_table"` (don't forget to use quotes for `migraine` to recognize it as a migration name).

A new file with the migration name will be created in the migrations folder in this format: `<unix_time>_<migration_name_formatted>.sql`.

If you open the file for editing, it should have the following structure:

```sql
--migraine-up


--migraine-down

```

- `--migraine-up`: contains SQL that makes changes to the database.
- `--migraine-down`: contains SQL that rolls back the changes made by `--migraine-up`.

Place your SQL queries immediately after these comments to help `migraine` distinguish between making changes and rolling them back.

**Example**:

```sql
--migraine-up
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL
);

--migraine-down
DROP TABLE users;
```

### Future Major Improvements

- Adding SQL generation through AI
- Postgresql monitoring features
