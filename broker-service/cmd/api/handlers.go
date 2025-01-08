package main

import (
	"broker-service/logs"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net/http"
	"time"
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
	case "log":
		app.logItem(w, requestPayload.Log)
	default:
		fmt.Println("BadRequest during action cases")
		app.errorJSON(w, errors.New("invalid action"))
	}
}

func (app *Config) logItem(w http.ResponseWriter, entry logPayload) {
	jsonData, err := json.MarshalIndent(entry, "", "\t") //change to marshall??
	if err != nil {
		log.Println("Error during marshalling jsonData in log service")
		app.errorJSON(w, err)
		return
	}

	logServiceURL := "http://log-service:82/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))

	if err != nil {
		log.Println("Error during creating request in log service")
		app.errorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Println("Error during doing request in log service")
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		log.Println("Response status code is not accepted in log service")
		app.errorJSON(w, err)
		return
	}

	var payLoad jsonResponse
	payLoad.Error = false
	payLoad.Message = "logged"

	app.writeJSON(w, http.StatusAccepted, payLoad)
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

func (app *Config) LogViagRPC(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		fmt.Println("Error in broker-service/handlers, 166")
		app.errorJSON(w, err)
		return
	}

	conn, err := grpc.NewClient("log-service:50001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println("Error in broker-service/handlers, 173")
		app.errorJSON(w, err)
		return
	}
	defer conn.Close()

	c := logs.NewLogServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, err = c.WriteLog(ctx, &logs.LogRequest{
		LogEntry: &logs.Log{
			Name: requestPayload.Log.Name,
			Data: requestPayload.Log.Data,
		},
	})
	if err != nil {
		fmt.Println("Error in broker-service/handlers, 190")
		app.errorJSON(w, err)
		return
	}

	var payLoad jsonResponse
	payLoad.Error = false
	payLoad.Message = "logged"

	app.writeJSON(w, http.StatusAccepted, payLoad)
}
