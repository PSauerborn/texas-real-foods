package parked_domain

import (
    "fmt"
    "time"
    "strings"
    "encoding/hex"
    "crypto/sha256"

    "github.com/google/uuid"
)

var (
    // define array condition(s) for parked domains
    ParkedDomainConditions = []func(body string) bool {
        GoDaddyDomainParked,
    }
)

// function used to check godaddy domain for parked domain.
// note that this is done by checking for a particular substring
func GoDaddyDomainParked(body string) bool {
    // convert to lowercase
    body = strings.ToLower(body)
    // check for common markers/messages
    parkedMessage := strings.Contains(body, "this web page is parked free, courtesy of godaddy")
    brokerMessage := strings.Contains(body, "our domain broker service may be able to get it for you")
    return parkedMessage || brokerMessage
}

// function used to generate notification hash
func generateNotificationHash(businessId uuid.UUID) string {
    notifyString := fmt.Sprintf("%s:%s", businessId, time.Now().Format("01-02-2006"))
    notificationHash := sha256.Sum256([]byte(notifyString))
    return hex.EncodeToString(notificationHash[0:])
}


