package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Workspaces struct {
	Workspaces []Workspace
}

type Workspace struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Visibility string `json:"visibility"`
}

func main() {
	workspaceId := getWorkspaceId()
	fmt.Println("Workspace Id: \t", workspaceId)
}

func getWorkspaceId() string {
	workspaces := getWorkspaces()

	for _, workspace := range workspaces {
		if "Pay Client" == workspace.Name {
			return workspace.Id
		}
	}

	panic(errors.New("empty name"))
}

func getWorkspaces() []Workspace {
	url := "https://api.getpostman.com/workspaces"
	data := getResponse(url, "GET", nil)
	workspacesApi := &Workspaces{}
	err := json.Unmarshal(data, workspacesApi)

	if err != nil {
		panic(err)
	}

	return workspacesApi.Workspaces
}

func getResponse(url string, method string, body io.Reader) []byte {
	apiKey := os.Getenv("API_KEY")
	headerKey := "X-API-Key"

	client := &http.Client{}
	req, _ := http.NewRequest(method, url, body)
	req.Header.Set(headerKey, apiKey)
	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	response, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	return response
}

func dumpJsonBody(body []byte) {

	var prettyJSON bytes.Buffer
	errJson := json.Indent(&prettyJSON, body, "", "\t")
	if errJson != nil {
		log.Println("JSON parse error: ", errJson)
		return
	}

	log.Println("dump json:", string(prettyJSON.Bytes()))
}
