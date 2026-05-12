# tuiference

A minimal terminal WordReference lookup TUI written in Go.

Results are rendered as compact WordReference-style sections, preserving translation groups where the page exposes them.

![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)

## Install

### Go install

```sh
go install github.com/Dandarprox/tuiference@latest
```

Make sure your Go bin directory is on `PATH`. It is usually `~/go/bin`.

### Install script

```sh
curl -fsSL https://raw.githubusercontent.com/Dandarprox/tuiference/main/scripts/install.sh | sh
```

By default this installs to `~/.local/bin`. To choose another directory:

```sh
curl -fsSL https://raw.githubusercontent.com/Dandarprox/tuiference/main/scripts/install.sh | INSTALL_DIR=/usr/local/bin sh
```

### Manual download

Download a release archive from:

```text
https://github.com/Dandarprox/tuiference/releases/latest
```

Extract the archive and move `tuiference` somewhere on your `PATH`.

### From source

```sh
git clone https://github.com/Dandarprox/tuiference.git
cd tuiference
go build -o tuiference .
```

## Usage

```sh
tuiference
```

## Controls

- `Tab`: cycle origin language
- `Shift+Tab`: cycle target language
- `Ctrl+W`: delete the current word in the lookup input
- `Enter`: look up the current word or phrase
- `Up` / `Down` / `PageUp` / `PageDown`: scroll results
- `Esc` / `Ctrl+C`: quit

## Languages

The first version supports WordReference combinations between English, Spanish, and French:

- English -> Spanish
- English -> French
- Spanish -> English
- Spanish -> French
- French -> English
- French -> Spanish

## Release

Releases are published by pushing a semver tag:

```sh
git tag v0.1.0
git push origin v0.1.0
```

GitHub Actions builds Linux, macOS, and Windows archives and attaches them to the GitHub release.
