package mail_relay

import (
    "io"
    "fmt"
    "encoding/json"

    log "github.com/sirupsen/logrus"
)

func parseZipCodeResponse(data io.ReadCloser) (ZipCodeDataResponse, error) {
    var response ZipCodeDataResponse
    if err := json.NewDecoder(data).Decode(&response); err != nil {
        log.Error(fmt.Sprintf("unable to parse API response into struct: %+v", err))
        return ZipCodeDataResponse{}, ErrInvalidJSONResponse
    }
    return response, nil
}

