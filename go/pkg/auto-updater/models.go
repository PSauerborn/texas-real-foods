package auto_updater

import (
    "texas_real_foods/pkg/connectors"
)

type ListBusinessResponse struct {
    HTTPCode int 		                   `json:"http_code"`
    Data     []connectors.BusinessMetadata `json:"data"`
}