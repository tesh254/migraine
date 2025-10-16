![migraine](https://github.com/user-attachments/assets/1f1f90d0-3a85-44c8-b84a-b23838bf35c2)

[![migraine release](https://github.com/tesh254/migraine/actions/workflows/release.yml/badge.svg)](https://github.com/tesh254/migraine/actions/workflows/release.yml)

# `migraine`

This is a robust CLI tool used to organize and automate complex workflows with templated commands. Users can define, store, and run sequences of shell commands efficiently, featuring variable substitution, pre-flight checks, and discrete actions.

> We recommend reading the docs details on (migraine docs)[https://migraine.wchr.xyz]. They will always be up to date but we will update the readme soon on basic info

### Security Notice
⚠️ **IMPORTANT**: The vault currently stores variables in an unencrypted SQLite database. While variables are stored locally, they are not encrypted at rest. We are actively working on adding encryption support in an upcoming release. For now, we recommend avoiding storing highly sensitive information like production API keys in the vault until encryption is implemented.

For any other issues, please check our [issue tracker](https://github.com/tesh254/migraine/issues) or submit a new issue.
