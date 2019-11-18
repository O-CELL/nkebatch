//Copyright  © O-CELL 2018 contact@o-cell.fr
//This source is released under the Apache License 2.0
//which can be found in LICENSE.txt
package nkebatch

import (
	"fmt"
	"log"
)

// HUFFMAN TABLE
const nbHUFFELEMENT = 16
const brHUFFMAXINDEXTABLE = 14
const brHUFFSIZEMAX = 11

// structure for Huffman Table

type symbol struct {
	Length uint8  //	size of the label
	Label  uint16 //	Label
}

type codebook []symbol

/* C syntax
const  huff [3][NB_HUFF_ELEMENT] huffmanNodes = {
	{{ 2,0x000},{ 2,0x001},{ 2,0x003},{ 3,0x005},{ 4,0x009},{ 5,0x011},{ 6,0x021},{ 7,0x041},{ 8,0x081},{10,0x200},{11,0x402},{11,0x403},{11,0x404},{11,0x405},{11,0x406},{11,0x407}},
	{{ 7,0x06f},{ 5,0x01a},{ 4,0x00c},{ 3,0x003},{ 3,0x007},{ 2,0x002},{ 2,0x000},{ 3,0x002},{ 6,0x036},{ 9,0x1bb},{ 9,0x1b9},{10,0x375},{10,0x374},{10,0x370},{11,0x6e3},{11,0x6e2}},
	{{ 4,0x009},{ 3,0x005},{ 2,0x000},{ 2,0x001},{ 2,0x003},{ 5,0x011},{ 6,0x021},{ 7,0x041},{ 8,0x081},{10,0x200},{11,0x402},{11,0x403},{11,0x404},{11,0x405},{11,0x406},{11,0x407}}
}
*/
var dictionaries = []codebook{{symbol{2, 0x000}, // book0
	symbol{2, 0x001},
	symbol{2, 0x003},
	symbol{3, 0x005},
	symbol{4, 0x009},
	symbol{5, 0x011},
	symbol{6, 0x021},
	symbol{7, 0x041},
	symbol{8, 0x081},
	symbol{10, 0x200},
	symbol{11, 0x402},
	symbol{11, 0x403},
	symbol{11, 0x404},
	symbol{11, 0x405},
	symbol{11, 0x406},
	symbol{11, 0x407},
}, {symbol{7, 0x06f}, //book 1
	symbol{5, 0x01a},
	symbol{4, 0x00c},
	symbol{3, 0x003},
	symbol{3, 0x007},
	symbol{2, 0x002},
	symbol{2, 0x000},
	symbol{3, 0x002},
	symbol{6, 0x036},
	symbol{9, 0x1bb},
	symbol{9, 0x1b9},
	symbol{10, 0x375},
	symbol{10, 0x374},
	symbol{10, 0x370},
	symbol{11, 0x6e3},
	symbol{11, 0x6e2},
}, {symbol{4, 0x009}, // book 2
	symbol{3, 0x005},
	symbol{2, 0x000},
	symbol{2, 0x001},
	symbol{2, 0x003},
	symbol{5, 0x011},
	symbol{6, 0x021},
	symbol{7, 0x041},
	symbol{8, 0x081},
	symbol{10, 0x200},
	symbol{11, 0x402},
	symbol{11, 0x403},
	symbol{11, 0x404},
	symbol{11, 0x405},
	symbol{11, 0x406},
	symbol{11, 0x407},
},
}

// buf2HuffmanSizeAndIndex ... Get the pattern from bitstream
func buf2HuffmanSizeAndIndex(src []byte, startBit *uint, codingtable uint32) (uint8, error) {
	index := *startBit

	for i := 2; i <= brHUFFSIZEMAX; i++ {
		lbl := buf2HuffmanPattern(src, index, uint16(i))
		if result, err := getHuffmanIndexFromPattern(uint8(i), lbl, codingtable) ; err == nil {
			if result != -1 {
				*startBit += uint(i)
				return uint8(result), nil
			}
		} else {
			return 0, err
		}
	}
	return 0, nil
}

//buf2HuffmanPattern retrieves nbbits from bit stream src starting at pos index and
//returns an uint16 containing the requested bits
func buf2HuffmanPattern(src []byte, index uint, nbbits uint16) uint16 {
	var pattern uint16
	var size = nbbits - 1
	var bittoread uint16

	// don t check the total size
	var idx int
	for nbbits > 0 {
		idx = int(index>>3)
		if idx >= len(src) {
			return pattern //TODO properly handle error
		}
		checkBit := (uint8(src[idx]) & uint8((1 << (index & 0x07))))
		if checkBit != 0 {
			pattern |= (1 << (size - bittoread))
		}

		nbbits--
		bittoread++
		index++
	}
	return pattern
}

//getHuffmanIndexFromPattern searches for label lbl of size size into Huffman coding table with index codingtable
//returns the index of the pattern in the Huffman table or -1 if not found
func getHuffmanIndexFromPattern(size uint8, lbl uint16, codingtable uint32) (idx int, err error) {
	if int(codingtable) >= len(dictionaries) {
		if blog {
			log.Printf("Invalid coding table id %d", codingtable)
		}
		return -1, fmt.Errorf("invalid coding table id %d", codingtable)
	}
	for j := 0; j < nbHUFFELEMENT; j++ {
		if (dictionaries[codingtable][j].Label == lbl) && (dictionaries[codingtable][j].Length == size) {
			if blog {
				log.Printf("label %d size %d \n", lbl, size)
			}
			return j, nil
		}
	}
	return -1, nil
}
