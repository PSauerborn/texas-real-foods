package timeseries_analyser

import (
    "fmt"
    "time"
    "crypto/sha256"
    "encoding/hex"

    "github.com/google/uuid"

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

// function used to generate notification hash
func generateNotificationHash(businessId uuid.UUID, source string) string {
    notifyString := fmt.Sprintf("%s:%s:%s", businessId,
        source, time.Now().Format("01-02-2006"))
    notificationHash := sha256.Sum256([]byte(notifyString))
    return hex.EncodeToString(notificationHash[0:])
}