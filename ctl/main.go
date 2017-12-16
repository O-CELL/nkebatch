package main

import (
	"encoding/json"
	"io/ioutil"
	"ocell/nkebatch/lib"
)

type config struct {
	Buf       []byte                `json:"buffer"`
	Labelsize uint                  `json:"labelsize"`
	Series    []nkebatch.SerieParam `json:"series"`
}

// Init series
func initFromFile(path string, cfg *config, series *nkebatch.NkeSeries) (err error) {

	raw, err := ioutil.ReadFile(path)

	if err != nil {
		// Cfg ...

		return err
	}

	json.Unmarshal(raw, cfg)

	nkebatch.Initialize(series, (*cfg).Labelsize, (*cfg).Series, false)

	return err
}

func main() {
	var cfg config
	var theseries nkebatch.NkeSeries

	err := initFromFile("../input/test.json", &cfg, &theseries)

	if err == nil {
		err = nkebatch.ProcessPayload(cfg.Buf, &theseries, false)

		if err != nil {
			err.Error()
		} else {
			nkebatch.Dump(theseries)
		}

	}
}
