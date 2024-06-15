package model

type DataPoint struct {
	Time   int64   `json:"t"`
	Open   float64 `json:"o"`
	High   float64 `json:"h"`
	Low    float64 `json:"l"`
	Close  float64 `json:"c"`
	Volume float64 `json:"v"`
}

type PolygonResponse struct {
	Ticker  string      `json:"ticker"`
	Results []DataPoint `json:"results"`
}
