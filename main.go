package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/codemonauts/alertmanager-fe2-bridge/alamos"
	"github.com/google/uuid"
	"github.com/prometheus/alertmanager/template"
	"gopkg.in/yaml.v2"
)

// Config represents the config yaml structure
type config struct {
	AlamosEndpoint string `yaml:"alamos_endpoint"`
	Sender         string `yaml:"alamos_sender"`
	Address        string `yaml:"alamos_address"`
	Authorization  string `yaml:"alamos_authorization"`
	Debug          bool   `yaml:"debug"`
	Listen         string `yaml:"listen"`
}

var (
	configPath string
)

// Function to write the input json from Prometheus to a file
func writeDebugFile(body []byte, identifier string) {
	f, err := os.Create(fmt.Sprintf("/tmp/alert-prometheus-%s.json", identifier))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err = f.Write(body)
	if err != nil {
		log.Fatal(err)
	}
}

// Helper function to create a simple md5 hash for unique external IDs for Alamos alarms
func hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

// Function handling the incomming request from Prometheus
func inputHandler(client *alamos.Client, address string, debug bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		body, _ := io.ReadAll(r.Body)
		data := template.Data{}

		debugIdentifier := uuid.NewString()

		if debug {
			writeDebugFile(body, debugIdentifier)
		}

		if err := json.Unmarshal(body, &data); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, err = w.Write([]byte(err.Error()))
			if err != nil {
				log.Fatal(err)
			}
			return
		}

		for _, alert := range data.Alerts {
			if alert.Status == "firing" || alert.Status == "resolved" {
				severity := alert.Labels["severity"]
				switch strings.ToUpper(severity) {
				case "PAGE":
					var alarmData alamos.Data

					// Prefix keyword when incident is resolved
					var keywordPrefix string = ""
					if alert.Status == "resolved" {
						keywordPrefix = "Resolved: "
					}

					// Set alert summary to the alarm keyword
					summary, ok := alert.Annotations["summary"]
					if ok {
						alarmData.Keyword = keywordPrefix + summary
					} else {
						alarmData.Keyword = keywordPrefix + alert.Labels["alertname"]
					}

					// Set alert message
					description, ok := alert.Annotations["description"]
					if ok {
						alarmData.Message = append(alarmData.Message, description)
					} else {
						alarmData.Message = append(alarmData.Message, fmt.Sprintf("Instance: %s", alert.Labels["instance"]))
					}

					// Set alert fingerprint to alarm external ID
					if alert.Fingerprint != "" {
						alarmData.ExternalID = alert.Fingerprint
					} else {
						alarmData.ExternalID = hash(alert.StartsAt.Format("20060102150405") + alert.Labels["alertname"] + alert.Labels["instance"])
					}

					// Set unit to send alarm to
					var unit alamos.Unit
					unit.Address = address
					alarmData.Units = append(alarmData.Units, unit)

					err := client.SendAlert(alarmData, debugIdentifier, debug)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						_, err = w.Write([]byte("ERROR"))
						if err != nil {
							log.Fatal(err)
						}
						return
					}

				default:
					fmt.Printf("no action on severity: %s", severity)
				}

			}
		}
		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.Fatal(err)
		}
	})
}

// Function to read the config file into the struct
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
	client := alamos.NewClient(cfg.AlamosEndpoint, cfg.Sender, cfg.Authorization, cfg.Debug)

	if cfg.Debug {
		fmt.Printf("Alamos Client: %+v\n", client)
	}

	http.Handle("/input", inputHandler(&client, cfg.Address, cfg.Debug))

	fmt.Printf("Listening on %q\n", cfg.Listen)
	log.Fatal(http.ListenAndServe(cfg.Listen, nil))
}
