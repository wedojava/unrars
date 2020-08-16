# unrars

unrars is a decompress app, use cpus as more as possible, decompress more tarballs as more as possible at one time.

**By default**, it'll treat tarballs at `./` and decompress to `./_decompress`

unrars can decompress archives: **rar, tar, tar.gz, tar.bz2, zip, 7z**

# Usage

The mose simple and directly way is just run `unrars`, it can decompress all archives beside the file unrars.

1. help:
```
unrars -h
```

2. specify path:
```
unrars -s "<your archives folder>"
unrars -d "<your decompress destination folder>"
unrars -s "<your archives folder>" -d "<your decompress destination folder>"
```

# Release

Because the library `unarr` I used are cgo library, so unrars cannot cross compiling.
I release only linux and windows version on amd64.
