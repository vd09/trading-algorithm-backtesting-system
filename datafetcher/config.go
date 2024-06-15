package datafetcher

func getAPIKey() string {
	apiKey := "oNxMGkW82EWM7tpAzqyPnto2wDsJ37rs"
	//if apiKey == "" {
	//	apiKey = "your_default_api_key" // default API key if not set
	//}
	return apiKey
}

func getAPIUrl() string {
	apiUrl := "https://api.polygon.io/v2/aggs/ticker"
	//if apiUrl == "" {
	//	apiUrl = "https://api.polygon.io/v2/aggs/ticker" // default API URL if not set
	//}
	return apiUrl
}
