# Build instructions

We currently are targeting platforms

## Prerequisites

* Target system: Windows x86_64
* Tested build environment:
	+ msys2 ucrt64
	+ mingw64
* `pacman -Syu`
* Required packages (`pacman -S`):
	+ `mingw-w64-x86_64-go`
	+ `mingw-w64-x86_64-portaudio`
    + `mingw-w64-x86_64-gcc` 

## CGo dependencies and compilation suite configuration

1. Install 
   ```bash
   pacman -S mingw-w64-x86_64-gcc
   pacman -S mingw-w64-x86_64-portaudio
   pacman -S mingw-w64-x86_64-pkg-config
   ```

2. Set `PATH` such that `mingw64/bin` takes precedence
   ```bash
   export PATH=/mingw64/bin:$PATH
   ```

4. Add `mingw64/lib/pkgconfig` to `PKG_CONFIG_PATH`
   ```bash
   export PKG_CONFIG_PATH=$PKG_CONFIG_PATH:/mingw64/lib/pkgconfig
   ```

## Go installation and configuration

1. Install Go using pacman: `pacman -S mingw-w64-x86_64-go`
2. Add Go bin directory to system PATH: `export PATH=/mingw64/lib/go/bin:$PATH`
3. Enabled CGO compilation: `go env -w CGO_ENABLED=1`


## PortAudio setup

1. Set up `PKG_CONFIG` path for PortAudio (`.pc` file)
2. Add pkg-config to system PATH ( suggested location: `/mingw64/bin` )

## Missing Steps

The following steps are missing and need to be completed:

* How to verify the installation of PortAudio and Go packages?
* Are there any specific configuration options or flags required for a successful build?

> They might be other missing informations to uncover.

**Contributions would help us improve the quality of this documentation.**
