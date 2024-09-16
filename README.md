# nearstand

Purpose:

Nearstand is a command-line tool designed to efficiently shrink image files while maintaining acceptable quality. It's particularly useful for batch processing large numbers of images or entire directories.

## Features

- Shrink individual image files or entire directories
- Support for JPG, JPEG, PNG, and GIF formats
- Optional reshrinking of already processed images
- Detailed statistics for each processed image and aggregate results
- Configurable verbosity and log format

## Example Usage

```bash
# Shrink a single image
nearstand shrink path/to/image.jpg

# Shrink all images in a directory
nearstand shrink path/to/image/directory

# Shrink images, including those already processed
nearstand shrink --reshrink path/to/image/directory

# Shrink images with verbose output
nearstand shrink -v path/to/image/directory

# Shrink images with JSON-formatted logs
nearstand shrink --log-format json path/to/image/directory

# Print version information
nearstand version
```

## Install nearstand

On macOS/Linux:
```bash
brew install gkwa/homebrew-tools/nearstand
```

On Windows:

```powershell
TBD
```

## Configuration

Nearstand can be configured using a YAML file. By default, it looks for `.nearstand.yaml` in your home directory. You can specify a different config file using the `--config` flag.

Example configuration:

```yaml
verbose: true
log-format: json
```

## Requirements

- ImageMagick must be installed on your system

## Building from Source

To build Nearstand from source, ensure you have Go installed, then run:

```bash
go build -o nearstand main.go
```

## Usage Options

```
Usage:
  nearstand [command]

Available Commands:
  help        Help about any command
  shrink      Shrink image file(s)
  version     Print the version number of nearstand

Flags:
      --config string     config file (default is $HOME/.nearstand.yaml)
  -h, --help              help for nearstand
      --log-format string json or text (default is text)
  -v, --verbose           enable verbose mode

Use "nearstand [command] --help" for more information about a command.
```
```
