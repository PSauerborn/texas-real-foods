package connectors

import (

)

// define interface for auto-updating of resources
type AutoUpdateDataConnector interface{
    // function used to collect data from connector source
    CollectData([]BusinessInfo) ([]BusinessInfo, error)
    Name() string
}