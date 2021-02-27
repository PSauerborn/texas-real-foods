package notifications

import (
	"fmt"
	"strings"
	"errors"

	log "github.com/sirupsen/logrus"
)

var (
	ErrInvalidFilterCondition = errors.New("Invalid filter condition")
)

// function used to check if a given notification metadata
// matches a given filter string. Conditions are passed
// as comma separated key:val pairs i.e. a valid string is
func MetadataMatchesFilter(meta map[string]interface{}, filter string) (bool, error) {
	// generate boolean array to store condition results
	matches := []bool{}
	for _, condition := range(strings.Split(filter, ",")) {
		// separate each filter condition into key val pairs
		keyValPair := strings.Split(condition, ":")
		if len(keyValPair) != 2 {
			log.Error(fmt.Errorf("received invalid filter condition %s", condition))
			return false, ErrInvalidFilterCondition
		}

		// check if key is in metadata
		if val, ok := meta[keyValPair[0]]; ok {
			// check that value of key matches filter condition
			if val == keyValPair[1] {
				matches = append(matches, true)
			} else {
				matches = append(matches, false)
			}
		} else {
			matches = append(matches, false)
		}
	}

	// only return as match if all conditions are satisfied
	for _, match := range(matches) {
		if !match {
			return false, nil
		}
	}
	return true, nil
}

// function used to filter notifications on a given
// metadata argument
func FilterNotificationsByMetadata(notifications []Notification,
	filter string) ([]Notification, error) {
	log.Debug(fmt.Sprintf("filtering %d notifications on filter %s", len(notifications), filter))
	filtered := []Notification{}
	for _, ele := range(notifications) {
		matches, err := MetadataMatchesFilter(ele.Notification.Metadata, filter)
		if err != nil {
			log.Error(fmt.Errorf("unable to filter notifications: %+v", err))
			return filtered, err
		} else if matches {
			filtered = append(filtered, ele)
		}
	}
	return filtered, nil
}