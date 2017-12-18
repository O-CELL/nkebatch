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

// Initialize ...
func Initialize(series *NkeSeries, labelsize uint, params []SerieParam, blog bool) {

	(*series).labelsize = labelsize
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
func ProcessPayload(buffer []byte, theseries *NkeSeries, blog bool) error {
	var index uint
	var currentts uint32
	var absTimestamp, lastTimestamp uint32

	err := decodeHeader(buffer, theseries, &index)

	if err != nil {
		return err
	}

	//counter
	theseries.counter = buf2Sample(StU8, buffer, &index, 3)
	if blog {
		log.Printf("series counter %d \n", (*theseries).counter)
	}

	// jump one reserved bit
	buf2Sample(StU8, buffer, &index, 1)

	measureTypeLoop(buffer, theseries, &index, &currentts, &absTimestamp, &lastTimestamp, getFirstMeasure, blog)

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
			measureTypeLoop(buffer, theseries, &index, &currentts, &absTimestamp, &lastTimestamp, getSeparatedMeasures, blog)

		}
	}

	getLastTimeStamp(buffer, theseries, &index, &absTimestamp, &lastTimestamp, blog)

	return nil
}

// convert types take an int and return a string value.
type traverser func(src []byte, theseries *NkeSeries, index *uint, nbType int, currentser int, absTS *uint32, lastTS *uint32, ts *uint32, blog bool) int

func measureTypeLoop(src []byte, theseries *NkeSeries, index *uint, currentts *uint32, absTS *uint32, lastTS *uint32, getMeasure traverser, blog bool) {
	// First loop on measure type
	for nbtype := 0; nbtype < int(theseries.nboftypeofmeasure); nbtype++ {
		// get tagbyte()
		//Tagfor nbtype := 0; nbtype < int(theseries.nboftypeofmeasure); nbtype++ {
		var tag = buf2Sample(StU32, src, index, (*theseries).labelsize)

		// get current serie
		currentser := getSeriesFromTag(tag, *theseries)

		if blog {
			log.Printf("current tag %d \n", (*theseries).Series[currentser].Params.Tag)
		}

		getMeasure(src, theseries, index, nbtype, currentser, absTS, lastTS, currentts, blog)

	}
}

// getFirstMeasure
func getFirstMeasure(src []byte, theseries *NkeSeries, index *uint, nbtype int, currentser int, absTS *uint32, lastTS *uint32, currentts *uint32, blog bool) int {

	// Get first timestamp (uncompressed)
	if nbtype == 0 {
		ts := buf2Sample(StU32, src, index, 32)

		(*theseries).Series[currentser].Samples = append((*theseries).Series[currentser].Samples, NkeSample{Timestamp: ts})
		if blog {
			log.Printf("uncompressed timestamp %d \n", ts)
		}
		*currentts = ts
		*absTS = ts
	} else {
		// Get secondary time stamp
		if blog {
			log.Printf("differential timestamp")
		}

		// Delta value
		bi := but2HuffmanSizeAndIndex(src, index, 1, blog)

		if blog {
			log.Printf("bi: %d\n", bi)
		}

		var ts uint32
		// from huffman index
		if bi <= brHUFFMAXINDEXTABLE {
			if bi > 0 {
				newTS := buf2Sample(StU32, src, index, uint(bi))
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
		}

		(*theseries).Series[currentser].Samples = append((*theseries).Series[currentser].Samples, NkeSample{Timestamp: ts})

	}

	*lastTS = *absTS
	if blog {
		log.Printf("currentts %d \n", *currentts)
	}

	// Get measure (uncompressed)
	(*theseries).Series[currentser].Samples[0].Sample = buf2Sample((*theseries).Series[currentser].Params.Type, src, index,
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
		// TODO should manage sign unsigned
		if blog {
			log.Printf("First measure %d \n", (*theseries).Series[currentser].Samples[0].Sample)
		}
	}

	if (*theseries).nosample == 0 {
		parseCoding(src, theseries, index, currentser, blog)
	}

	return 0
}

func getSeparatedMeasures(src []byte, theseries *NkeSeries, index *uint, nbType int, currentser int, absTS *uint32, lastTS *uint32, currentts *uint32, blog bool) int {

	//number of samples

	nbsamples := buf2Sample(StU8, src, index, mapTypeSize[StU8])
	if blog {
		log.Printf("nb samples %d \n", nbsamples)
	}

	if nbsamples > 0 {
		// get timestamp coding

		tscoding := buf2Sample(StU8, src, index, 2)
		if blog {
			log.Printf(" TimeStamp Coding: %d\n", tscoding)
		}

		// samples loop

		var currentTS uint32

		for j := 0; j < int(nbsamples); j++ {

			bi := but2HuffmanSizeAndIndex(src, index, tscoding, blog)
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

					newTS := buf2Sample(StU32, src, index, uint(bi))
					f := math.Pow(2, float64(bi))
					currentTS = newTS + currentTS + uint32(f) - 1
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
			bi = but2HuffmanSizeAndIndex(src, index, (*theseries).Series[currentser].codingTable, blog)

			if blog {
				log.Printf("bi: %d\n", bi)
			}

			// from huffman index
			if bi <= brHUFFMAXINDEXTABLE {
				if bi > 0 {
					value := buf2Sample(StU32, src, index, uint(bi))
					if blog {
						log.Printf("raw: %d\n", value)
					}

					// get last samples
					cur := len((*theseries).Series[currentser].Samples)
					if cur > 0 {

						(*theseries).Series[currentser].Samples[cur-1].Sample = value
						convertValue(theseries, currentser, bi, uint(cur-1), blog)
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

func getLastTimeStamp(src []byte, theseries *NkeSeries, index *uint, absTS *uint32, lastTS *uint32, blog bool) {
	// Time stamp of the sending
	if *absTS == 0 {
		(*theseries).Timestamp = buf2Sample(StU32, src, index, 32)
	} else {
		bi := but2HuffmanSizeAndIndex(src, index, 1, blog)
		if blog {
			log.Printf(" Final timestamp bi: %d\n", bi)
		}

		// from huffman index
		if bi <= brHUFFMAXINDEXTABLE {
			if bi > 0 {
				newTS := buf2Sample(StU32, src, index, uint(bi))
				f := math.Pow(2, float64(bi))
				(*theseries).Timestamp = newTS + *lastTS + uint32(f) - 1
			} else {
				(*theseries).Timestamp = *lastTS
			}
		} else {
			(*theseries).Timestamp = buf2Sample(StU32, src, index, 32)
		}

	}

	if blog {
		log.Printf("last timestamp %d \n", (*theseries).Timestamp)
	}

	// print last time stamp
}
