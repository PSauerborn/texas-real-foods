package connectors

import (

)

// define interface for auto-updating of resources. all connectors
// implement the CollectData() function, which ingests a collection
// of BusinessMetadata instances and returns a serious of BusinessUpdate
// items that contain updated business information
//
// Additionally, each connector should implement the Name() function,
// which is used to identify the connector and the data source in
// various places
type AutoUpdateDataConnector interface{
    // function used to collect data from connector source
    CollectData(businesses []BusinessMetadata) ([]BusinessUpdate, error)
    Name() string
}

type StreamedAutoUpdateDataConnector interface{
    // function used to collect data from connector source
    StreamData(updates chan BusinessUpdate, businesses []BusinessMetadata) error
    Name() string
}