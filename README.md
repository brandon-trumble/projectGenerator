# Project Generator

A fast, interactive CLI tool to scaffold Go projects with boilerplate code and best practices. Spin up new HTTP backends, CLI applications, or empty Go projects in seconds.

## Overview

Project Generator is a Go-based project scaffolding tool that removes the friction of starting new projects. Answer a few interactive questions, and get a fully structured project with:

- Pre-configured project layouts
- Optional database integrations (PostgreSQL, MySQL)
- Test case templates for HTTP backends
- Multiple CLI frameworks (Cobra, Bubble Tea, Huh)
- Cross-platform support (Windows, macOS, Linux)

Perfect for rapidly prototyping, learning Go, or maintaining consistent project structure across your team.

## Features

**Interactive Setup**
- Beautiful terminal UI with form validation
- Step-by-step project configuration
- Smart conditional prompts (database options only for HTTP backends)

**Project Templates**
- HTTP Backend - RESTful API with optional database and tests
- CLI Application - Command-line tools with multiple framework options
- Empty Project - Minimal Go project structure

**Database Support**
- PostgreSQL (PGX driver)
- MySQL
- None (for non-database projects)

**Testing**
- Optional test case templates for HTTP backends
- Pre-configured testing structure

**CLI Frameworks**
- Cobra - Standard flag-based CLI applications
- Bubble Tea - Interactive TUI (Text User Interface) applications
- Huh - Beautiful form-based CLI tools

## Installation

Two routes: grab a prebuilt binary from the releases page, or build it yourself.
Either way you finish by running the binary once, which sets up your PATH.

### From a release (no Go needed)

Prebuilt archives are published for every tagged version by
[GoReleaser](https://goreleaser.com):

| Platform | Architecture | Archive |
| --- | --- | --- |
| Linux | x86_64 | `project-gen_<version>_linux_amd64.tar.gz` |
| Linux | ARM64 | `project-gen_<version>_linux_arm64.tar.gz` |
| macOS | Intel | `project-gen_<version>_darwin_amd64.tar.gz` |
| macOS | Apple Silicon | `project-gen_<version>_darwin_arm64.tar.gz` |
| Windows | x86_64 | `project-gen_<version>_windows_amd64.zip` |
| Windows | ARM64 | `project-gen_<version>_windows_arm64.zip` |

Download the one for your machine from the
[releases page](https://github.com/bran7230/projectGenerator/releases), then:

```bash
# Linux / macOS
tar -xzf project-gen_<version>_linux_amd64.tar.gz
./project-gen
```

```powershell
# Windows: extract the .zip, then from that folder
.\project-gen.exe
```

The binaries are statically linked with `CGO_ENABLED=0`, so there is no runtime
to install — not even Go. On the first run you are asked whether to install it
onto your PATH; see [Setting up PATH](#setting-up-path).

Every release also ships a `checksums.txt`. To verify your download before
running it:

```bash
sha256sum -c checksums.txt --ignore-missing
```

macOS may quarantine a downloaded binary. If Gatekeeper blocks it:

```bash
xattr -d com.apple.quarantine project-gen
```

### Prerequisites (building from source)
- Go 1.26 or higher

### From Source

```bash
git clone https://github.com/bran7230/projectGenerator.git
cd projectGenerator
go build -o project-gen main.go
```

### Or run the makefile for path setup

```bash
git clone https://github.com/bran7230/projectGenerator.git
cd projectGenerator
make install
```

`make install` builds the binary and immediately runs its installer. Whichever
route you take, the command you end up typing is `project-gen`.

## Setting up PATH

There is nothing to configure by hand. Run the binary once and it offers to
install itself:

```
┃ Install project-gen so you can run it from anywhere?
┃ Copies this binary to $HOME/.local/bin and adds it to your PATH.
┃
┃  Yes    Not now
```

Answer **Yes** and it copies itself to a per-user bin directory, then makes sure
that directory is on your PATH:

| Platform | Installed to | PATH updated in |
| --- | --- | --- |
| Linux / macOS (bash) | `~/.local/bin` | `~/.bashrc` and `~/.profile` |
| Linux / macOS (zsh) | `~/.local/bin` | `~/.zshrc` and `~/.profile` |
| Linux / macOS (fish) | `~/.local/bin` | `~/.config/fish/config.fish` |
| Windows | `%LOCALAPPDATA%\Programs\project-gen` | the per-user `Path` variable |

Your shell is detected from `$SHELL`. Both locations are inside your own home
directory, so **no `sudo` and no admin rights are needed**, and nothing outside
your user account is touched.

### One catch: open a new terminal

A program cannot change the PATH of the shell that launched it — that is an OS
boundary, not a limitation of this tool. The installer writes the change to disk,
but your current shell has already read its config. So after installing:

```bash
# open a new terminal, then from any directory:
project-gen
```

Or reload the current shell instead of opening a new one:

```bash
source ~/.bashrc
```

### What the installer actually does

1. Resolves the path of the running binary, following symlinks.
2. Checks whether it is already installed. If so it prints
   `Already installed` and skips everything below.
3. Asks for confirmation. Answering **Not now** records your choice in
   `~/.config/project-gen/skip-install` so you are only asked once.
4. Copies the binary to the bin directory with mode `0755`, writing to a
   temporary file first and renaming it into place. That way reinstalling over a
   copy that is currently running cannot fail or corrupt the binary.
5. Adds the bin directory to PATH — but only if it is not on your PATH already.
   On many Linux setups `~/.local/bin` is preconfigured, in which case no config
   file is touched at all.
6. Prints exactly which files it changed.

On Linux and macOS the appended block is tagged so you can find it later, and so
repeat installs never duplicate it:

```bash
# added by project-gen
export PATH="$HOME/.local/bin:$PATH"
```

On Windows the per-user `Path` environment variable is updated through
PowerShell. The system-wide `Path` is left alone.

### Installing later

If you answered **Not now**, or you want to reinstall after rebuilding, run the
installer explicitly at any time:

```bash
./project-gen --install
```

This ignores the earlier "not now" and reruns the whole process.

### Uninstalling

There is no uninstall command; removal is two manual steps.

```bash
# 1. delete the binary
rm ~/.local/bin/project-gen

# 2. delete the "# added by project-gen" block from your shell config
#    (~/.bashrc, ~/.zshrc, ~/.profile, or ~/.config/fish/config.fish)
```

On Windows, delete `%LOCALAPPDATA%\Programs\project-gen` and remove that path
from your user `Path` under *System Properties → Environment Variables*.

### Troubleshooting

**`project-gen: command not found` after installing** — you are still in the old
shell. Open a new terminal, or `source` your shell config as shown above.

**Confirm the directory made it onto PATH:**

```bash
echo $PATH | tr ':' '\n' | grep local/bin
```

**The install prompt never appears** — that is expected in three cases: it is
already installed, you previously answered "Not now", or you launched it with
`go run` (there is no permanent binary to install). Use `--install` to force it.

## Usage

Run the generator:
```bash
./project-gen
# or if installed globally
project-gen
```

### Flags

| Flag | What it does |
| --- | --- |
| *(none)* | Runs the interactive generator |
| `--install` | Installs the binary onto your PATH, then exits |
| `--version` | Prints the version, commit and build date, then exits |

Both flags also accept the `-install` and `install` spellings.

### Interactive Prompts

1. **Project Name** - Enter a name for your new project
2. **Project Type** - Choose between:
   - Http-Backend
   - Cli-Application
   - Empty-Project
3. **Database** (HTTP Backend only) - Select database framework
4. **Test Cases** (HTTP Backend only) - Include test templates
5. **CLI Framework** (CLI Application only) - Select your CLI tool

### Example Workflow

```
$ project-gen

     ____             _           __          ______
    / __ \_________  (_)__  _____/ /_        / ____/___  ____
   / /_/ / ___/ __ \/ / _ \/ ___/ __/_______/ / __/ __ \/ __ \
  / ____/ /  / /_/ / /  __/ /__/ /_/_______/ /_/ /  __/ / / /
 /_/   /_/   \____/ /\___/\___/\__/        \____/\___/_/ /_/
              /___/

Enter project name: my-awesome-api
Confirm project name? Yes
Choose a project: Http-Backend
Confirm project type? Yes
Choose a database framework: PostgreSQL(PGX driver)
Confirm database? Yes
Include test cases? Yes

Project scaffolded successfully!
Project location: /path/to/my-awesome-api
```

Then start developing:
```bash
cd my-awesome-api
go run main.go
```

## Project Structure

```
project-generator/
├── main.go                        # CLI entry point
├── go.mod                         # Go module definition
├── go.sum                         # Dependency checksums
├── makefile                       # Build automation
├── .goreleaser.yaml               # Cross-platform release builds
├── .github/
│   └── workflows/release.yml      # Publishes a release on every v* tag
├── project_generator/             # Core generation logic
│   ├── project_gen.go             # Scaffolding and template rendering
│   ├── forms/                     # Interactive prompt definitions
│   ├── installer/                 # Self-install and PATH setup
│   └── templates/                 # Embedded project templates
└── .gitignore                     # Git configuration
```

## Technologies Used

- Charm Huh - Beautiful terminal forms
- Charm Lipgloss - Terminal styling
- Bubble Tea - TUI framework (included in templates)
- Cobra - CLI framework (included in templates)

## Cross-Platform Support

Prebuilt binaries ship for all six combinations:

- Windows — x86_64, ARM64
- macOS — Intel, Apple Silicon
- Linux — x86_64, ARM64

The tool automatically detects your operating system and adjusts behavior
accordingly, including where it installs itself and how it updates your PATH.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Development

```bash
# Build the project
go build -o project-gen main.go

# Test it
./project-gen
```

Makefile targets:

| Target | What it does |
| --- | --- |
| `make build` | Builds the binary |
| `make install` | Builds, then installs it onto your PATH |
| `make run` | Runs from source with `go run` |
| `make vet` | Runs `go vet` |
| `make release-check` | Validates `.goreleaser.yaml` |
| `make snapshot` | Builds all release archives into `dist/`, publishes nothing |
| `make clean` | Removes the built binary and `dist/` |

### Releasing

Releases are automated. Pushing a tag that starts with `v` triggers
[`.github/workflows/release.yml`](.github/workflows/release.yml), which runs
GoReleaser against [`.goreleaser.yaml`](.goreleaser.yaml):

```bash
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0
```

That single push produces, for all six platform/architecture pairs:

- cross-compiled binaries, built with `CGO_ENABLED=0` for static linking
- `.tar.gz` archives for Linux and macOS, `.zip` for Windows, each containing
  the binary and `README.md`
- a `checksums.txt` with a SHA-256 for every archive
- a GitHub release, with a changelog generated from the commits since the
  previous tag

No secrets to configure — the workflow uses the `GITHUB_TOKEN` that Actions
provides automatically. Tags such as `v1.0.0-rc1` are published as pre-releases,
because `prerelease: auto` reads the suffix.

**Version stamping.** GoReleaser injects the version, commit and build date
through `-ldflags`, so a released binary can identify itself:

```bash
$ project-gen --version
project-gen 0.1.0 (commit 8dad389, built 2026-08-15T21:37:17Z)
```

A build straight from source has nothing injected and reports
`project-gen dev (built from source)`.

**Before tagging**, dry-run the whole release locally. This builds every archive
into `dist/` and publishes nothing:

```bash
make snapshot
```

Requires [GoReleaser](https://goreleaser.com/install/) locally; the CI workflow
installs its own copy, so it is only needed for local dry runs.

## License

This project is open source and available under the MIT License.

## Support

Have questions or found a bug? Please open an issue on GitHub.

---

Happy scaffolding!
