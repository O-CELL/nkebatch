//Copyright  © O-CELL 2018 contact@o-cell.fr
//This source is released under the Apache License 2.0
//which can be found in LICENSE.txt
package nkebatch

import (
	"fmt"
	"log"
	"math"
)

// NkeSample ...
type NkeSample struct {
	Timestamp uint32
	Sample    uint32
	Samplef   float32
}

// SerieParam ...
type SerieParam struct {
	Tag        uint32 `json:"tag"`
	Resolution uint32 `json:"resolution"`
	Type       uint   `json:"type"`
}

// NkeSerie ...
type NkeSerie struct {
	Samples     []NkeSample
	Params      SerieParam
	codingType  uint32
	codingTable uint32
}

// NkeSeries ...
type NkeSeries struct {
	Series            []NkeSerie
	Timestamp         uint32
	counter           uint32
	withcts           byte
	commoncts         uint32
	nosample          byte
	batchReq          byte
	nboftypeofmeasure byte
	labelsize         uint
}

// Config ...
type Config struct {
	Buf       []byte       `json:"buffer"`
	Labelsize uint         `json:"labelsize"`
	Series    []SerieParam `json:"series"`
}

var blog bool = false //Debug flag

// Initialize ...
func Initialize(series *NkeSeries, labelsize uint, params []SerieParam, debug bool) {

	(*series).labelsize = labelsize
	blog = debug
	for _, param := range params {
		// Create series db
		ser := NkeSerie{Params: param}
		(*series).Series = append((*series).Series, ser)
		if blog {
			log.Printf("Label %d \n", param.Tag)
		}
	}

}

// Dump result
func Dump(theseries NkeSeries) {
	fmt.Println(theseries)
}

// ProcessPayload ...
func ProcessPayload(buffer []byte, theseries *NkeSeries) error {
	var index uint
	var absTimestamp, lastTimestamp uint32

	decodeHeader(buffer, theseries, &index)

	//counter
	theseries.counter = buf2Sample(buffer, &index, 3)
	if blog {
		log.Printf("series counter %d \n", (*theseries).counter)
	}

	// jump one reserved bit
	buf2Sample(buffer, &index, 1)

	if err := measureTypeLoop(buffer, theseries, &index, &absTimestamp, &lastTimestamp, getFirstMeasure) ; err != nil {
		return err
	}

	if theseries.nosample == 0 {
		if theseries.commoncts != 0 {
			// common time stamp
			return &Error{ERRNOTSUPPORTEDCTS, mapErrorMessage[ERRNOTSUPPORTEDCTS]}
		} else {
			// separate time stamp
			if blog {
				if blog {
					log.Printf("Separate 	time stamp \n")
				}
			}
			if err := measureTypeLoop(buffer, theseries, &index, &absTimestamp, &lastTimestamp, getSeparatedMeasures) ; err != nil {
				return err
			}
		}
	}

	getLastTimeStamp(buffer, theseries, &index, &absTimestamp, &lastTimestamp)

	return nil
}

// convert types take an int and return a string value.
type traverser func(src []byte, theseries *NkeSeries, index *uint, nbType int, currentser int, absTS *uint32, lastTS *uint32) int

func measureTypeLoop(src []byte, theseries *NkeSeries, index *uint, absTS *uint32, lastTS *uint32, getMeasure traverser) error {
	// First loop on measure type
	for nbtype := 0; nbtype < int(theseries.nboftypeofmeasure); nbtype++ {
		var tag = buf2Sample(src, index, (*theseries).labelsize)
		// get current serie
		currentser := getSeriesFromTag(tag, *theseries)
		if currentser == -1 {
			if blog {
				log.Printf("INVALID CONFIG : Could not retrieve series for index %d, config claims %d series", nbtype, (*theseries).nboftypeofmeasure)
			}
			return &Error{ERRINVALIDCONFIG, mapErrorMessage[ERRINVALIDCONFIG]}
		}
		if blog {
			log.Printf("current tag %d \n", (*theseries).Series[currentser].Params.Tag)
		}
		getMeasure(src, theseries, index, nbtype, currentser, absTS, lastTS)
	}
	return nil
}

//getFirstMeasure traverser specialised in retrieving the first measure of a series with index currentser
//starting at index in buffer src with asbolute timestamp absTS and last time stamp lastTS
func getFirstMeasure(src []byte, theseries *NkeSeries, index *uint, nbtype int, currentser int, absTS *uint32, lastTS *uint32) int {

	// Get first timestamp (uncompressed)
	if nbtype == 0 {
		ts := buf2Sample(src, index, 32)

		(*theseries).Series[currentser].Samples = append((*theseries).Series[currentser].Samples, NkeSample{Timestamp: ts})
		if blog {
			log.Printf("uncompressed timestamp %d \n", ts)
		}
		*absTS = ts
	} else {
		// Get secondary time stamp
		if blog {
			log.Printf("differential timestamp")
		}

		// Delta value
		bi := buf2HuffmanSizeAndIndex(src, index, 1)

		if blog {
			log.Printf("bi: %d\n", bi)
		}

		var ts uint32
		// from huffman index
		if bi <= brHUFFMAXINDEXTABLE {
			if bi > 0 {
				newTS := buf2Sample(src, index, uint(bi))
				if blog {
					log.Printf("raw: %d\n", newTS)
				}
				f := math.Pow(2, float64(bi))
				ts = newTS + uint32(f) - 1 + *absTS
			} else {
				if blog {
					log.Printf("no proper huffman index non processed \n")
				}
				ts = *absTS
			}
		} else {
			ts = buf2Sample(src, index, 32)
		}
		(*theseries).Series[currentser].Samples = append((*theseries).Series[currentser].Samples, NkeSample{Timestamp: ts})
	}

	*lastTS = *absTS
	if blog {
		log.Printf("lastTS %d \n", *lastTS)
	}

	// Get measure (uncompressed)
	(*theseries).Series[currentser].Samples[0].Sample = buf2Sample(src, index,
		mapTypeSize[theseries.Series[currentser].Params.Type])

	if blog {
		log.Printf("type  %d \n", mapTypeSize[(*theseries).Series[currentser].Params.Type])
	}

	if (*theseries).Series[currentser].Params.Type == StFL {
		f := math.Float32frombits((*theseries).Series[currentser].Samples[0].Sample)
		(*theseries).Series[currentser].Samples[0].Samplef = f
		if blog {
			log.Printf("First measure %f \n", f)
		}
	} else {
		expandSign(&(*theseries).Series[currentser].Samples[0].Sample, (*theseries).Series[currentser].Params.Type)
		if blog {
			log.Printf("First measure %d \n", (*theseries).Series[currentser].Samples[0].Sample)
		}
	}

	if (*theseries).nosample == 0 {
		parseCoding(src, theseries, index, currentser)
	}

	return 0
}

//getSeparatedMeasures traverser specialised in retrieving the subsequent measures of the series with index currentser
//starting at index in buffer src with asbolute timestamp absTS and last time stamp lastTS
func getSeparatedMeasures(src []byte, theseries *NkeSeries, index *uint, nbType int, currentser int, absTS *uint32, lastTS *uint32) int {

	//number of samples

	nbsamples := buf2Sample(src, index, mapTypeSize[StU8])
	if blog {
		log.Printf("nb samples %d \n", nbsamples)
	}

	if nbsamples > 0 {
		// get timestamp coding

		tscoding := buf2Sample(src, index, 2)
		if blog {
			log.Printf(" TimeStamp Coding: %d\n", tscoding)
		}

		// samples loop

		var currentTS uint32

		for j := 0; j < int(nbsamples); j++ {

			bi := buf2HuffmanSizeAndIndex(src, index, tscoding)
			if blog {
				log.Printf("  j: %d bi: %d\n", j, bi)
			}

			// from huffman index
			if bi <= brHUFFMAXINDEXTABLE {
				currentIndex := len((*theseries).Series[currentser].Samples) - 1
				if blog {
					log.Printf("serie %d length %d\n", currentser, len((*theseries).Series[currentser].Samples))
				}
				currentTS = (*theseries).Series[currentser].Samples[currentIndex].Timestamp

				if bi > 0 {
					newTS := buf2Sample(src, index, uint(bi))
					f := math.Pow(2, float64(bi))
					currentTS += newTS + uint32(f) - 1
				}

				if blog {
					log.Printf("timestamp %d\n", currentTS)
				}
				// Append the new value
				(*theseries).Series[currentser].Samples = append((*theseries).Series[currentser].Samples, NkeSample{Timestamp: currentTS})

			} else {
				if blog {
					log.Printf("no proper huffman index non processed \n")
				}
			}

			if currentTS > *lastTS {
				*lastTS = currentTS
			}

			// Delta value
			bi = buf2HuffmanSizeAndIndex(src, index, (*theseries).Series[currentser].codingTable)

			if blog {
				log.Printf("bi: %d\n", bi)
			}

			// from huffman index
			if bi <= brHUFFMAXINDEXTABLE {
				if bi > 0 {
					value := buf2Sample(src, index, uint(bi))
					if blog {
						log.Printf("raw: %d\n", value)
					}
					// get last samples
					cur := len((*theseries).Series[currentser].Samples)
					if cur > 0 {
						(*theseries).Series[currentser].Samples[cur-1].Sample = value
						convertValue(theseries, currentser, bi, uint(cur-1))
					}
				} else {
					// get last samples
					cur := len((*theseries).Series[currentser].Samples)
					if cur > 1 {
						(*theseries).Series[currentser].Samples[cur-1].Sample = (*theseries).Series[currentser].Samples[cur-2].Sample
					}
				}
			} else {

				if blog {
					log.Printf("no proper huffman index non processed \n")
				}
			}
		}
	}
	return 0
}

func getLastTimeStamp(src []byte, theseries *NkeSeries, index *uint, absTS *uint32, lastTS *uint32) {
	// Time stamp of the sending
	if *absTS == 0 {
		(*theseries).Timestamp = buf2Sample(src, index, 32)
	} else {
		bi := buf2HuffmanSizeAndIndex(src, index, 1)
		if blog {
			log.Printf(" Final timestamp bi: %d\n", bi)
		}

		// from huffman index
		if bi <= brHUFFMAXINDEXTABLE {
			if bi > 0 {
				newTS := buf2Sample(src, index, uint(bi))
				f := math.Pow(2, float64(bi))
				(*theseries).Timestamp = newTS + *lastTS + uint32(f) - 1
			} else {
				(*theseries).Timestamp = *lastTS
			}
		} else {
			(*theseries).Timestamp = buf2Sample(src, index, 32)
		}
	}
	if blog {
		log.Printf("last timestamp %d \n", (*theseries).Timestamp)
	}
}
