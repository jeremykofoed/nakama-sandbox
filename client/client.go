package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"bytes"
	"io"
)

type Session struct {
	Token string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

//@JWK TODO: Stress test if you get time making concurrent calls.

//See link for Client API: https://heroiclabs.com/docs/nakama/api/client/
//You can also look at the nakama api explorer api call AuthenticationDevice to get payload setup example.
//http://localhost:7351/#/apiexplorer?endpoint=AuthenticateDevice
func main () {
	//API info.
	serverKey := "defaultkey"
	host := "localhost"
	port := 7350
	urlBase := fmt.Sprintf("http://%s:%d", host, port)
	
	//Get Session.
	session := AuthAPI(urlBase, serverKey)

	//RPC: Load game.
	response := LoadGameRPC(urlBase, session)
	fmt.Printf("Response: %+v\n", response)
}

//Get the session information which is a JWT token/refresh token.
func AuthAPI(urlBase string, serverKey string) Session {
	ctx := context.Background()
	deviceID := "device.1234.test"
	createAccount := false
	displayName := "Jeremy-Test"

	//Payload info.
	payload := map[string]interface{}{
		"id": deviceID,
	}
	pBody, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Error trying to marshal payload: %v", err)
	}

	//API call info
	endpoint := "/v2/account/authenticate/device"
	urlAPI := fmt.Sprintf("%s%s?create=%t&username=%s", urlBase, endpoint, createAccount, displayName)

	//Request setup.
	req, err := http.NewRequestWithContext(ctx, "POST", urlAPI, bytes.NewBuffer(pBody))
	if err != nil {
		log.Fatalf("Endpoint: %+v; Error: %v", urlAPI, err)
	}
	req.SetBasicAuth(serverKey, "") //
	req.Header.Set("Content-Type", "application/json")
	
	//Build client.
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer res.Body.Close()

	//Read response
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Response: %+v; Error: %v", res, err)
	}
	if res.StatusCode != http.StatusOK {
		log.Fatalf("Non 200 code, response: %+v; Endpoint: %s", res, urlAPI)
	}
	fmt.Printf("resBody: %s\n", resBody)

	var session Session
	json.Unmarshal(resBody, &session)
	
	return session
}

//Makes a call to the RPC to load game.
func LoadGameRPC(urlBase string, session Session) interface{} {
	ctx := context.Background()

	//API call info
	endpoint := "/v2/rpc/load_game"
	urlAPI := fmt.Sprintf("%s%s", urlBase, endpoint)

	//Request setup.
	req, err := http.NewRequestWithContext(ctx, "POST", urlAPI, nil)
	if err != nil {
		log.Fatalf("Endpoint: %+v; Error: %v", urlAPI, err)
	}
	req.Header.Set("Authorization", "Bearer " + session.Token) //JWT
	req.Header.Set("Content-Type", "application/json")
	
	//Build client.
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer res.Body.Close()

	//Read response
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Response: %+v; Error: %v", res, err)
	}
	if res.StatusCode != http.StatusOK {
		log.Fatalf("Non 200 code, response: %+v; Endpoint: %s", res, urlAPI)
	}
	fmt.Printf("resBody: %s\n", resBody)

	var response interface{}
	json.Unmarshal(resBody, &response)
	
	return response
}