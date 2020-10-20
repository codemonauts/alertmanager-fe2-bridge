package alamos

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/imroc/req"
)

type AlamosMessage struct {
	Message   string            `json:"message"`
	Type      string            `json:"type"`
	Sender    string            `json:"sender"`
	Address   string            `json:"address"`
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

func NewClient(host string, sender string, address string, test bool) AlamosClient {

	return AlamosClient{
		URL:     fmt.Sprintf("%s%s", host, RestURL),
		Sender:  sender,
		Address: address,
		Test:    test,
	}
}

func (client *AlamosClient) SendAlert(alertMessage string, data map[string]string) error {
	message := client.newAlamosMessage()
	message.Message = alertMessage
	message.Data = data

	body, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Couldn't turn message into JSON")
		return err
	}

	r, err := req.Post(client.URL, req.BodyJSON(body))
	if err != nil {
		fmt.Println(err)
		return err
	}

	response := AlamosResponse{}
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
