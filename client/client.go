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

//@JWK TODO: Stress test if you get time making concurrent calls.

// Returned by Nakama Auth
type Session struct {
	Token string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

// Response represents the top-level API response
type Payload struct {
	Data string `json:"payload"`
}

// Payload contains the player_data field
type PlayerData struct {
	Data Player `json:"player_data"`
}

// PlayerData holds all player-related information
type Player struct {
	ID           string         `json:"id"`
	DisplayName  string         `json:"display_name"`
	Level        int            `json:"level"`
	Experience   int            `json:"experience"`
	Health       int            `json:"health"`
	Currency     []CurrencyItem `json:"currency"`
	StatusEffects []interface{}  `json:"status_effects"`
	BattleState  BattleState    `json:"battle_state"`
	BattleStats  map[string]interface{} `json:"battle_stats"`
	Attributes   map[string]interface{} `json:"attributes"`
	CreatedAt    int64          `json:"created_at"`
	UpdatedAt    int64          `json:"updated_at"`
}

// CurrencyItem represents a single currency entry
type CurrencyItem struct {
	Type   string `json:"type"`
	Amount int    `json:"amount"`
}

// BattleState contains battle-related data
type BattleState struct {
	Enemies map[string]Enemy `json:"enemies"`
}

// Enemy represents an enemy in the battle state
type Enemy struct {
	Type          string         `json:"type"`
	Health        int            `json:"health"`
	AttackModifier float64        `json:"attack_modifier"`
	StatusEffects []interface{}  `json:"status_effects"`
	Rewards       []RewardItem   `json:"rewards"`
}

// RewardItem represents a reward from an enemy
type RewardItem struct {
	Type   string `json:"type"`
	Amount int    `json:"amount"`
}

//See link for Client API: https://heroiclabs.com/docs/nakama/api/client/
//You can also look at the nakama api explorer api call AuthenticationDevice to get payload setup example.
//http://localhost:7351/#/apiexplorer?endpoint=AuthenticateDevice
func main () {
	fmt.Println("\n")
	//API info.
	serverKey := "defaultkey"
	host := "localhost"
	port := 7350
	urlBase := fmt.Sprintf("http://%s:%d", host, port)
	
	//Get Session.
	session := AuthAPI(urlBase, serverKey)
	fmt.Println("-------------------------------------------------------------\n")

	//RPC: Load game.
	var targetID string = ""
	player := LoadGameRPC(urlBase, session)
	fmt.Printf("LoadGameRPC -> Player: %+v\n\n", player.Data)
	battleState := player.Data.BattleState //Get battle state.
	if battleState.Enemies != nil { //Make sure Enemies property exists.
		enemies := battleState.Enemies
		fmt.Printf("LoadGameRPC -> Enemies: %+v\n", enemies)
		if len(enemies) > 0 { //Make sure it isn't empty.
			for id, e := range enemies {
				if e.Type != "" { //Make sure ID property exists.
					targetID = id
					break
				}
			}
		}
	}
	fmt.Println("-------------------------------------------------------------\n")

	//If there is a targetID then do the RPC: Attack target.
	if targetID != "" {
		attackType := "punch"
		player := AttackTargetRPC(urlBase, session, targetID, attackType)
		fmt.Printf("AttackTargetRPC -> Player: %+v\n", player)
	} else {
		fmt.Println("AttackTargetRPC -> Failed to do an attack, didn't find targetID.")
	}
	fmt.Println("-------------------------------------------------------------\n")

	//RPC: Player info to get player health, status effects, number of enemy types killed.
	response := PlayerInfoRPC(urlBase, session)
	fmt.Printf("PlayerInfoRPC -> Response: %+v\n", response)
	fmt.Println("-------------------------------------------------------------\n")
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
		log.Fatalf("Non 200 code: %d; Response: %+v; Endpoint: %s", res.StatusCode, string(resBody), urlAPI)
	}
	fmt.Printf("AuthAPI -> Response: %s\n", resBody)

	var session Session
	err = json.Unmarshal(resBody, &session)
	if err != nil {
		log.Fatalf("Response: %s; Error: %v", resBody, err)
	}
	
	return session
}

//Makes a call to the RPC to load game.
func LoadGameRPC(urlBase string, session Session) PlayerData {
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
		log.Fatalf("Non 200 code: %d; Response: %+v; Endpoint: %s", res.StatusCode, string(resBody), urlAPI)
	}
	//fmt.Printf("resBody: %s\n", resBody)

	var payload Payload
	err = json.Unmarshal(resBody, &payload)
	if err != nil {
		log.Fatalf("Response: %s; Error: %v", resBody, err)
	}
	
	var player PlayerData
	err = json.Unmarshal([]byte(payload.Data), &player)
	if err != nil {
		log.Fatalf("Response: %+v; Error: %v", payload, err)
	}
	//fmt.Printf("resBody: %+v\n", player)
	
	return player
}

//Makes a call to the RPC to load game.
func AttackTargetRPC(urlBase string, session Session, targetID string, attackType string) PlayerData {
	ctx := context.Background()

	//API call info
	endpoint := "/v2/rpc/attack_target"
	urlAPI := fmt.Sprintf("%s%s", urlBase, endpoint)

	//Request setup.
	reqPacket := map[string]string{
		"target_id": targetID,
		"attack": attackType,
	}
	reqJSON, err := json.Marshal(reqPacket)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	reqPayload := []byte(fmt.Sprintf("%q", string(reqJSON))) //Payload needs to be a string.
	req, err := http.NewRequestWithContext(ctx, "POST", urlAPI, bytes.NewBuffer(reqPayload))
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
		log.Fatalf("Non 200 code: %d;Response: %+v; Endpoint: %s", res.StatusCode, string(resBody), urlAPI)
	}
	//fmt.Printf("resBody: %s\n", resBody)

	var resPayload Payload
	err = json.Unmarshal(resBody, &resPayload)
	if err != nil {
		log.Fatalf("Response: %s; Error: %v", resBody, err)
	}
	
	var player PlayerData
	err = json.Unmarshal([]byte(resPayload.Data), &player)
	if err != nil {
		log.Fatalf("Response: %+v; Error: %v", resPayload, err)
	}
	//fmt.Printf("resBody: %+v\n", player)
	
	return player
}

//Makes a call to the RPC to load game.
func PlayerInfoRPC(urlBase string, session Session) interface{} {
	ctx := context.Background()

	//API call info
	endpoint := "/v2/rpc/player_info"
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
		log.Fatalf("Non 200 code: %d; Response: %+v; Endpoint: %s", res.StatusCode, string(resBody), urlAPI)
	}
	//fmt.Printf("resBody: %s\n", resBody)

	var payload Payload
	err = json.Unmarshal(resBody, &payload)
	if err != nil {
		log.Fatalf("Response: %s; Error: %v", resBody, err)
	}
	
	var player interface{}
	err = json.Unmarshal([]byte(payload.Data), &player)
	if err != nil {
		log.Fatalf("Response: %+v; Error: %v", payload, err)
	}
	//fmt.Printf("resBody: %+v\n", player)
	
	return player
}