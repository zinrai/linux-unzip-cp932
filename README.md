# linux-unzip-cp932

linux-unzip-cp932 is a tool designed to correctly extract ZIP archives containing CP932 (Shift_JIS) encoded filenames on Linux environments, typically created on Windows systems. Password-protected ZIP files are also supported.

Standard `unzip` commands often fail to properly handle CP932 encoded filenames, resulting in garbled characters. This tool resolves such issues, allowing for smooth extraction of Windows-created ZIP files in Linux environments, including password-protected archives.

## Features

- Correctly decodes CP932 encoded filenames
- Supports the GPB11 flag (UTF-8 encoding flag) in ZIP files
- Intelligently detects filename encoding (CP932 vs UTF-8)
- Extracts ZIP archive contents to a specified directory
- Preserves directory structure during extraction
- Supports password-protected ZIP files

## Notes

- This tool supports filenames encoded in CP932 (Microsoft's extension of Shift_JIS)
- It automatically handles the GPB11 flag in ZIP headers (which indicates UTF-8 encoding)
- It does not alter the encoding of the file contents within the ZIP archive
- For encrypted ZIP files, make sure to provide the correct password using the `-password` option

## Installation

Build the tool:

```
$ go build
```

## Usage

The basic usage is as follows:

```
$ ./linux-unzip-cp932 -input <zip_file> [-output <output_directory>] [-password <zip_password>]
```

Examples:

```
$ ./linux-unzip-cp932 -input example.zip -output ./extracted
$ ./linux-unzip-cp932 -input encrypted.zip -output ./extracted -password mypassword
```

## How It Works

The tool performs the following steps during extraction:

1. Checks if the GPB11 flag is set for each file in the ZIP archive
   - If GPB11 is set, the filename is treated as UTF-8 encoded
   - If GPB11 is not set, the tool attempts to determine if the filename is CP932 encoded
2. For files with CP932 encoded filenames, it decodes them to UTF-8
3. It creates the necessary directory structure
4. It extracts the file contents to the corresponding locations

## License

This project is licensed under the [MIT License](./LICENSE).
