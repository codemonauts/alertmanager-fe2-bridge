package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/codemonauts/alertmanager-fe2-bridge/alamos"
)

func TestInputHandler(t *testing.T) {
	file, err := os.Open("./testdata.json")
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/input", file)
	if err != nil {
		t.Fatal(err)
	}

	server := alarmosMock()
	defer server.Close()

	client := alamos.NewClient(server.URL, "", "", true)
	debug := true

	rr := httptest.NewRecorder()
	handler := inputHandler(&client, "", debug)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func alarmosMock() *httptest.Server {
	handler := http.NewServeMux()
	handler.HandleFunc("/rest/external/http/v2", inputMock)
	srv := httptest.NewServer(handler)

	return srv
}

func inputMock(w http.ResponseWriter, r *http.Request) {
	resp := alamos.Response{
		Status: "OK",
	}
	data, _ := json.Marshal(resp)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
