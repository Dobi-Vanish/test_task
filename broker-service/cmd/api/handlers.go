package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   authPayload `json:"auth,omitempty"`
	Log    logPayload  `json:"log,omitempty"`
}

type authPayload struct {
	Email string `json:"email"`
	Pass  string `json:"password"`
}

type logPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "OK",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		fmt.Println("BadRequest during reading requestpayload")
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)
	default:
		fmt.Println("BadRequest during action cases")
		app.errorJSON(w, errors.New("invalid action"))
	}
}

func (app *Config) authenticate(w http.ResponseWriter, a authPayload) {
	// create some json we'll send to the auth microservice
	jsonData, _ := json.MarshalIndent(a, "", "\t")

	// call the service
	request, err := http.NewRequest("POST", "http://auth-service:82/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("BadRquest during calling auth service")
		app.errorJSON(w, err)
		return
	}

	client := &http.Client{}
	fmt.Println(client)
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("BadRquest during doing new request in authenticate func", err)
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	//
	//
	if response.StatusCode == http.StatusUnauthorized {
		fmt.Println("Status Unathorized")
		app.errorJSON(w, errors.New("unauthorized"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		fmt.Println("error calling auth service")
		app.errorJSON(w, errors.New("error calling auth service"))
		return
	}

	var jsonFromService jsonResponse
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		fmt.Println("BadRquest during decoding response")
		app.errorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		fmt.Println("BadRquest is jsonFromService")
		app.errorJSON(w, err, http.StatusUnauthorized)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "OK"
	payload.Data = jsonFromService.Data

	app.writeJSON(w, http.StatusOK, payload)
}
