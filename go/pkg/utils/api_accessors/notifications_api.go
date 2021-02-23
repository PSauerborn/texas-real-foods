package utils

import (
    "fmt"
    "bytes"
    "errors"
    "encoding/json"

    "github.com/google/uuid"
    log "github.com/sirupsen/logrus"

    "texas_real_foods/pkg/utils"
    "texas_real_foods/pkg/notifications"
)

var (
    ErrNotificationAlreadyExists = errors.New("Notification hash already exists")
)

type NotificationsAPIAccessor struct {
    *utils.BaseAPIAccessor
}

// function to generate new API accessor for Texas Real Foods API
func NewNotificationsApiAccessor(host, protocol string, port *int) *NotificationsAPIAccessor {
    baseAccessor := utils.BaseAPIAccessor{
        Host: host,
        Port: port,
        Protocol: protocol,
    }
    return &NotificationsAPIAccessor{
        &baseAccessor,
    }
}

// function to generate new API accessor for Texas Real Foods API
func NewNotificationsApiAccessorFromConfig(config utils.APIDependencyConfig) *NotificationsAPIAccessor {
    baseAccessor := utils.NewAPIAccessorFromConfig(config)
    return &NotificationsAPIAccessor{
        baseAccessor,
    }
}

type Notification struct {
    NotificationId uuid.UUID                        `json:"notification_id"`
    Notification   notifications.ChangeNotification `json:"notification"`
}

type NotificationsResponse struct {
    HTTPCode      int 			 `json:"http_code"`
    Count         int            `json:"count"`
    Notifications []Notification `json:"notifications"`
}

// function to retrieve all notifications from API
func(accessor *NotificationsAPIAccessor) GetNotifications() ([]Notification, error) {
    log.Debug("retrieving all notifications from notifications API...")
    var response NotificationsResponse
    url := accessor.FormatURL("notifications/all")

    // generate new JSON request and execute
    req, err := accessor.NewJSONRequest("GET", url, nil, nil)
    if err != nil {
        log.Error(fmt.Errorf("unable to create request: %+v", err))
        return response.Notifications, err
    }
    resp, err := accessor.ExecuteRequest(req)
    if err != nil {
        log.Error(fmt.Errorf("unable to execute API request: %+v", err))
        return response.Notifications, err
    }
    defer resp.Body.Close()

    // handle response based on code
    switch resp.StatusCode {
    case 200:
        log.Debug(fmt.Sprintf("successfully retrieved notifications"))
        // decode JSON response from API and return
        if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
            log.Error(fmt.Errorf("unable to parse JSON response from API: %+v", err))
            return response.Notifications, err
        }
        return response.Notifications, nil
    default:
        log.Error(fmt.Sprintf("failed retrieve notifications: received invalid %d response from API",
            resp.StatusCode))
        return response.Notifications, utils.ErrInvalidAPIResponse
    }
}

// function to retrieve all unread notifications from API
func(accessor *NotificationsAPIAccessor) GetUnreadNotifications() ([]Notification, error) {
    log.Debug("retrieving all unread notifications from notifications API...")
    var response NotificationsResponse
    url := accessor.FormatURL("notifications/unread")

    // generate new JSON request and execute
    req, err := accessor.NewJSONRequest("GET", url, nil, nil)
    if err != nil {
        log.Error(fmt.Errorf("unable to create request: %+v", err))
        return response.Notifications, err
    }
    resp, err := accessor.ExecuteRequest(req)
    if err != nil {
        log.Error(fmt.Errorf("unable to execute API request: %+v", err))
        return response.Notifications, err
    }
    defer resp.Body.Close()

    // handle response based on code
    switch resp.StatusCode {
    case 200:
        log.Debug(fmt.Sprintf("successfully retrieved notifications"))
        // decode JSON response from API and return
        if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
            log.Error(fmt.Errorf("unable to parse JSON response from API: %+v", err))
            return response.Notifications, err
        }
        return response.Notifications, nil
    default:
        log.Error(fmt.Sprintf("failed retrieve notifications: received invalid %d response from API",
            resp.StatusCode))
        return response.Notifications, utils.ErrInvalidAPIResponse
    }
}

type HTTPMessageResponse struct {
    HTTPCode int    `json:"http_code"`
    Message  string `json:"message"`
}

func(accessor *NotificationsAPIAccessor) UpdateNotification(notificationId uuid.UUID) (
    HTTPMessageResponse, error) {
    log.Debug(fmt.Sprintf("updating notification %s...", notificationId))

    // construct URL using parameters
    url := accessor.FormatURL(fmt.Sprintf("notifications/update/%s", notificationId))
    var response HTTPMessageResponse
    // generate new JSON request and execute
    req, err := accessor.NewJSONRequest("PATCH", url, nil, nil)
    if err != nil {
        log.Error(fmt.Errorf("unable to create request: %+v", err))
        return response, err
    }
    resp, err := accessor.ExecuteRequest(req)
    if err != nil {
        log.Error(fmt.Errorf("unable to execute API request: %+v", err))
        return response, err
    }
    defer resp.Body.Close()

    // handle response based on code
    switch resp.StatusCode {
    case 200:
        log.Debug(fmt.Sprintf("successfully updated notification"))
        // decode JSON response from API and return
        if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
            log.Error(fmt.Errorf("unable to parse JSON response from API: %+v", err))
            return response, err
        }
        return response, nil
    default:
        log.Error(fmt.Sprintf("failed to update notification: received invalid %d response from API",
            resp.StatusCode))
        return response, utils.ErrInvalidAPIResponse
    }
}

// function used to create new notification
func(accessor *NotificationsAPIAccessor) CreateNotification(notification notifications.ChangeNotification) (
    HTTPMessageResponse, error) {
    log.Debug("creating new notification...")
    var response HTTPMessageResponse

    jsonBody, err := json.Marshal(notification)
    if err != nil {
        log.Error(fmt.Errorf("unable to convert notification to JSON format"))
        return response, utils.ErrInvalidRequestBodyJSON
    }

    // construct URL using parameters
    url := accessor.FormatURL("notifications/new")
    // generate new JSON request and execute
    req, err := accessor.NewJSONRequest("POST", url, bytes.NewBuffer(jsonBody), nil)
    if err != nil {
        log.Error(fmt.Errorf("unable to create request: %+v", err))
        return response, err
    }
    resp, err := accessor.ExecuteRequest(req)
    if err != nil {
        log.Error(fmt.Errorf("unable to execute API request: %+v", err))
        return response, err
    }
    defer resp.Body.Close()

    // handle response based on code
    switch resp.StatusCode {
    case 201:
        log.Debug(fmt.Sprintf("successfully created notification"))
        // decode JSON response from API and return
        if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
            log.Error(fmt.Errorf("unable to parse JSON response from API: %+v", err))
            return response, err
        }
        return response, nil
    case 409:
        log.Warn("unable to create notification: notification hash already exists")
        return response, ErrNotificationAlreadyExists
    default:
        log.Error(fmt.Sprintf("failed to create notification: received invalid %d response from API",
            resp.StatusCode))
        return response, utils.ErrInvalidAPIResponse
    }
}
