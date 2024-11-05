# TIFF reader written in Go

## OpenSlideLake § OpenLakeSlide

https://github.com/dsoprea/go-exif?tab=readme-ov-file
https://openslide.org/
https://openslide.cs.cmu.edu/download/openslide-testdata/

https://docs.scanimage.org/Appendix/ScanImage%2BBigTiff%2BSpecification.html#ScanImageBigTiffSpecification-MagicNumb

HEADER

+--------------------+--------------------+-----------------------------+ 
| Byte Order (2) | TIFF Identifier (2) | Offset du Premier IFD (4) | 
+--------------------+--------------------+-----------------------------+ 
| 0x49 0x49 | 0x2A | 0x00000020 (exemple d'offset) | 
+--------------------+--------------------+-----------------------------+

In the file header, BigTIFF is declared as 0x002B (43) at offset 2 bytes as compared with the TIFF version of 0x002A (42)
Bytesize of offsets is always 8 in BigTIFF (not present in TIFF)
IFD (Image File Directory) takes up 20 bytes in BigTIFF as opposed to 12 in TIFF
11_081553.svs (bigTIFF)

0x49 49 / 2b 00 / 08 00 00 00 / 7e d4 64 6c 00 00 00 00 / ff d8 ff e0 …

CMU-1.svs (TIFF)

0x49 49 / 2a 00 / 10 7a 5d 09 / 00 00 00 00 00 00 00 00 / ff d8 ff c0 …

FFD8 est le marqueur de début d'une image JPEG.

IFD

+----------------+---------------------------------+ | Nombre de Tags | 2 octets | +----------------+---------------------------------+ | Tag ID | 2 octets | | Type | 2 octets | | Count | 4 octets | | Value/Offset | 4 octets | +----------------+---------------------------------+ | Tag ID | 2 octets | | Type | 2 octets | | Count | 4 octets | | Value/Offset | 4 octets | +----------------+---------------------------------+ | ... (n tags) | | +----------------+---------------------------------+ | Offset vers le | 4 octets (offset ou 0) | | prochain IFD | | +----------------+---------------------------------+

TAG

+---------+---------+---------+----------+ | Tag ID | Type | Count | Value/ | | | | | Offset | +---------+---------+---------+----------+ | 256 | 3 | 1 | 0x10 | (ex. ImageWidth) +---------+---------+---------+----------+

Types de Tags :

0x0001 - BYTE
0x0002 - ASCII
0x0003 - SHORT
0x0004 - LONG
0x0005 - RATIONAL
0x0006 - SBYTE
0x0007 - UNDEFINED
0x0008 - SSHORT
0x0009 - SLONG
0x000A - SRATIONAL
0x000B - FLOAT
0x000C - DOUBLE
0x000D - LONG8 (pour 64 bits)