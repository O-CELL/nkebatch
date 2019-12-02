//Copyright  © O-CELL 2018 contact@o-cell.fr
//This source is released under the Apache License 2.0
//which can be found in LICENSE.txt
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

//Maps each type to it's bits size
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

//getSeriesFromTag returns the index of the series in theseries with tag tag
func getSeriesFromTag(tag uint32, theseries NkeSeries) int {
	for i, ser := range theseries.Series {
		if ser.Params.Tag == tag {
			return i
		}
	}
	return -1
}

//decodeHeader decodes the header from bitstream src starting at index index and stores the result in theseries
func decodeHeader(src []byte, theseries *NkeSeries, index *uint) {

	b := byte(buf2Sample(src, index, 8))
	(*theseries).withcts = (b & 2) >> 1
	(*theseries).nosample = (b & 4) >> 2
	(*theseries).batchReq = (b & 8) >> 3
	(*theseries).nboftypeofmeasure = (b & 240) >> 4
}

//buf2Sample returns an int32 of nbBits from bitstream src starting at bit startBit
func buf2Sample(src []byte, startBit *uint, nbBits uint) uint32 {
	var sample uint32
	var bitToRead uint16
	// per byte step
	nBytes := (nbBits-1)/8 + 1
	nBitsFromByte := nbBits % 8
	startbit := *startBit

	// byte start
	if (nBitsFromByte == 0) && (nBytes > 0) {
		nBitsFromByte = 8
	}
	var idx int
	for nBytes > 0 {
		bitToRead = 0
		for nBitsFromByte > 0 {
			idx = int(startbit >> 3)
			if idx >= len(src) {
				return sample //TODO properly handle error returning
			}
			checkBit := (src[idx] & uint8((1 << (startbit & 0x07))))
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
	// This is performed in expandSign by the caller when needed
	*startBit += nbBits
	return sample
}

//expandSign propagates the sign bit up to position 32 for value if tp is StI4, StI8, StI16 or StI24
func expandSign(value *int32, tp uint) {
	switch tp {
	case StI4, StI8, StI16, StI24:
		nbBits := mapTypeSize[tp]
		if (*value)&(0x1<<(nbBits-1)) != 0 {
			(*value) = int32(int64(*value) - (int64(1) << nbBits))
		}
	}
}

//parseCoding retrieves coding information for series at index current in theseries from bitstream src starting at startBit
func parseCoding(src []byte, theseries *NkeSeries, startBit *uint, current int) {

	(*theseries).Series[current].codingType = buf2Sample(src, startBit, 2)
	(*theseries).Series[current].codingTable = buf2Sample(src, startBit, 2)

	if blog {
		log.Printf("coding type %d coding table %d\n", (*theseries).Series[current].codingType,
			(*theseries).Series[current].codingTable)
	}
}

//convertValue converts the value at sampleindex for series at seriesindex in theseries for huffman table index bi
//Resulting value is stored in the sample .Sample and .Samplef accordingly
func convertValue(theseries *NkeSeries, serieindex int, bi uint8, sampleindex uint) {

	if (*theseries).Series[serieindex].codingType == 0 {
		f := float32(math.Pow(2, float64(bi-1)))
		if (*theseries).Series[serieindex].Samples[sampleindex].Sample >= int32(f) {
			if (*theseries).Series[serieindex].Params.Type != StFL {
				(*theseries).Series[serieindex].Samples[sampleindex].Sample *= (*theseries).Series[serieindex].Params.Resolution
				(*theseries).Series[serieindex].Samples[sampleindex].Sample += (*theseries).Series[serieindex].Samples[sampleindex-1].Sample
			} else {
				(*theseries).Series[serieindex].Samples[sampleindex].Samplef *= float32((*theseries).Series[serieindex].Params.Resolution)
				(*theseries).Series[serieindex].Samples[sampleindex].Samplef += (*theseries).Series[serieindex].Samples[sampleindex-1].Samplef
			}
		} else {
			f := float32(math.Pow(2, float64(bi)))
			if (*theseries).Series[serieindex].Params.Type != StFL {
				(*theseries).Series[serieindex].Samples[sampleindex].Sample += (1 - int32(f))
				(*theseries).Series[serieindex].Samples[sampleindex].Sample *= (*theseries).Series[serieindex].Params.Resolution
				(*theseries).Series[serieindex].Samples[sampleindex].Sample += (*theseries).Series[serieindex].Samples[sampleindex-1].Sample
			} else {
				(*theseries).Series[serieindex].Samples[sampleindex].Samplef += (1.0 - f)
				(*theseries).Series[serieindex].Samples[sampleindex].Samplef *= float32((*theseries).Series[serieindex].Params.Resolution)
				(*theseries).Series[serieindex].Samples[sampleindex].Samplef += (*theseries).Series[serieindex].Samples[sampleindex-1].Samplef
			}
		}
	} else if (*theseries).Series[serieindex].codingType == 1 {
		f := float32(math.Pow(2, float64(bi)))
		if (*theseries).Series[serieindex].Params.Type != StFL {
			(*theseries).Series[serieindex].Samples[sampleindex].Sample += (int32(f) - 1)
			(*theseries).Series[serieindex].Samples[sampleindex].Sample *= (*theseries).Series[serieindex].Params.Resolution
			(*theseries).Series[serieindex].Samples[sampleindex].Sample += (*theseries).Series[serieindex].Samples[sampleindex-1].Sample
		} else {
			(*theseries).Series[serieindex].Samples[sampleindex].Samplef += (f - 1.0)
			(*theseries).Series[serieindex].Samples[sampleindex].Samplef *= float32((*theseries).Series[serieindex].Params.Resolution)
			(*theseries).Series[serieindex].Samples[sampleindex].Samplef += (*theseries).Series[serieindex].Samples[sampleindex-1].Samplef
		}
	} else {
		f := float32(math.Pow(2, float64(bi)))
		if (*theseries).Series[serieindex].Params.Type != StFL {
			(*theseries).Series[serieindex].Samples[sampleindex].Sample += int32(f) - 1
			(*theseries).Series[serieindex].Samples[sampleindex].Sample *= (*theseries).Series[serieindex].Params.Resolution
			(*theseries).Series[serieindex].Samples[sampleindex].Sample = (*theseries).Series[serieindex].Samples[sampleindex-1].Sample - (*theseries).Series[serieindex].Samples[sampleindex].Sample
		} else {
			(*theseries).Series[serieindex].Samples[sampleindex].Samplef += (f - 1.0)
			(*theseries).Series[serieindex].Samples[sampleindex].Samplef *= float32((*theseries).Series[serieindex].Params.Resolution)
			(*theseries).Series[serieindex].Samples[sampleindex].Samplef = (*theseries).Series[serieindex].Samples[sampleindex-1].Samplef - (*theseries).Series[serieindex].Samples[sampleindex].Samplef
		}
	}
}
