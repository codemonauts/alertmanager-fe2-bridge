package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/codemonauts/alertmanager-fe2-bridge/alamos"
	"github.com/prometheus/alertmanager/template"
	"gopkg.in/yaml.v2"
)

type config struct {
	AlamosHost string `yaml:"alamos_host"`
	Sender     string `yaml:"alamos_sender"`
	Receiver   string `yaml:"alamos_receiver"`
	Debug      bool   `yaml:"debug"`
	Listen     string `yaml:"listen"`
}

var (
	configPath string
)

func writeDebugFile(body []byte) {
	timestamp := int32(time.Now().Unix())
	f, err := os.Create(fmt.Sprintf("/tmp/alert-%d.json", timestamp))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	f.Write(body)
}

func inputHandler(client *alamos.AlamosClient, debug bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)
		data := template.Data{}

		if debug {
			writeDebugFile(body)
		}

		if err := json.Unmarshal(body, &data); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		for _, alert := range data.Alerts {
			if alert.Status == "firing" {
				severity := alert.Labels["severity"]
				switch strings.ToUpper(severity) {
				case "PAGE":
					var alertText string

					summary, ok := alert.Annotations["summary"]
					if ok {
						alertText = summary
					} else {
						alertText = fmt.Sprintf("%s on %s", alert.Labels["alertname"], alert.Labels["instance"])
					}

					err := client.SendAlert(alertText)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte("ERROR"))
					}

				default:
					log.Printf("no action on severity: %s", severity)
				}

			}
		}
		w.Write([]byte("OK"))
	})
}

func readConfigFile() config {
	cfg := config{}
	f, err := os.Open(configPath)

	if err != nil {
		log.Fatalf("Couldn't open config file: %s", err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		log.Fatalf("Couldn't decode content of config file: %s", err)
	}

	return cfg
}

func init() {
	flag.StringVar(&configPath, "config", "./config.yaml", "Path to the config file")
}

func main() {
	flag.Parse()

	cfg := readConfigFile()
	client := alamos.NewClient(cfg.AlamosHost, cfg.Sender, cfg.Receiver, cfg.Debug)

	http.Handle("/input", inputHandler(&client, cfg.Debug))

	fmt.Printf("Listening on %q\n", cfg.Listen)
	log.Fatal(http.ListenAndServe(cfg.Listen, nil))
}
