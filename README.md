# bindata

convert any file into manageable Go source code for http service

## Usage
```
bindata gen -i /path/to/your/assets -i /path/to/your/another/assets -p yourPackageName -o /path/to/your/dist_gen.go
```

## Docker
docker image is available `docker pull f1shl3gs/bindata:latest`

## Features
- Build Tags
- Gzip Level, decompress only happened when the asset first access 
    (`--gzip-best-compress` is useful when you have a lot asset
    or want minimal dist size)
- Transform file path, eg: trim prefix
