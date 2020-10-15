package alamos

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/imroc/req"
)

type AlamosMessage struct {
	Message   string            `json:"message"`
	Type      string            `json:"type"`
	Sender    string            `json:"sender"`
	Timestamp int64             `json:"timestamp"`
	Data      map[string]string `json:"data"`
}

type AlamosClient struct {
	URL     string
	Sender  string
	Address string
	Test    bool
}

type AlamosResponse struct {
	Status string
	Error  string
}

const RestURL = "/rest/external/http"

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func (c *AlamosClient) newAlamosMessage() AlamosMessage {
	m := AlamosMessage{
		Timestamp: makeTimestamp(),
	}
	if c.Test {
		m.Type = "Test"
	} else {
		m.Type = "ALARM"
	}

	return m
}

func NewClient(host string, sender string, address string, test bool) AlamosClient {

	return AlamosClient{
		URL:     fmt.Sprintf("%s%s", host, RestURL),
		Sender:  sender,
		Address: address,
		Test:    test,
	}
}

func (client *AlamosClient) SendAlert(alertMessage string) error {
	message := client.newAlamosMessage()
	message.Sender = client.Sender
	message.Message = alertMessage

	body, err := json.Marshal(message)
	if err != nil {
		log.Println("Couldn't turn message into JSON")
		return err
	}

	resp, err := req.Post(client.URL, body)
	if err != nil {
		log.Println(err)
		return err
	}

	data := AlamosResponse{}
	err = resp.ToJSON(&data)
	if err != nil {
		log.Println("Couldn't parse response from FE2")
		return err
	}

	code := resp.Response().StatusCode
	if code != 200 || data.Status != "OK" {
		log.Printf("Got an error from FE2: %d (%s)", code, data.Status)
		return fmt.Errorf("bad response from FE2")
	}

	return nil
}
