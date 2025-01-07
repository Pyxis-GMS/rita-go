package ritago_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	ritago "github.com/Pyxis-GMS/rita-go"
)

type env struct {
	Url    string `json:"url"`
	ApiKey string `json:"apiKey"`
}

var ritaConfig *ritago.RitaConfig
var client *ritago.RitaClient

func init() {
	file, _ := os.Open("env.test.json")
	defer file.Close()
	decoder := json.NewDecoder(file)

	env := env{}
	err := decoder.Decode(&env)
	if err != nil {
		panic(err)
	}

	ritaConfig = &ritago.RitaConfig{
		Url:    env.Url,
		ApiKey: env.ApiKey,
	}

	client = ritago.NewRitaClient(ritaConfig)
}

/*
func TestGetEvents(t *testing.T) {
	events, err := client.GetEventsSince("test", "1736187360563-0")

	if err != nil {
		t.Error(err)
	}

	for _, event := range events {
		fmt.Println(event)
	}
}
*/

func TestSubEvent(t *testing.T) {
	channel := "test"

	events, _ := client.SubEvent(channel)

	for event := range events {
		fmt.Println(event)
	}
}

/*
func TestSendEvent(t *testing.T) {

	event := map[string]interface{}{
		"test": "test",
	}

	eventId, err := client.SendEvent("test", event)

	if err != nil {
		t.Error(err)
	}

	fmt.Println(eventId)
}
*/
/*
func TestGetCursor(t *testing.T) {
	cursor, err := client.GetCursor("test")

	if err != nil {
		t.Error(err)
	}

	fmt.Println(cursor)
}ç
*/
