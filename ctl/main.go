//Copyright  © O-CELL 2018 contact@o-cell.fr
//This source is released under the Apache License 2.0
//which can be found in LICENSE.txt
package main

import (
	"encoding/json"
	"io/ioutil"
	"o-cell/nkebatch/lib"
)

// Init series
func initFromFile(path string, cfg *nkebatch.Config, series *nkebatch.NkeSeries) (err error) {

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
	var cfg nkebatch.Config
	var theseries nkebatch.NkeSeries

	err := initFromFile("../input/test.json", &cfg, &theseries)

	if err == nil {
		err = nkebatch.ProcessPayload(cfg.Buf, &theseries)

		if err != nil {
			err.Error()
		} else {
			nkebatch.Dump(theseries)
		}

	}
}
