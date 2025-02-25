package alamos

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/imroc/req"
)

// A Message, Unit and Data represents an alert which we send to FE2
type Message struct {
	Type          string `json:"type"`
	Timestamp     string `json:"timestamp"`
	Sender        string `json:"sender"`
	Authorization string `json:"authorization"`
	Data          Data   `json:"data"`
}
type Unit struct {
	Address string `json:"address"`
}
type Data struct {
	Message    []string `json:"message"`
	Keyword    string   `json:"keyword"`
	ExternalID string   `json:"externalId"`
	Units      []Unit   `json:"units"`
}

// Client represents the client configuration for a server
type Client struct {
	URL           string
	Sender        string
	Authorization string
	Test          bool
}

// A Response represents the answer we got from the FE2 server
type Response struct {
	Status string
	Error  string
}

// Function to write the output json to Alamos to a file
func writeDebugFile(body []byte, identifier string) {
	f, err := os.Create(fmt.Sprintf("/tmp/alert-alamos-%s.json", identifier))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	f.Write(body)
}

// Function to create a timestamp in ISO8601
func makeTimestamp() string {
	return time.Now().Format("2006-01-02T15:04:05-07:00")
}

// Create a new message struct with all client informations
func (c *Client) newMessage() Message {
	m := Message{
		Timestamp:     makeTimestamp(),
		Sender:        c.Sender,
		Authorization: c.Authorization,
	}
	if c.Test {
		m.Type = "TEST"
	} else {
		m.Type = "ALARM"
	}

	return m
}

// NewClient creates a new client to talk to an FE2 server
func NewClient(endpoint string, sender string, authorization string, test bool) Client {

	return Client{
		URL:           endpoint,
		Sender:        sender,
		Authorization: authorization,
		Test:          test,
	}
}

// SendAlert creates and sends an alert to the FE2 server
func (c *Client) SendAlert(data Data, debugIdentifier string, debug bool) error {
	message := c.newMessage()
	message.Data = data

	body, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Couldn't turn message into JSON")
		return err
	}

	if debug {
		writeDebugFile(body, debugIdentifier)
	}

	r, err := req.Post(c.URL, req.BodyJSON(body))
	if err != nil {
		fmt.Println(err)
		return err
	}

	response := Response{}
	err = r.ToJSON(&response)
	if err != nil {
		fmt.Printf("Alamos response status: %+v\n", r.Response().Body)
		fmt.Println("Couldn't parse response from FE2")
		return err
	}

	code := r.Response().StatusCode
	if code != 200 || response.Status != "OK" {
		if code == 400 {
			fmt.Printf("Errormessage was: %s", response.Error)
		}
		return fmt.Errorf("bad response from FE2")
	}

	fmt.Println("Sucessfully send alert to FE2")

	return nil
}
