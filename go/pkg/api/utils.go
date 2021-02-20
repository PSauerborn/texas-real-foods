package api

import (
    "fmt"
    "errors"
    "time"
    "encoding/json"

    log "github.com/sirupsen/logrus"
    jsonpatch "github.com/evanphx/json-patch"
)

var (
    // define custom errors
    ErrInvalidPatch        = errors.New("Invalid JSON patch operation")
    ErrInvalidBusinessMeta = errors.New("Invalid business metadata")
    ErrInvalidStartTime    = errors.New("Invalid start timestamp")
    ErrInvalidEndTime      = errors.New("Invalid end timestamp")
    ErrInvalidTimeRange    = errors.New("Invalid time range")
)

func PatchBusinessMeta(business BusinessInfo,
    operation []map[string]interface{}) (map[string]interface{}, error) {

    patchJson, err := json.Marshal(operation)
    if err != nil {
        log.Error(fmt.Errorf("unable to convert patch operation to JSON: %+v", err))
        return map[string]interface{}{}, ErrInvalidPatch
    }

    // decode JSON patch operation
    patch, err := jsonpatch.DecodePatch(patchJson)
    if err != nil {
        log.Error(fmt.Errorf("unable to parse Json Patch operation: %+v", err))
        return map[string]interface{}{}, ErrInvalidPatch
    }

    // convert metadata to json
    var metaJson []byte
    if business.Metadata == nil {
        // set metadata to empty JSON string if not exists
        metaJson = []byte(`{}`)
    } else {
        metaJson, err = json.Marshal(business.Metadata)
        if err != nil {
            log.Error(fmt.Errorf("unable to convert business meta to JSON: %+v", err))
            return map[string]interface{}{}, ErrInvalidBusinessMeta
        }
    }

    // apply JSON patch operation
    modified, err := patch.Apply(metaJson)
    if err != nil {
        log.Error(fmt.Errorf("unable to apply JSON patch: %+v", err))
        return map[string]interface{}{}, ErrInvalidPatch
    }

    log.Debug(fmt.Sprintf("successfully applied JSON patch to metadata: %s", modified))
    // convert final JSON string back to interface
    var meta map[string]interface{}
    if err := json.Unmarshal(modified, &meta); err != nil {
        return meta, ErrInvalidBusinessMeta
    }
    return meta, nil
}

// function used to group timeseries data elements by source
func GroupTimeseriesDataBySource(data []TimeSeriesData) (map[string][]TimeSeriesData) {
    grouped := map[string][]TimeSeriesData{}
    for _, element := range(data) {
        // if source is already present in mapping, add to current values
        _, ok := grouped[element.Source]; if ok {
            grouped[element.Source] = append(grouped[element.Source], element)
        } else {
            // add new array of values to map
            grouped[element.Source] = []TimeSeriesData{element}
        }
    }
    return grouped
}


type TimeRange struct {
    Start time.Time
    End time.Time
}

// function to parse start and end timestamps in DD-MM-YYYY format
func ParseTimeRange(start, end string) (TimeRange, error) {

    format := "2006-01-02T15:04"
    // parse start timestamp
    startTimestamp, err := time.Parse(format, start)
    if err != nil {
        log.Error(fmt.Errorf("unable to parse start timestamp: %+v", err))
        return TimeRange{}, ErrInvalidStartTime
    }

    // parse end timestamp
    endTimestamp, err := time.Parse(format, end)
    if err != nil {
        log.Error(fmt.Errorf("unable to parse end timestamp: %+v", err))
        return TimeRange{}, ErrInvalidEndTime
    }

    // return error if start time is after end time
    if endTimestamp.Before(startTimestamp) {
        return TimeRange{}, ErrInvalidTimeRange
    }
    timeRange := TimeRange{startTimestamp, endTimestamp}
    return timeRange, nil
}