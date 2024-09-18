# linux-unzip-cp932

linux-unzip-cp932 is a tool designed to correctly extract ZIP archives containing CP932 (Shift_JIS) encoded filenames on Linux environments, typically created on Windows systems.

Standard `unzip` commands often fail to properly handle CP932 encoded filenames, resulting in garbled characters. This tool resolves such issues, allowing for smooth extraction of Windows-created ZIP files in Linux environments.

## Notes

- This tool supports filenames encoded in CP932 (Microsoft's extension of Shift_JIS).
- It does not alter the encoding of the file contents within the ZIP archive.

## Features

- Correctly decodes CP932 encoded filenames
- Extracts ZIP archive contents to a specified directory
- Preserves directory structure during extraction

## Installation

Build the tool:

```
$ go build
```

## Usage

The basic usage is as follows:

```
$ linux-unzip-cp932 -input <zip_file> [-output <output_directory>]
```

Example:

```
$ linux-unzip-cp932 -input example.zip -output ./extracted
```

## Testing

To run the tests, execute the following command in the project root directory:

```
$ go test -v
```

## License

This project is licensed under the MIT License - see the [LICENSE](https://opensource.org/license/mit) for details.
