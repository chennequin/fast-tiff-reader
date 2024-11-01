package model

// TIFF Tag identifiers with descriptions in English
const (
	NewSubfileType            uint16 = 254   // General indication of the kind of data in the subfile
	SubfileType               uint16 = 255   // Deprecated; replaced by NewSubfileType
	ImageWidth                uint16 = 256   // Width of the image in pixels
	ImageLength               uint16 = 257   // Height of the image in pixels
	BitsPerSample             uint16 = 258   // Number of bits per color channel
	Compression               uint16 = 259   // Compression scheme used on the image data
	PhotometricInterpretation uint16 = 262   // Pixel composition and color interpretation
	Threshholding             uint16 = 263   // Type of thresholding or halftoning applied
	CellWidth                 uint16 = 264   // Width of the dithering or halftoning matrix
	CellLength                uint16 = 265   // Height of the dithering or halftoning matrix
	FillOrder                 uint16 = 266   // Order in which bits are arranged within a byte
	DocumentName              uint16 = 269   // Name of the document from which the image was scanned
	ImageDescription          uint16 = 270   // Description or comments about the image
	Make                      uint16 = 271   // Scanner or camera manufacturer
	Model                     uint16 = 272   // Scanner or camera model
	StripOffsets              uint16 = 273   // Offset to the beginning of each strip in the image
	Orientation               uint16 = 274   // Orientation of the image relative to the rows and columns
	SamplesPerPixel           uint16 = 277   // Number of color channels per pixel
	RowsPerStrip              uint16 = 278   // Number of rows in each strip
	StripByteCounts           uint16 = 279   // Number of bytes per strip after compression
	MinSampleValue            uint16 = 280   // Minimum sample value
	MaxSampleValue            uint16 = 281   // Maximum sample value
	XResolution               uint16 = 282   // Horizontal resolution in pixels per resolution unit
	YResolution               uint16 = 283   // Vertical resolution in pixels per resolution unit
	PlanarConfiguration       uint16 = 284   // Data arrangement: chunky or planar format
	PageName                  uint16 = 285   // Page name
	XPosition                 uint16 = 286   // X offset in resolution units
	YPosition                 uint16 = 287   // Y offset in resolution units
	FreeOffsets               uint16 = 288   // Offsets to unused data
	FreeByteCounts            uint16 = 289   // Sizes of unused data
	GrayResponseUnit          uint16 = 290   // Precision of the gray response curve
	GrayResponseCurve         uint16 = 291   // Gray level mapping of the image
	ResolutionUnit            uint16 = 296   // Unit of measure for X and Y resolution
	PageNumber                uint16 = 297   // Page number of multi-page image
	ColorResponseCurves       uint16 = 300   // Color response curves
	Software                  uint16 = 305   // Name and version of software used to create the image
	DateTime                  uint16 = 306   // Date and time of image creation
	Artist                    uint16 = 315   // Name of the creator of the image
	HostComputer              uint16 = 316   // Computer and/or operating system used to create the image
	Predictor                 uint16 = 317   // Compression predictor for LZW (0 or 1)
	WhitePoint                uint16 = 318   // Chromaticity of the white point of the image
	PrimaryChromaticities     uint16 = 319   // Chromaticities of the primary colors
	ColorMap                  uint16 = 320   // Color map for palette-based images
	HalftoneHints             uint16 = 321   // Gray levels for highlight and shadow
	TileWidth                 uint16 = 322   // Width of a tile in pixels
	TileLength                uint16 = 323   // Height of a tile in pixels
	TileOffsets               uint16 = 324   // Offset to the beginning of each tile
	TileByteCounts            uint16 = 325   // Number of bytes in each tile
	InkSet                    uint16 = 332   // Set of inks used
	InkNames                  uint16 = 333   // Names of inks used
	NumberOfInks              uint16 = 334   // Number of inks
	DotRange                  uint16 = 336   // Range of dot values in halftone images
	TargetPrinter             uint16 = 337   // Target printer for image data
	ExtraSamples              uint16 = 338   // Extra components for each pixel
	SampleFormat              uint16 = 339   // Format of image samples
	SMinSampleValue           uint16 = 340   // Minimum sample value for signed formats
	SMaxSampleValue           uint16 = 341   // Maximum sample value for signed formats
	TransferRange             uint16 = 342   // Transfer range for color components
	ClipPath                  uint16 = 343   // Path used to clip the image
	XClipPathUnits            uint16 = 344   // Units for the X clipping path
	YClipPathUnits            uint16 = 345   // Units for the Y clipping path
	Indexed                   uint16 = 346   // Whether the image data is indexed
	JPEGTables                uint16 = 347   // JPEG quantization and Huffman tables
	OPIProxy                  uint16 = 351   // Indicates if the image is a proxy
	JPEGProc                  uint16 = 512   // JPEG processing mode
	JPEGRestartInterval       uint16 = 515   // Restart interval in JPEG compressed data
	JPEGQTables               uint16 = 517   // Offsets to the quantization tables
	JPEGDCTables              uint16 = 518   // Offsets to the Huffman DC tables
	JPEGACTables              uint16 = 519   // Offsets to the Huffman AC tables
	YCbCrCoefficients         uint16 = 529   // Color conversion coefficients
	YCbCrSubSampling          uint16 = 530   // YCbCr subsampling factors
	YCbCrPositioning          uint16 = 531   // Positioning of chrominance components
	ReferenceBlackWhite       uint16 = 532   // Reference black and white values
	XMP                       uint16 = 700   // XMP metadata in XML format
	Rating                    uint16 = 18246 // Rating from 0 (unrated) to 5 (highly rated)
	RatingPercent             uint16 = 18249 // Rating as a percentage
	ImageID                   uint16 = 32781 // Identifier for the image
	WangAnnotation            uint16 = 32932 // Annotation data in Wang format
	Copyright                 uint16 = 33432 // Copyright notice for the image
	ExifIFD                   uint16 = 34665 // Offset to Exif IFD (metadata for digital images)
	ICCProfile                uint16 = 34675 // ICC profile for color management
	GPSInfoIFD                uint16 = 34853 // Offset to GPS IFD for location information
	InterColorProfile         uint16 = 34857 // Embedded ICC profile
	ExifVersion               uint16 = 36864 // Version of the Exif standard supported
	ShutterSpeedValue         uint16 = 37377 // Shutter speed in APEX units
	ApertureValue             uint16 = 37378 // Lens aperture in APEX units
	BrightnessValue           uint16 = 37379 // Brightness value in APEX units
	ExposureBiasValue         uint16 = 37380 // Exposure bias value in APEX units
	MaxApertureValue          uint16 = 37381 // Maximum lens aperture in APEX units
	SubjectDistance           uint16 = 37382 // Distance to the subject in meters
	MeteringMode              uint16 = 37383 // Metering mode used for the image
	LightSource               uint16 = 37384 // Light source used for the image
	Flash                     uint16 = 37385 // Flash status and settings
	FocalLength               uint16 = 37386 // Focal length of the lens in mm
	MakerNote                 uint16 = 37500 // Manufacturer-specific data
	UserComment               uint16 = 37510 // User comments
	FlashpixVersion           uint16 = 40960 // Supported FlashPix format version
	ColorSpace                uint16 = 40961 // Color space information
	PixelXDimension           uint16 = 40962 // Valid image width
	PixelYDimension           uint16 = 40963 // Valid image height
	InteroperabilityIFD       uint16 = 40965 // Offset to Interoperability IFD
	FocalPlaneXResolution     uint16 = 41486 // Focal plane resolution in X direction
	FocalPlaneYResolution     uint16 = 41487 // Focal plane resolution in Y direction
	FocalPlaneResolutionUnit  uint16 = 41488 // Units for focal plane resolution
	SubjectLocation           uint16 = 41492 // Subject location
	ExposureIndex             uint16 = 41493 // Exposure index setting
	SensingMethod             uint16 = 41495 // Type of image sensor used
	FileSource                uint16 = 41728 // Source of the file
	SceneType                 uint16 = 41729 // Scene type
	CFAPattern                uint16 = 41730 // Color filter array pattern
	CustomRendered            uint16 = 41985 // Whether image is custom-rendered
	ExposureMode              uint16 = 41986 // Exposure mode setting
	WhiteBalance              uint16 = 41987 // White balance setting
	DigitalZoomRatio          uint16 = 41988 // Digital zoom ratio
	FocalLengthIn35mmFilm     uint16 = 41989 // Equivalent focal length in 35mm film
	SceneCaptureType          uint16 = 41990 // Scene capture type
	GainControl               uint16 = 41991 // Amount of gain applied
	Contrast                  uint16 = 41992 // Contrast setting
	Saturation                uint16 = 41993 // Saturation setting
	Sharpness                 uint16 = 41994 // Sharpness setting
	DeviceSettingDescription  uint16 = 41995 // Device setting description
	SubjectDistanceRange      uint16 = 41996 // Distance to the subject
	ImageUniqueID             uint16 = 42016 // Unique identifier for the image
)

// TagsIDsLabels contains a map of TIFF tag IDs to their corresponding names
var TagsIDsLabels = map[uint16]string{
	254:   "NewSubfileType",
	255:   "SubfileType",
	256:   "ImageWidth",
	257:   "ImageLength",
	258:   "BitsPerSample",
	259:   "Compression",
	262:   "PhotometricInterpretation",
	263:   "Threshholding",
	264:   "CellWidth",
	265:   "CellLength",
	266:   "FillOrder",
	269:   "DocumentName",
	270:   "ImageDescription",
	271:   "Make",
	272:   "Model",
	273:   "StripOffsets",
	274:   "Orientation",
	277:   "SamplesPerPixel",
	278:   "RowsPerStrip",
	279:   "StripByteCounts",
	280:   "MinSampleValue",
	281:   "MaxSampleValue",
	282:   "XResolution",
	283:   "YResolution",
	284:   "PlanarConfiguration",
	285:   "PageName",
	286:   "XPosition",
	287:   "YPosition",
	288:   "FreeOffsets",
	289:   "FreeByteCounts",
	290:   "GrayResponseUnit",
	291:   "GrayResponseCurve",
	296:   "esolutionUnit",
	297:   "PageNumber",
	300:   "ColorResponseCurves",
	305:   "Software",
	306:   "DateTime",
	315:   "Artist",
	316:   "HostComputer",
	317:   "Predictor",
	318:   "WhitePoint",
	319:   "PrimaryChromaticities",
	320:   "ColorMap",
	321:   "HalftoneHints",
	322:   "TileWidth",
	323:   "TileLength",
	324:   "TileOffsets",
	325:   "TileByteCounts",
	332:   "InkSet",
	333:   "InkNames",
	334:   "NumberOfInks",
	336:   "DotRange",
	337:   "TargetPrinter",
	338:   "ExtraSamples",
	339:   "SampleFormat",
	340:   "SMinSampleValue",
	341:   "SMaxSampleValue",
	342:   "TransferRange",
	343:   "ClipPath",
	344:   "XClipPathUnits",
	345:   "YClipPathUnits",
	346:   "Indexed",
	347:   "JPEGTables",
	351:   "OPIProxy",
	512:   "JPEGProc",
	515:   "JPEGRestartInterval",
	517:   "JPEGQTables",
	518:   "JPEGDCTables",
	519:   "JPEGACTables",
	529:   "YCbCrCoefficients",
	530:   "YCbCrSubSampling",
	531:   "YCbCrPositioning",
	532:   "ReferenceBlackWhite",
	700:   "XMP",
	18246: "Rating",
	18249: "RatingPercent",
	32781: "ImageID",
	32932: "WangAnnotation",
	33432: "Copyright",
	34665: "ExifIFD",
	34675: "ICCProfile",
	34853: "GPSInfoIFD",
	34857: "InterColorProfile",
	36864: "ExifVersion",
	37377: "ShutterSpeedValue",
	37378: "ApertureValue",
	37379: "BrightnessValue",
	37380: "ExposureBiasValue",
	37381: "MaxApertureValue",
	37382: "SubjectDistance",
	37383: "MeteringMode",
	37384: "LightSource",
	37385: "Flash",
	37386: "FocalLength",
	37500: "MakerNote",
	37510: "UserComment",
	40960: "FlashpixVersion",
	40961: "ColorSpace",
	40962: "PixelXDimension",
	40963: "PixelYDimension",
	40965: "InteroperabilityIFD",
	41486: "FocalPlaneXResolution",
	41487: "FocalPlaneYResolution",
	41488: "FocalPlaneResolutionUnit",
	41492: "SubjectLocation",
	41493: "ExposureIndex",
	41495: "SensingMethod",
	41728: "FileSource",
	41729: "SceneType",
	41730: "CFAPattern",
	41985: "CustomRendered",
	41986: "ExposureMode",
	41987: "WhiteBalance",
	41988: "DigitalZoomRatio",
	41989: "FocalLengthIn35mmFilm",
	41990: "SceneCaptureType",
	41991: "GainControl",
	41992: "Contrast",
	41993: "Saturation",
	41994: "Sharpness",
	41995: "DeviceSettingDescription",
	41996: "SubjectDistanceRange",
	42016: "ImageUniqueID",
}
