package tags

type TagID uint16

// TIFF Tag identifiers with descriptions in English
const (
	NewSubfileType            = TagID(uint16(254))   // General indication of the kind of data in the subfile
	SubfileType               = TagID(uint16(255))   // Deprecated; replaced by NewSubfileType
	ImageWidth                = TagID(uint16(256))   // Width of the image in pixels
	ImageLength               = TagID(uint16(257))   // Height of the image in pixels
	BitsPerSample             = TagID(uint16(258))   // Number of bits per color channel
	Compression               = TagID(uint16(259))   // Compression scheme used on the image data
	PhotometricInterpretation = TagID(uint16(262))   // Pixel composition and color interpretation
	Threshholding             = TagID(uint16(263))   // Type of thresholding or halftoning applied
	CellWidth                 = TagID(uint16(264))   // Width of the dithering or halftoning matrix
	CellLength                = TagID(uint16(265))   // Height of the dithering or halftoning matrix
	FillOrder                 = TagID(uint16(266))   // Order in which bits are arranged within a byte
	DocumentName              = TagID(uint16(269))   // Name of the document from which the image was scanned
	ImageDescription          = TagID(uint16(270))   // Description or comments about the image
	Make                      = TagID(uint16(271))   // Scanner or camera manufacturer
	Model                     = TagID(uint16(272))   // Scanner or camera model
	StripOffsets              = TagID(uint16(273))   // Offset to the beginning of each strip in the image
	Orientation               = TagID(uint16(274))   // Orientation of the image relative to the rows and columns
	SamplesPerPixel           = TagID(uint16(277))   // Number of color channels per pixel
	RowsPerStrip              = TagID(uint16(278))   // Number of rows in each strip
	StripByteCounts           = TagID(uint16(279))   // Number of bytes per strip after compression
	MinSampleValue            = TagID(uint16(280))   // Minimum sample value
	MaxSampleValue            = TagID(uint16(281))   // Maximum sample value
	XResolution               = TagID(uint16(282))   // Horizontal resolution in pixels per resolution unit
	YResolution               = TagID(uint16(283))   // Vertical resolution in pixels per resolution unit
	PlanarConfiguration       = TagID(uint16(284))   // Data arrangement: chunky or planar format
	PageName                  = TagID(uint16(285))   // Page name
	XPosition                 = TagID(uint16(286))   // X offset in resolution units
	YPosition                 = TagID(uint16(287))   // Y offset in resolution units
	FreeOffsets               = TagID(uint16(288))   // Offsets to unused data
	FreeByteCounts            = TagID(uint16(289))   // Sizes of unused data
	GrayResponseUnit          = TagID(uint16(290))   // Precision of the gray response curve
	GrayResponseCurve         = TagID(uint16(291))   // Gray level mapping of the image
	ResolutionUnit            = TagID(uint16(296))   // Unit of measure for X and Y resolution
	PageNumber                = TagID(uint16(297))   // Page number of multi-page image
	ColorResponseCurves       = TagID(uint16(300))   // Color response curves
	Software                  = TagID(uint16(305))   // Name and version of software used to create the image
	DateTime                  = TagID(uint16(306))   // Date and time of image creation
	Artist                    = TagID(uint16(315))   // Name of the creator of the image
	HostComputer              = TagID(uint16(316))   // Computer and/or operating system used to create the image
	Predictor                 = TagID(uint16(317))   // Compression predictor for LZW (0 or 1)
	WhitePoint                = TagID(uint16(318))   // Chromaticity of the white point of the image
	PrimaryChromaticities     = TagID(uint16(319))   // Chromaticities of the primary colors
	ColorMap                  = TagID(uint16(320))   // Color map for palette-based images
	HalftoneHints             = TagID(uint16(321))   // Gray levels for highlight and shadow
	TileWidth                 = TagID(uint16(322))   // Width of a tile in pixels
	TileLength                = TagID(uint16(323))   // Height of a tile in pixels
	TileOffsets               = TagID(uint16(324))   // Offset to the beginning of each tile
	TileByteCounts            = TagID(uint16(325))   // Number of bytes in each tile
	InkSet                    = TagID(uint16(332))   // Set of inks used
	InkNames                  = TagID(uint16(333))   // Names of inks used
	NumberOfInks              = TagID(uint16(334))   // Number of inks
	DotRange                  = TagID(uint16(336))   // Range of dot values in halftone images
	TargetPrinter             = TagID(uint16(337))   // Target printer for image data
	ExtraSamples              = TagID(uint16(338))   // Extra components for each pixel
	SampleFormat              = TagID(uint16(339))   // Format of image samples
	SMinSampleValue           = TagID(uint16(340))   // Minimum sample value for signed formats
	SMaxSampleValue           = TagID(uint16(341))   // Maximum sample value for signed formats
	TransferRange             = TagID(uint16(342))   // Transfer range for color components
	ClipPath                  = TagID(uint16(343))   // Path used to clip the image
	XClipPathUnits            = TagID(uint16(344))   // Units for the X clipping path
	YClipPathUnits            = TagID(uint16(345))   // Units for the Y clipping path
	Indexed                   = TagID(uint16(346))   // Whether the image data is indexed
	JPEGTables                = TagID(uint16(347))   // JPEG quantization and Huffman tables
	OPIProxy                  = TagID(uint16(351))   // Indicates if the image is a proxy
	JPEGProc                  = TagID(uint16(512))   // JPEG processing mode
	JPEGRestartInterval       = TagID(uint16(515))   // Restart interval in JPEG compressed data
	JPEGQTables               = TagID(uint16(517))   // Offsets to the quantization tables
	JPEGDCTables              = TagID(uint16(518))   // Offsets to the Huffman DC tables
	JPEGACTables              = TagID(uint16(519))   // Offsets to the Huffman AC tables
	YCbCrCoefficients         = TagID(uint16(529))   // Color conversion coefficients
	YCbCrSubSampling          = TagID(uint16(530))   // YCbCr subsampling factors
	YCbCrPositioning          = TagID(uint16(531))   // Positioning of chrominance components
	ReferenceBlackWhite       = TagID(uint16(532))   // Reference black and white values
	XMP                       = TagID(uint16(700))   // XMP metadata in XML format
	Rating                    = TagID(uint16(18246)) // Rating from 0 (unrated) to 5 (highly rated)
	RatingPercent             = TagID(uint16(18249)) // Rating as a percentage
	ImageID                   = TagID(uint16(32781)) // Identifier for the image
	WangAnnotation            = TagID(uint16(32932)) // Annotation data in Wang format
	Copyright                 = TagID(uint16(33432)) // Copyright notice for the image
	ExifIFD                   = TagID(uint16(34665)) // Offset to Exif IFD (metadata for digital images)
	ICCProfile                = TagID(uint16(34675)) // ICC profile for color management
	GPSInfoIFD                = TagID(uint16(34853)) // Offset to GPS IFD for location information
	InterColorProfile         = TagID(uint16(34857)) // Embedded ICC profile
	ExifVersion               = TagID(uint16(36864)) // Version of the Exif standard supported
	ShutterSpeedValue         = TagID(uint16(37377)) // Shutter speed in APEX units
	ApertureValue             = TagID(uint16(37378)) // Lens aperture in APEX units
	BrightnessValue           = TagID(uint16(37379)) // Brightness value in APEX units
	ExposureBiasValue         = TagID(uint16(37380)) // Exposure bias value in APEX units
	MaxApertureValue          = TagID(uint16(37381)) // Maximum lens aperture in APEX units
	SubjectDistance           = TagID(uint16(37382)) // Distance to the subject in meters
	MeteringMode              = TagID(uint16(37383)) // Metering mode used for the image
	LightSource               = TagID(uint16(37384)) // Light source used for the image
	Flash                     = TagID(uint16(37385)) // Flash status and settings
	FocalLength               = TagID(uint16(37386)) // Focal length of the lens in mm
	MakerNote                 = TagID(uint16(37500)) // Manufacturer-specific data
	UserComment               = TagID(uint16(37510)) // User comments
	FlashpixVersion           = TagID(uint16(40960)) // Supported FlashPix format version
	ColorSpace                = TagID(uint16(40961)) // Color space information
	PixelXDimension           = TagID(uint16(40962)) // Valid image width
	PixelYDimension           = TagID(uint16(40963)) // Valid image height
	InteroperabilityIFD       = TagID(uint16(40965)) // Offset to Interoperability IFD
	FocalPlaneXResolution     = TagID(uint16(41486)) // Focal plane resolution in X direction
	FocalPlaneYResolution     = TagID(uint16(41487)) // Focal plane resolution in Y direction
	FocalPlaneResolutionUnit  = TagID(uint16(41488)) // Units for focal plane resolution
	SubjectLocation           = TagID(uint16(41492)) // Subject location
	ExposureIndex             = TagID(uint16(41493)) // Exposure index setting
	SensingMethod             = TagID(uint16(41495)) // Type of image sensor used
	FileSource                = TagID(uint16(41728)) // Source of the file
	SceneType                 = TagID(uint16(41729)) // Scene type
	CFAPattern                = TagID(uint16(41730)) // Color filter array pattern
	CustomRendered            = TagID(uint16(41985)) // Whether image is custom-rendered
	ExposureMode              = TagID(uint16(41986)) // Exposure mode setting
	WhiteBalance              = TagID(uint16(41987)) // White balance setting
	DigitalZoomRatio          = TagID(uint16(41988)) // Digital zoom ratio
	FocalLengthIn35mmFilm     = TagID(uint16(41989)) // Equivalent focal length in 35mm film
	SceneCaptureType          = TagID(uint16(41990)) // Scene capture type
	GainControl               = TagID(uint16(41991)) // Amount of gain applied
	Contrast                  = TagID(uint16(41992)) // Contrast setting
	Saturation                = TagID(uint16(41993)) // Saturation setting
	Sharpness                 = TagID(uint16(41994)) // Sharpness setting
	DeviceSettingDescription  = TagID(uint16(41995)) // Device setting description
	SubjectDistanceRange      = TagID(uint16(41996)) // Distance to the subject
	ImageUniqueID             = TagID(uint16(42016)) // Unique identifier for the image
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
	296:   "ResolutionUnit",
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
