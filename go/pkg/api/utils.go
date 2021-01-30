package api

import (
	"fmt"
	"errors"
	"encoding/json"

	log "github.com/sirupsen/logrus"
    jsonpatch "github.com/evanphx/json-patch"
)

var (
	ErrInvalidPatch        = errors.New("Invalid JSON patch operation")
	ErrInvalidBusinessMeta = errors.New("Invalid business metadata")
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