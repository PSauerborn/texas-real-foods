package notifications

import (

)


type ChangeNotification struct{}

// define interface for engine
type NotificationEngine interface{
	SendNotification(notification ChangeNotification) error
}