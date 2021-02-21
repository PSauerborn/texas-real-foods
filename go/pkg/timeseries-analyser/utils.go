package timeseries_analyser

import (
    api "texas_real_foods/pkg/utils/api_accessors"
)

func stringSliceEqual(a, b []string) bool {
    if len(a) != len(b) {
        return false
    }
    for i, v := range a {
        if v != b[i] {
            return false
        }
    }
    return true
}

// function used to determine if timeseries entries differ
func timeSeriesEntriesDiffer(a, b api.TimeseriesDataEntry) bool {
    return a.WebsiteLive != b.WebsiteLive || !stringSliceEqual(a.BusinessPhones, b.BusinessPhones) || a.BusinessOpen != b.BusinessOpen
}