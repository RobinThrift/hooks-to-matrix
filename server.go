package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/russross/blackfriday"
	"io/ioutil"
	"net/http"
	"text/template"
)

type GitHubRepoConfig struct {
	Name       string
	Homeserver string
	Secret     string
	Rooms      []string
	Username   string
	Password   string
	Format     string
	Template   *template.Template
}

type GitHubHookResponseRepo struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
	Url      string `json:"html_url"`
}

type GitHubHookResponseSender struct {
	Id        int    `json:"id"`
	Login     string `json:"login"`
	AvaratUrl string `json:"avatar_url"`
	Url       string `json:"html_url"`
	Type      string `Json:"type"`
}

type GitHubHookResponse struct {
	Sender     GitHubHookResponseSender `json:"sender"`
	Repository GitHubHookResponseRepo   `json:"repository"`
	Ref        string                   `json:"ref"`
}

func (ghc *GitHubRepoConfig) FillStruct(m map[string]interface{}) error {
	for k, v := range m {
		if k == "rooms" {
			continue
		}
		err := SetField(ghc, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewGitHubRepoConfig(name string, props map[string]interface{}) (*GitHubRepoConfig, error) {
	ghc := &GitHubRepoConfig{
		Name:  name,
		Rooms: make([]string, 0),
	}

	if err := ghc.FillStruct(props); err != nil {
		return nil, err
	}

	rooms, ok := props["rooms"]

	if !ok {
		return nil, errors.New("Invalid rooms declaration")
	}

	for _, r := range rooms.([]interface{}) {
		ghc.Rooms = append(ghc.Rooms, r.(string))
	}

	tmpl, err := template.New(ghc.Name).Parse(ghc.Format)

	if err != nil {
		return nil, err
	}

	ghc.Template = tmpl

	return ghc, nil
}

func handleGithub(ghc *GitHubRepoConfig) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Webhook received for %s\n", ghc.Name)

		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var hookData GitHubHookResponse
		json.Unmarshal(body, &hookData)

		msgBuffer := bytes.NewBuffer(make([]byte, 0))
		ghc.Template.Execute(msgBuffer, hookData)
		msg := blackfriday.MarkdownCommon(msgBuffer.Bytes())

		err = SendMatrixMessage(&MatrixConfig{
			Homeserver:       ghc.Homeserver,
			Rooms:            ghc.Rooms,
			Username:         ghc.Username,
			Password:         ghc.Password,
			Message:          msgBuffer.String(),
			FormattedMessage: string(msg[:]),
		})

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write([]byte(fmt.Sprintf("Message sent to %s", ghc.Name)))
	}
}

func StartServer(port string, repoConfigs []*GitHubRepoConfig) {
	for _, ghc := range repoConfigs {
		http.HandleFunc(fmt.Sprintf("/github/%s", ghc.Name), handleGithub(ghc))
	}

	http.ListenAndServe(port, nil)
}
