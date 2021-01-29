package syncer

import (
    "fmt"
    "errors"
    "crypto/sha256"
    "encoding/json"
    "encoding/hex"

    log "github.com/sirupsen/logrus"
)

var (
    ErrInvalidJSONFormat = errors.New("Invalid JSON format")
)

// function used to hash a map of values
func HashMap(values interface{}) (string, error) {
    // convert to JSON string for hashing
    jsonString, err := json.Marshal(values)
    if err != nil {
        log.Error(fmt.Errorf("unable to convert map to JSON string: %+v", err))
        return "", ErrInvalidJSONFormat
    }
    // hash JSON string and return
    sum := sha256.Sum256(jsonString)
    return hex.EncodeToString(sum[0:]), nil
}