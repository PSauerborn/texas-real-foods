package syncer

import (

)

type ReducedBusinessData struct{
    WebsiteLive    bool     `json:"website_live"`
    BusinessPhones []string `json:"business_phones"`
}