package alamos

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/imroc/req"
)

// A Message represents an alert which we send to FE2
type Message struct {
	Message   string            `json:"message"`
	Type      string            `json:"type"`
	Sender    string            `json:"sender"`
	Address   string            `json:"address"`
	Timestamp int64             `json:"timestamp"`
	Data      map[string]string `json:"data"`
}

// Client represents the client configuration for a server
type Client struct {
	URL     string
	Sender  string
	Address string
	Test    bool
}

// A Response represents the answer we got from the FE2 server
type Response struct {
	Status string
	Error  string
}

// RestURL is the endpoint where we expect the input plugin
const RestURL = "/rest/external/http"

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func (c *Client) newMessage() Message {
	m := Message{
		Timestamp: makeTimestamp(),
		Sender:    c.Sender,
		Address:   c.Address,
	}
	if c.Test {
		m.Type = "TEST"
	} else {
		m.Type = "ALARM"
	}

	return m
}

// NewClient creates a new client to talk to an FE2 server
func NewClient(host string, sender string, address string, test bool) Client {

	return Client{
		URL:     fmt.Sprintf("%s%s", host, RestURL),
		Sender:  sender,
		Address: address,
		Test:    test,
	}
}

// SendAlert creates and sends an alert to the FE2 server
func (c *Client) SendAlert(alertMessage string, data map[string]string) error {
	message := c.newMessage()
	message.Message = alertMessage
	message.Data = data

	body, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Couldn't turn message into JSON")
		return err
	}

	r, err := req.Post(c.URL, req.BodyJSON(body))
	if err != nil {
		fmt.Println(err)
		return err
	}

	response := Response{}
	err = r.ToJSON(&response)
	if err != nil {
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
