package model

type JpegBinBlock []byte

type Jpeg struct {
	APPn []JpegBinBlock // Application Segments
	DQT  []JpegBinBlock // Define Quantization Table
	DHT  []JpegBinBlock // Define Huffman Table
	SOF  JpegBinBlock   // Start of Frame
	SOS  JpegBinBlock   // Start of Scan
}

func (j *Jpeg) TotalSize() int {
	var totalSize int

	// Sum the sizes of each segment
	for _, app := range j.APPn {
		totalSize += len(app)
	}
	for _, dqt := range j.DQT {
		totalSize += len(dqt)
	}
	for _, dht := range j.DHT {
		totalSize += len(dht)
	}
	totalSize += len(j.SOF)
	totalSize += len(j.SOS)

	return totalSize
}

//var SOI = [2]byte{0xFF, 0xD8}
//var APPn = [2]byte{0xFF, 0xE0}
//var DQT = [2]byte{0xFF, 0xC4}
//var SOF = [2]byte{0xFF, 0xC0}
//var SOS = [2]byte{0xFF, 0xDA}
