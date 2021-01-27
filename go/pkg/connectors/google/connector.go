package connectors

import (

)

type GoogleAPIConnector struct{
	BaseAPIUrl string
	APIKey     string
}

func NewGoogleAPIConnector(baseUrl, apiKey string) *GoogleAPIConnector {
	return &GoogleAPIConnector{
		BaseAPIUrl: baseUrl,
		APIKey: apiKey,
	}
}
