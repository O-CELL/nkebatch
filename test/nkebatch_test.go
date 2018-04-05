package nkebatch

import (
	"fmt"
	"o-cell/nkebatch/lib"
	"testing"
)

func decodeCheck(t *testing.T, config nkebatch.Config, expectedseries nkebatch.NkeSeries) {
	var theseries nkebatch.NkeSeries

	nkebatch.Initialize(&theseries, config.Labelsize, config.Series, false)
	err := nkebatch.ProcessPayload(config.Buf, &theseries)

	if err == nil {
		// Check the series time stamp
		if expectedseries.Timestamp != theseries.Timestamp {
			t.Errorf("Timestamp: expected %d, actual %d", expectedseries.Timestamp, theseries.Timestamp)
		}
		// loop the serie

		for i, ser := range theseries.Series {
			for j, samps := range ser.Samples {
				if samps.Timestamp != expectedseries.Series[i].Samples[j].Timestamp {
					t.Errorf("Timestamp: expected %d, actual %d", expectedseries.Series[i].Samples[j].Timestamp, samps.Timestamp)
				}
				if expectedseries.Series[i].Params.Type == nkebatch.StFL {
					if samps.Samplef != expectedseries.Series[i].Samples[j].Samplef {
						t.Errorf("Value float: expected %f, actual %f", expectedseries.Series[i].Samples[j].Samplef, samps.Samplef)
					}
				} else {
					if samps.Sample != expectedseries.Series[i].Samples[j].Sample {
						t.Errorf("Value: expected %d, actual %d", expectedseries.Series[i].Samples[j].Sample, samps.Sample)
					}
				}
			}
		}
	} else {
		t.Error("Processing failed")
	}
}

func TestDecoder(t *testing.T) {
	testCases := []struct {
		cfg    nkebatch.Config
		result nkebatch.NkeSeries
	}{
		// Test 1
		{cfg: nkebatch.Config{Buf: []byte{16, 39, 0, 128, 3, 147, 32, 24, 0, 128, 16, 129, 131, 7, 13, 69, 133, 16, 5},
			Labelsize: 3,
			Series: []nkebatch.SerieParam{{Tag: 2,
				Resolution: 1,
				Type:       12}}},
			result: nkebatch.NkeSeries{Series: []nkebatch.NkeSerie{
				{Samples: []nkebatch.NkeSample{
					{Timestamp: 1830, Samplef: 11},
					{Timestamp: 1845, Samplef: 13},
					{Timestamp: 1860, Samplef: 14},
					{Timestamp: 1875, Samplef: 21},
					{Timestamp: 1876, Samplef: 100},
				},
					Params: nkebatch.SerieParam{Tag: 2, Resolution: 1, Type: 12},
				}},
				Timestamp: 1944}},
		// Test 2
		{cfg: nkebatch.Config{Buf: []byte{38, 21, 0, 32, 224, 96, 1, 215, 30, 0, 0, 160, 101, 15},
			Labelsize: 1,
			Series: []nkebatch.SerieParam{{Tag: 0,
				Resolution: 1,
				Type:       10}, {Tag: 1,
				Resolution: 100,
				Type:       6}}},
			result: nkebatch.NkeSeries{Series: []nkebatch.NkeSerie{
				{Samples: []nkebatch.NkeSample{
					{Timestamp: 263, Sample: 45},
				},
					Params: nkebatch.SerieParam{Tag: 0, Resolution: 1, Type: 10},
				}, {Samples: []nkebatch.NkeSample{
					{Timestamp: 263, Sample: 3000},
				},
					Params: nkebatch.SerieParam{Tag: 1, Resolution: 100, Type: 6},
				}},
				Timestamp: 263}},
		// Test 4 flasho
		{cfg: nkebatch.Config{Buf: []byte{16, 32, 192, 34, 1, 4, 160, 33, 75, 53, 28, 180, 91, 22, 59, 137, 101, 183, 44, 118, 203, 98, 183, 44, 118, 203, 98, 183, 44, 246},
			Labelsize: 1,
			Series: []nkebatch.SerieParam{{Tag: 0,
				Resolution: 1,
				Type:       10}}},
			result: nkebatch.NkeSeries{Series: []nkebatch.NkeSerie{
				{Samples: []nkebatch.NkeSample{
					{Timestamp: 18221344, Sample: 874922},
					{Timestamp: 18221944, Sample: 874922},
					{Timestamp: 18223144, Sample: 874922},
					{Timestamp: 18223744, Sample: 874922},
					{Timestamp: 18224344, Sample: 874922},
					{Timestamp: 18224944, Sample: 874922},
					{Timestamp: 18225544, Sample: 874922},
					{Timestamp: 18226144, Sample: 874922},
				},
					Params: nkebatch.SerieParam{Tag: 0, Resolution: 1, Type: 10},
				}},
				Timestamp: 18226144}},
		// Test 5 flasho
		{cfg: nkebatch.Config{Buf: []byte{16, 39, 224, 98, 10, 19, 160, 33, 235, 181, 20, 180, 91, 214, 103, 183, 44, 160, 221, 178, 110, 187, 101, 5, 119, 203, 234, 174},
			Labelsize: 1,
			Series: []nkebatch.SerieParam{{Tag: 0,
				Resolution: 1,
				Type:       10}}},
			result: nkebatch.NkeSeries{Series: []nkebatch.NkeSerie{
				{Samples: []nkebatch.NkeSample{
					{Timestamp: 18305944, Sample: 874927},
					{Timestamp: 18306544, Sample: 874945},
					{Timestamp: 18307144, Sample: 875024},
					{Timestamp: 18307744, Sample: 875045},
					{Timestamp: 18308344, Sample: 875092},
					{Timestamp: 18308944, Sample: 875146},
				},
					Params: nkebatch.SerieParam{Tag: 0, Resolution: 1, Type: 10},
				}},
				Timestamp: 18308946}},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test %d ", i), func(t *testing.T) {
			decodeCheck(t, tc.cfg, tc.result)
		})
	}
}
