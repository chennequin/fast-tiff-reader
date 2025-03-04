# TIFF reader written in Go

## OpenSlideLake § OpenLakeSlide

By default, TIFF files are loaded from the "assets" directory.

```bash
go mod tidy
go run cmd/gin/main.go
```

Place some TIFF files inside your "assets" at the top of this project.
Then you can query the tiles.

Example:
http://localhost:8080/open/file/generic/CMU-1.tiff
http://localhost:8080/files/mZWa05SMtUVTD9yYpJXZuV2Z/levels/7/tiles/0_0.jpeg

```bash
[GIN] 2025/02/24 - 10:35:11 | 200 |     28.3952ms |             ::1 | GET      "/open/file/generic/CMU-1.tiff"
[GIN] 2025/02/24 - 10:34:54 | 200 |       542.9µs |             ::1 | GET      "/files/mZWa05SMtUVTD9yYpJXZuV2Z/levels/7/tiles/0_0.jpeg"
```

# NOTES:

## Assets
https://camelyon17.grand-challenge.org/Data/

https://openslide.org/
https://openslide.cs.cmu.edu/download/openslide-testdata/

## Cache Algo.
https://s3fifo.com/
https://cachemon.github.io/SIEVE-website/

## S3
https://github.com/gaul/s3proxy
https://hub.docker.com/r/andrewgaul/s3proxy/

## QuPath
https://qupath.readthedocs.io/en/stable/docs/tutorials/index.html

## Cytomine
https://cytomine.com/

## Wasabi
https://wasabi.com/fr/cloud-object-storage