package model

import "sort"

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

func (pr *PolygonResponse) MergeResponse(newData *PolygonResponse) {
	combined := make(map[int64]DataPoint)

	for _, result := range pr.Results {
		combined[result.Time] = result
	}
	for _, result := range newData.Results {
		combined[result.Time] = result
	}

	var mergedResults []DataPoint
	for _, result := range combined {
		mergedResults = append(mergedResults, result)
	}

	sort.Slice(mergedResults, func(i, j int) bool {
		return mergedResults[i].Time < mergedResults[j].Time
	})

	pr.Results = mergedResults
}
