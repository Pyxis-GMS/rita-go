package ritago

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type RitaClient struct {
	urlEventSend string
	urlEventSub  string
	urlGetCursor string

	server string
	apikey string
}

const LAST_EVENT = "$"

// NewRitaClient creates a new instance of RitaClient with the provided configuration.
//
// Parameters:
//   - config: A pointer to a RitaConfig struct containing the configuration for the client.
//
// Returns:
//   - *RitaClient: A pointer to the newly created RitaClient instance.
//
// Example:
//
//	config := &RitaConfig{
//	    Url: "https://example.com",
//	    ApiKey: "your-api-key",
//	}
//	client := ritago.NewRitaClient(config)
func NewRitaClient(config *RitaConfig) *RitaClient {
	urlEventSend := "/v1/event/$"
	urlEventSub := "/v1/event/$"
	urlGetCursor := "/v1/event/$/last"

	return &RitaClient{
		urlEventSend: urlEventSend,
		urlEventSub:  urlEventSub,
		urlGetCursor: urlGetCursor,
		server:       strings.TrimSpace(config.Url),
		apikey:       strings.TrimSpace(config.ApiKey),
		//LogInConsole: config.LogInConsole,
	}
}

// Return the last event id of the channel passed by parameter
//
// Parameters:
//   - channel: The name of the channel for which to retrieve the cursor.
//
// Returns:
//   - string: The cursor for the specified channel.
//   - error: An error if the request fails or the channel cannot be accessed.
//
/*
# Example

	...
	client := ritago.NewRitaClient(ritaConfig)

	cursor, err := client.GetCursor("test")

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(cursor)
	...
*/
func (c *RitaClient) GetCursor(channel string) (string, error) {
	channel, err := c.ensureCan(channel)
	if err != nil {
		return "", err
	}
	url, err := c.createUrl(channel, c.urlGetCursor, nil)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", c.apikey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var cursorResponse getCursorResponse

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		err = json.Unmarshal(body, &cursorResponse)
		if err != nil {
			return "", err
		}

		return cursorResponse.EventId, nil
	case 401:
		return "", NotAuthorized
	case 403, 404:
		return "", Forbidden
	default:
		return "", UnknownError
	}
}

// SendEvent sends an event to the specified channel with the provided data.
//
// Parameters:
//   - channel: The name of the channel to which the event will be sent.
//   - data: The data to be sent as the event. This May be any type that can be marshaled into JSON.
//
// Returns:
//   - string: The event ID of the sent event.
//   - error: An error if the request fails or the event cannot be sent.
//
// Example:
//
//	...
//	client := ritago.NewRitaClient(ritaConfig)
//
//	eventID, err := client.SendEvent("test", map[string]interface{}{"key": "value"})
//
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	fmt.Println(eventID)
//	...
func (c *RitaClient) SendEvent(channel string, data interface{}) (string, error) {
	channel, err := c.ensureCan(channel)
	if err != nil {
		return "", err
	}

	url, err := c.createUrl(channel, c.urlEventSend, nil)
	if err != nil {
		return "", err
	}

	_bytes, err := json.Marshal(data)
	if err != nil {
		return "", JsonNotValid
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(_bytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", c.apikey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var cursorResponse getCursorResponse

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		err = json.Unmarshal(body, &cursorResponse)
		if err != nil {
			return "", err
		}

		return cursorResponse.EventId, nil
	case 401:
		return "", NotAuthorized
	case 403, 404:
		return "", Forbidden
	default:
		return "", UnknownError
	}
}

/*
SubEvent returns a channel that will receive events from the specified channel.

Parameters:
  - channel: The name of the channel from which to receive events.

Returns:
  - chan *RitaEvent: A channel that will receive events from the specified channel.
  - error: An error if the request fails or the channel cannot be accessed.

# Example

	...
	client := ritago.NewRitaClient(ritaConfig)

	events, _ := client.SubEvent("test")
	for event := range events {
		fmt.Println(event)
	}
	...
*/
func (c *RitaClient) SubEvent(channel string) (chan *RitaEvent, error) {
	return c.SubEventSince(channel, "")
}

/*
SubEventSince returns a channel that will receive events from the specified channel starting from the specified event ID.

For subscribe to the channel in the last event readed, you should use LAST_EVENT constant as eventId.

Parameters:
  - channel: The name of the channel from which to receive events.
  - eventId: The ID of the event from which to start receiving events.

Returns:
  - chan *RitaEvent: A channel that will receive events from the specified channel.
  - error: An error if the request fails or the channel cannot be accessed.

# Example

	...
	client := ritago.NewRitaClient(ritaConfig)

	events, _ := client.SubEvent("test", "event-id")
	for event := range events {
		fmt.Println(event)
	}
	...
*/
func (c *RitaClient) SubEventSince(channel string, eventId string) (chan *RitaEvent, error) {
	channel, err := c.ensureCan(channel)
	if err != nil {
		return nil, err
	}

	queryParams := map[string]string{
		"eventId": "",
		"sub":     "true",
	}

	if strings.TrimSpace(eventId) != "" {
		queryParams["eventId"] = eventId
	}

	url, err := c.createUrl(channel, c.urlEventSub, &queryParams)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", c.apikey)
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Accept", "text/event-stream")

	client := &http.Client{}
	transport := &http.Transport{}
	transport.DisableCompression = true
	client.Transport = transport

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case 200:
		ch := make(chan *RitaEvent)

		reader := bufio.NewReader(resp.Body)
		if reader == nil {
			return nil, UnknownError
		}

		go func() {
			for {
				line, err := reader.ReadBytes('\n')
				if err != nil {
					fmt.Println(err)
					resp.Body.Close()
					close(ch)
					break
				}

				strLine := strings.TrimSpace(string(line))

				if strings.HasPrefix(strLine, "data:") {
					eventData := strings.TrimPrefix(strLine, "data:")
					eventData = strings.TrimSpace(eventData)

					if eventData == "" || eventData == "ping" {
						continue
					}

					var event RitaEvent
					err := json.Unmarshal([]byte(eventData), &event)

					if err != nil {
						fmt.Println(err)
						continue
					}

					ch <- &event
				}
			}
		}()

		return ch, nil
	case 401:
		return nil, NotAuthorized
	case 403, 404:
		return nil, Forbidden
	default:
		return nil, UnknownError
	}
}

/*
GetEvents returns a list of events from the specified channel.

Parameters:
  - channel: The name of the channel from which to receive events.

Returns:
  - []RitaEvent: A list of events from the specified channel.
  - error: An error if the request fails or the channel cannot be accessed.
*/
func (c *RitaClient) GetEvents(channel string) ([]RitaEvent, error) {
	return c.GetEventsSince(channel, "")
}

/*
GetEventsSince returns a list of events from the specified channel starting from the specified event ID.
For get since the last event readed in subscription, you should use LAST_EVENT constant as eventId.

Parameters:
  - channel: The name of the channel from which to receive events.
  - eventId: The ID of the event from which to start receiving events.

Returns:
  - []RitaEvent: A list of events from the specified channel.
  - error: An error if the request fails or the channel cannot be accessed.
*/
func (c *RitaClient) GetEventsSince(channel string, eventId string) ([]RitaEvent, error) {
	channel, err := c.ensureCan(channel)
	if err != nil {
		return make([]RitaEvent, 0), err
	}

	queryParams := map[string]string{
		"eventId": "",
		"sub":     "false",
	}

	if strings.TrimSpace(eventId) != "" {
		queryParams["eventId"] = eventId
	}

	url, err := c.createUrl(channel, c.urlEventSub, &queryParams)
	if err != nil {
		return make([]RitaEvent, 0), err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return make([]RitaEvent, 0), err
	}

	req.Header.Set("Authorization", c.apikey)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return make([]RitaEvent, 0), err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var r eventsResponse

		body, err := io.ReadAll(resp.Body)

		if err != nil {
			return make([]RitaEvent, 0), err
		}

		err = json.Unmarshal(body, &r)
		if err != nil {
			return make([]RitaEvent, 0), err
		}

		//fmt.Println(r.Events)

		return r.Events, nil
	case 401:
		return nil, NotAuthorized
	case 403, 404:
		return nil, Forbidden
	default:
		return nil, UnknownError
	}
}

func (c *RitaClient) ensureCan(channel string) (string, error) {
	channel = strings.TrimSpace(channel)
	channel = strings.ToLower(channel)

	if c.server == "" {
		return "", ServerNotConfig
	}

	if c.apikey == "" {
		return "", ApikeyNotConfig
	}

	if channel == "" {
		return "", ChannelNotValid
	}

	return channel, nil
}

func (c *RitaClient) createUrl(channel, _url string, queryParams *map[string]string) (string, error) {
	u, err := url.Parse(c.server)
	if err != nil {
		return "", err
	}

	if queryParams != nil {
		q := u.Query()
		for k, v := range *queryParams {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	regReplace := regexp.MustCompile(`([^:]\/)\/+`)
	_url = regReplace.ReplaceAllString(_url, "$1")

	u.Path = strings.Replace(_url, "$", channel, 1)

	return u.String(), nil
}
