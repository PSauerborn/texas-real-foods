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
    for _, entry := range(a) {
        if !stringSliceContains(b, entry) {
            return false
        }
    }
    return true
}

func stringSliceContains(slice []string, element string) bool {
    for _, entry := range(slice) {
        if entry == element {
            return true
        }
    }
    return false
}

// function used to determine if timeseries entries differ.
// both a boolean value indicating if fields have changes
// is returned as well as a list of changed fields
func timeSeriesEntriesDiffer(a, b api.TimeseriesDataEntry) (bool, []string) {
    changedFields := []string{}
    // check if website live entries match
    if a.WebsiteLive != b.WebsiteLive {
        changedFields = append(changedFields, "website_live")
    }
    // check if phone numbers have changed
    if !stringSliceEqual(a.BusinessPhones, b.BusinessPhones) {
        changedFields = append(changedFields, "phonenumber")
    }
    // check if business open has changed
    if a.BusinessOpen != b.BusinessOpen {
        changedFields = append(changedFields, "business_open")
    }
    return len(changedFields) > 0, changedFields
}

// function used to generate notification hash. notifications hashes are
// generate as a combination of business ID, source and the current date
// to ensure that one unique notification is sent per business, per source
//  per day
func generateNotificationHash(businessId uuid.UUID, source string) string {
    notifyString := fmt.Sprintf("%s:%s:%s", businessId,
        source, time.Now().Format("01-02-2006"))
    notificationHash := sha256.Sum256([]byte(notifyString))
    return hex.EncodeToString(notificationHash[0:])
}