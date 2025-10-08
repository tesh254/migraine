# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v2.0.0] - 2025-01-08

### Added
- **Complete storage migration from Badger to SQLite** - All data now stored in normalized SQLite tables
- **YAML workflow support** - Workflows can now be defined as YAML files with comments and structured format
- **Vaulted variables system** - Secure variable storage with scope awareness (global, project, workflow)
- **Enhanced CLI commands**:
  - `workflow init [name]` - Scaffold new YAML workflows with commented sections
  - `workflow validate [path]` - Validate workflow files
  - `workflow run [name]` - Execute workflows with variable resolution
  - `vars set/get/list/delete` - Manage vault variables with scope flags
- **Workflow discovery from current directory** - Automatically find YAML workflows in ./workflows/
- **Migration utilities** - Automatic migration of existing Badger data to SQLite

### Changed
- **Breaking**: Storage system changed from Badger to SQLite
- **Breaking**: Workflow structure updated to support YAML format and vault variables
- **Enhanced**: Variable resolution now supports vault fallback with scope precedence
- **Improved**: CLI command structure with clearer subcommands

### Deprecated
- Legacy JSON template system (still supported for migration)

### Removed
- BadgerDB dependency

### Fixed
- Improved error handling for variable resolution
- Better workflow execution with proper variable substitution

## [v1.0.6] - 2025-06-19

### Added
- `v1.0.6` [2025-06-19] introduce fang usage
- `v1.0.5` [2025-04-30] ensure store exists before read operations
- `v1.0.4` [2025-04-30] ensure store exists before read operations
- `v1.0.3` [2025-04-17] terminal output restyling
- `v1.0.0` [2025-04-10] stabilize read and write operations
- `v0.1.7` [2025-03-25] restructure project files
- `v0.1.6` [2025-03-25] emulate default terminal behavior and environment when executing workflows
- `v0.1.5` [2025-03-25] support concurrent execution of workflows
- `v0.1.4` [2025-03-25] storage service upgrade
- `v0.1.3` [2025-03-25] add storage service to support read-only mode in operations
- `v0.1.2` [2025-03-25] minor bug fixes
- `v0.1.1` [2025-03-25] minor bug fixes
- `v0.1.0`  [2024-11-24] add `info` command for workflow details
- `v0.0.12` [2024-11-22] update shorthand command to mgr over mig
- `v0.0.11` [2024-11-21] use default user shell over strictly using bash
- `v0.0.9`  [2024-11-21] add ability to load templates from url
- `v0.0.7`  [2024-11-21] remove npm publish, update installation docs
- `v0.0.3`  [2024-11-21] add workflow features
