# nkebatch
Library allowing to decode ZCL compressed "batch" frames sent by NKE IOT sensors

Check http://support.nke-watteco.com/#BatchReport for explanation about the different series parameters 

To use the library first initialize the decoder with the series parameters (serieParams contains an array of SerieParam which include Tag, Resolution and Type): 

```
var series nkebatch.NkeSeries
nkebatch.Initialize(&series, labelSize, serieParams, debug)
```
Then process the raw payload stored in `payload` variable ([]byte): 

```
if err := nkebatch.ProcessPayload(payload, &series) ; err != nil {
    return nil, fmt.Errorf("failed to process NKE batch frame %v : %v", payload, err)
}
````
  
It returns an array of Series containing each the Samples (deltaTimeStamp & value) for the given Series
