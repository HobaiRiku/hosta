# Hosta

Hosta is a fast, interactive SSH host launcher built on top of OpenSSH.

It discovers hosts from your existing SSH config, provides fuzzy search and shell completion, and delegates every connection to the system `ssh` client.

## Features

- Interactive host search and selection
- Native OpenSSH config and `Include` support
- Direct `list`, `show`, and `connect` commands
- Bash, Zsh, Fish, and PowerShell completion
- Linux, macOS, and Windows support

## Install

Go 1.25 or later is required:

```bash
go install github.com/HobaiRiku/hosta/cmd/hosta@latest
```

Make sure `$(go env GOPATH)/bin` is in your `PATH`.

## Usage

```bash
hosta                  # Open the interactive launcher
hosta list             # List discovered hosts
hosta show home        # Show the resolved host configuration
hosta connect home     # Connect with OpenSSH
hosta doctor           # Check the local setup
hosta version          # Print build information
```

Use another SSH config with `--config`:

```bash
hosta --config /path/to/ssh_config list
```

Install or update shell completion with:

```bash
hosta completion zsh
```

This writes `~/.hosta_completion_zsh` and adds an idempotent source block to `~/.zshrc`. Use `--stdout` to print the script instead.

## Development

```bash
git clone git@github.com:HobaiRiku/hosta.git
cd hosta
make check
make build
./build/bin/hosta version
```

Run from source with `make run`. Use `make print-version` to inspect the build metadata that will be injected into the binary.

## License

[MIT](LICENSE)
