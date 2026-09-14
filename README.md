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

### macOS (Homebrew)

```bash
brew install HobaiRiku/tap/hosta
```

### Prebuilt binaries

For Linux, macOS, and Windows, download the archive matching your OS and CPU
architecture from the [latest release](https://github.com/HobaiRiku/hosta/releases/latest), extract it, and put the `hosta` binary on your `PATH`.

For example, on Linux:

```bash
tar -xzf hosta_*_linux_*.tar.gz
sudo install -m 755 hosta /usr/local/bin/hosta
```

On Windows, run the following in PowerShell. It downloads the latest x64 release,
installs `hosta.exe` in your user-local programs directory, and adds that directory
to your user `PATH`:

```powershell
$release = Invoke-RestMethod https://api.github.com/repos/HobaiRiku/hosta/releases/latest
$asset = $release.assets | Where-Object { $_.name -match '_windows_amd64\.zip$' } | Select-Object -First 1
if (-not $asset) { throw 'No Windows x64 archive found in the latest release.' }

$installDir = Join-Path $env:LOCALAPPDATA 'Programs\hosta'
$archive = Join-Path $env:TEMP $asset.name
Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $archive
New-Item -ItemType Directory -Force -Path $installDir | Out-Null
Expand-Archive -Path $archive -DestinationPath $installDir -Force

$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if ($userPath -notlike "*$installDir*") {
  [Environment]::SetEnvironmentVariable('Path', "$userPath;$installDir", 'User')
}
```

Open a new PowerShell window, then run `hosta version` to verify the installation.

### Install from source

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

In the interactive launcher, press `y` to copy the selected host's resolved `ssh -p <port> <user>@<host>` command.

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
