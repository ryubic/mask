# mask

A small command-line tool that masks sensitive information from terminal output.

## Installation

1. Download the latest release for your operating system and architecture.
2. Extract the ZIP archive.
3. Add the extracted directory to your `PATH`.

Both `mask` and `cmask` should be in the same directory.

## Usage

Pipe command output into `mask`:

```bash
echo "User: $(whoami), IP: 192.168.1.10" | mask
```

Without `mask`:

```text
User: alice, IP: 192.168.1.10
```

With `mask`:

```text
User: [USER], IP: [IP_HIDDEN]
```

To mask the output and copy it to the clipboard, use `cmask`:

```bash
echo "User: $(whoami), IP: 192.168.1.10" | cmask
```

Alternatively, you can use:

```bash
echo "User: $(whoami), IP: 192.168.1.10" | mask -c
```

`cmask` is a shortcut for `mask -c`.

## Configuration

Rules are configured in `config.json`.

You can enable or disable rules and customize their patterns and replacements.

## Building from Source

Make sure you have `go`, `zip`, and `sha256sum` installed.

Run:

```bash
chmod +x build.sh
./build.sh
```

The compiled releases will be placed in the `dist/` directory.

## Note

This project is mostly **AI-generated** code since I don't really know Go and wanted a simple and lightweight cross-platform solution. It has been tested, but it may still contain bugs or unexpected behavior. 

Please feel free to report bugs, suggest improvements, or contribute to the project.
