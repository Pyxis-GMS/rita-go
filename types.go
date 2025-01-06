package ritago

import (
	"time"
)

// REQUEST TYPES

type getCursorResponse struct {
	EventId string `json:"eventId"`
}

// CONFIG TYPES

type RitaConfig struct {
	Url    string
	ApiKey string
	//LogInConsole bool
}

// RESPONSE TYPES

type eventsResponse struct {
	Events []RitaEvent `json:"events"`
}

type RitaEvent struct {
	Id        string
	CreatedAt time.Time
	Data      any
}

// ERROR

type ritaError int

const (
	ChannelNotValid ritaError = iota
	ServerNotConfig
	ApikeyNotConfig
	JsonNotValid
	ServerUrlNotValid
	NotAuthorized
	Forbidden
	UnknownError
)

func (e ritaError) String() string {
	switch e {
	case ChannelNotValid:
		return "the channel name is not valid"
	case ServerNotConfig:
		return "the server url is not setted"
	case ApikeyNotConfig:
		return "the apikey is not setted"
	case JsonNotValid:
		return "the object sent is not a json"
	case ServerUrlNotValid:
		return "the server url is not valid"
	case NotAuthorized:
		return "not authorized"
	case Forbidden:
		return "Forbidden"
	case UnknownError:
		return "forbidden"
	default:
		return "unknown error"
	}
}

func (e ritaError) Error() string {
	return e.String()
}
