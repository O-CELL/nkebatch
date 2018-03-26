package nkebatch

import (
	"log"
	"math"
)

/* Type defintion */
const (
	StUndef = 0
	StBl    = iota
	StU4
	StI4
	StU8
	StI8
	StU16
	StI16
	StU24
	StI24
	StU32
	StI32
	StFL
)

var mapTypeSize = map[uint]uint{
	StUndef: 0,
	StBl:    1,
	StU4:    4,
	StI4:    4,
	StU8:    8,
	StI8:    8,
	StU16:   16,
	StI16:   16,
	StU24:   24,
	StI24:   24,
	StU32:   32,
	StI32:   32,
	StFL:    32,
}

func getSeriesFromTag(tag uint32, theseries NkeSeries) int {
	for i, ser := range theseries.Series {
		if ser.Params.Tag == tag {
			return i
		}
	}
	return -1
}

func decodeHeader(src []byte, theseries *NkeSeries, index *uint) error {

	b := byte(buf2Sample(StU8, src, index, 8))

	(*theseries).withcts = (b & 2) >> 1
	(*theseries).nosample = (b & 4) >> 2
	(*theseries).batchReq = (b & 8) >> 3
	(*theseries).nboftypeofmeasure = (b & 240) >> 4

	return nil
}

// fetch data
func buf2Sample(Type uint, src []byte, startBit *uint, nbBits uint) uint32 {
	// Check function
	var sample uint32
	// per byte step
	var nBytes = (nbBits-1)/8 + 1
	var nBitsFromByte = nbBits % 8
	var bitToRead uint16
	startbit := *startBit

	// byte start
	if (nBitsFromByte == 0) && (nBytes > 0) {
		nBitsFromByte = 8
	}

	for nBytes > 0 {
		bitToRead = 0
		for nBitsFromByte > 0 {
			checkBit := (uint8(src[startbit>>3]) & uint8((1 << (startbit & 0x07))))
			if checkBit != 0 {
				sample |= (1 << ((uint16(nBytes-1) * 8) + bitToRead))
			}
			nBitsFromByte--
			bitToRead++
			startbit++
		}
		nBytes--
		nBitsFromByte = 8
	}

	// Propagate the sign bit if 1 for Integer type
	// TO DO STI4 STI24
	*startBit += nbBits
	return sample
}

func parseCoding(src []byte, theseries *NkeSeries, startBit *uint, current int, blog bool) {

	(*theseries).Series[current].codingType = buf2Sample(StU8, src, startBit, 2)
	(*theseries).Series[current].codingTable = buf2Sample(StU8, src, startBit, 2)

	if blog {
		log.Printf("coding type %d coding table %d\n", (*theseries).Series[current].codingType,
			(*theseries).Series[current].codingTable)
	}
}

func convertValue(theseries *NkeSeries, serieindex int, bi uint8, sampleindex uint, blog bool) {

	if (*theseries).Series[serieindex].codingType == 0 {
		f := float32(math.Pow(2, float64(bi-1)))
		if (*theseries).Series[serieindex].Samples[sampleindex].Sample >= uint32(f) {
			if (*theseries).Series[serieindex].Params.Type != StFL {
				(*theseries).Series[serieindex].Samples[sampleindex].Sample = (*theseries).Series[serieindex].Samples[sampleindex].Sample * (*theseries).Series[serieindex].Params.Resolution
				(*theseries).Series[serieindex].Samples[sampleindex].Sample = (*theseries).Series[serieindex].Samples[sampleindex].Sample + (*theseries).Series[serieindex].Samples[sampleindex-1].Sample
			} else {
				fs := math.Float32frombits((*theseries).Series[serieindex].Samples[sampleindex].Sample)
				fs = f * float32((*theseries).Series[serieindex].Params.Resolution)
				(*theseries).Series[serieindex].Samples[sampleindex].Samplef = fs + (*theseries).Series[serieindex].Samples[sampleindex-1].Samplef + 1.0 - f
			}
		} else {
			f := float32(math.Pow(2, float64(bi)))
			if (*theseries).Series[serieindex].Params.Type != StFL {
				(*theseries).Series[serieindex].Samples[sampleindex].Sample = (*theseries).Series[serieindex].Samples[sampleindex].Sample + 1 - uint32(f)
				(*theseries).Series[serieindex].Samples[sampleindex].Sample = (*theseries).Series[serieindex].Samples[sampleindex].Sample * (*theseries).Series[serieindex].Params.Resolution
				(*theseries).Series[serieindex].Samples[sampleindex].Sample = (*theseries).Series[serieindex].Samples[sampleindex].Sample + (*theseries).Series[serieindex].Samples[sampleindex-1].Sample
			} else {
				(*theseries).Series[serieindex].Samples[sampleindex].Samplef = float32((*theseries).Series[serieindex].Samples[sampleindex].Sample) + 1.0 - f
				(*theseries).Series[serieindex].Samples[sampleindex].Sample = (*theseries).Series[serieindex].Samples[sampleindex].Sample * (*theseries).Series[serieindex].Params.Resolution
				(*theseries).Series[serieindex].Samples[sampleindex].Samplef = (*theseries).Series[serieindex].Samples[sampleindex].Samplef + (*theseries).Series[serieindex].Samples[sampleindex-1].Samplef
			}
		}
	} else if (*theseries).Series[serieindex].codingType == 1 {
		f := float32(math.Pow(2, float64(bi)))
		if (*theseries).Series[serieindex].Params.Type != StFL {
			(*theseries).Series[serieindex].Samples[sampleindex].Sample = (*theseries).Series[serieindex].Samples[sampleindex].Sample + uint32(f) - 1
			(*theseries).Series[serieindex].Samples[sampleindex].Sample = (*theseries).Series[serieindex].Samples[sampleindex].Sample * (*theseries).Series[serieindex].Params.Resolution
			(*theseries).Series[serieindex].Samples[sampleindex].Sample = (*theseries).Series[serieindex].Samples[sampleindex].Sample + (*theseries).Series[serieindex].Samples[sampleindex-1].Sample
		} else {
			(*theseries).Series[serieindex].Samples[sampleindex].Samplef = float32((*theseries).Series[serieindex].Samples[sampleindex].Sample) + f - 1.0
			(*theseries).Series[serieindex].Samples[sampleindex].Sample = (*theseries).Series[serieindex].Samples[sampleindex].Sample * (*theseries).Series[serieindex].Params.Resolution
			(*theseries).Series[serieindex].Samples[sampleindex].Samplef = (*theseries).Series[serieindex].Samples[sampleindex].Samplef + (*theseries).Series[serieindex].Samples[sampleindex-1].Samplef
		}
	} else {
		f := float32(math.Pow(2, float64(bi)))
		if (*theseries).Series[serieindex].Params.Type != StFL {
			(*theseries).Series[serieindex].Samples[sampleindex].Sample = (*theseries).Series[serieindex].Samples[sampleindex].Sample + uint32(f) - 1
			(*theseries).Series[serieindex].Samples[sampleindex].Sample = (*theseries).Series[serieindex].Samples[sampleindex].Sample * (*theseries).Series[serieindex].Params.Resolution
			(*theseries).Series[serieindex].Samples[sampleindex].Sample = (*theseries).Series[serieindex].Samples[sampleindex-1].Sample - (*theseries).Series[serieindex].Samples[sampleindex].Sample
		} else {
			(*theseries).Series[serieindex].Samples[sampleindex].Samplef = float32((*theseries).Series[serieindex].Samples[sampleindex].Sample) + f - 1.0
			(*theseries).Series[serieindex].Samples[sampleindex].Sample = (*theseries).Series[serieindex].Samples[sampleindex].Sample * (*theseries).Series[serieindex].Params.Resolution
			(*theseries).Series[serieindex].Samples[sampleindex].Samplef = (*theseries).Series[serieindex].Samples[sampleindex-1].Samplef - (*theseries).Series[serieindex].Samples[sampleindex].Samplef
		}
	}
}
