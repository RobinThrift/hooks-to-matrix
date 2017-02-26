package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
)

type MatrixConfig struct {
	Homeserver       string
	Rooms            []string
	Username         string
	Password         string
	Message          string
	FormattedMessage string
}

type auth struct {
	Type     string `json:"type"`
	Username string `json:"user"`
	Password string `json:"password"`
}

type session struct {
	Token       string `json:"access_token"`
	RefeshToken string `json:"refresh_token"`
	Homeserver  string `json:"home_server"`
	UserId      string `json:"user_id"`
}

type msg struct {
	Type          string `json:"msgtype"`
	Body          string `json:"body"`
	FormattedBody string `json:"formatted_body"`
	Format        string `json:"format"`
}

const matrixClientUrl = "/_matrix/client/r0"

var currentSession session

func SendMatrixMessage(config *MatrixConfig) error {
	if len(currentSession.Token) == 0 {
		if err := login(config); err != nil {
			return err
		}
	}

	for _, room := range config.Rooms {
		fmt.Printf("sending a message to room %s\n", room)
		reqUrl := config.Homeserver + matrixClientUrl + "/rooms/" + room + "/send/m.room.message/" + strconv.Itoa(rand.Int()) + "?access_token=" + currentSession.Token
		msgBody, _ := json.Marshal(msg{"m.text", config.Message, config.FormattedMessage, "org.matrix.custom.html"})
		req, _ := http.NewRequest(http.MethodPut, reqUrl, bytes.NewBuffer(msgBody))

		req.Header.Add("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)

		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			return errors.New(fmt.Sprintf("Error while posting to room %s. [Code %d]", room, resp.StatusCode))
		}

		fmt.Printf("successfully posted a message to %s\n", room)
	}

	return nil
}

func login(config *MatrixConfig) error {
	loginData, err := json.Marshal(auth{"m.login.password", config.Username, config.Password})

	if err != nil {
		return err
	}

	resp, err := http.Post(config.Homeserver+matrixClientUrl+"/login", "application/json", bytes.NewBuffer(loginData))

	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusBadRequest {
		return errors.New("Invalid login data")
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return errors.New("Invalid login credentials")
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	json.Unmarshal(body, &currentSession)

	return nil
}
