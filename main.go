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
	"time"
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

type QueryParams struct {
	Key   string
	Value string
}

type RequestData struct {
	Body        io.Reader
	QueryParams []QueryParams
}

type Collections struct {
	Collections []CollectionFromCollections
}

type CollectionFromCollections struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	Owner     string    `json:"owner"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Uid       string    `json:"uid"`
	Fork      struct {
		Label     string    `json:"label"`
		CreatedAt time.Time `json:"createdAt"`
		From      string    `json:"from"`
	} `json:"fork"`
	IsPublic bool `json:"isPublic"`
}

type CollectionResponseObject struct {
	Collection Collection `json:"collection"`
}

type Collection struct {
	Info struct {
		PostmanId   string `json:"_postman_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Schema      string `json:"schema"`
		Fork        struct {
			Label     string    `json:"label"`
			CreatedAt time.Time `json:"createdAt"`
			From      string    `json:"from"`
		} `json:"fork"`
		UpdatedAt time.Time `json:"updatedAt"`
	}
	Item []CollectionItem
}

type CollectionItem struct {
	Name                    string           `json:"name"`
	Item                    []CollectionItem `json:"item"`
	Id                      string           `json:"id"`
	ProtocolProfileBehavior struct {
		DisableBodyPruning bool `json:"disableBodyPruning"`
	} `json:"protocolProfileBehavior"`
	Request struct {
		Method string        `json:"method"`
		Header []interface{} `json:"header"`
		Url    struct {
			Raw  string   `json:"raw"`
			Host []string `json:"host"`
			Path []string `json:"path"`
		} `json:"url"`
	} `json:"request"`
	Response []interface{} `json:"response"`
}

func main() {
	//Добавить валидацию, что Env задана
	workspaceId := getWorkspaceId()
	collectionId := getCollectionId(workspaceId)
	_ = getUrlsFromPostman(collectionId)
	fmt.Println("Workspace Id: \t", workspaceId)
	fmt.Println("CollectionFromCollections Id: \t", collectionId)
}

func getUrlsFromPostman(collectionId string) int {
	url := "https://api.getpostman.com/collections/" + collectionId
	requestData := RequestData{}
	data := getResponse(url, "GET", requestData)
	collectionResponseObject := &CollectionResponseObject{}
	err := json.Unmarshal(data, collectionResponseObject)
	if err != nil {
		panic(err)
	}

	// TODO переделать на slice который после можно будет сравнивать с выводом получаемым от php artisan route:list
	printRequestUrlRaw(collectionResponseObject.Collection.Item)

	return 0
}

func printRequestUrlRaw(items []CollectionItem) {
	for _, item := range items {
		if item.Item != nil {
			printRequestUrlRaw(item.Item)
		}
		fmt.Println(item.Request.Url.Raw)
	}
}

func getCollectionId(workspaceId string) string {
	collections := getCollections(workspaceId)
	for _, collection := range collections {
		if collection.Fork.Label == "staging" {
			return collection.Uid
		}
	}

	panic(errors.New("collection not found"))
}

func getCollections(workspaceId string) []CollectionFromCollections {
	url := "https://api.getpostman.com/collections"
	queryParams := []QueryParams{
		{Key: "workspace", Value: workspaceId},
	}
	requestData := RequestData{
		QueryParams: queryParams,
	}
	data := getResponse(url, "GET", requestData)
	collections := &Collections{}
	err := json.Unmarshal(data, collections)

	if err != nil {
		panic(err)
	}

	return collections.Collections
}

func getWorkspaceId() string {
	workspaces := getWorkspaces()

	for _, workspace := range workspaces {
		if "Pay Client" == workspace.Name {
			return workspace.Id
		}
	}

	panic(errors.New("workspace not found"))
}

func getWorkspaces() []Workspace {
	url := "https://api.getpostman.com/workspaces"
	requestData := RequestData{}
	data := getResponse(url, "GET", requestData)
	workspacesApi := &Workspaces{}
	err := json.Unmarshal(data, workspacesApi)

	if err != nil {
		panic(err)
	}

	return workspacesApi.Workspaces
}

func getResponse(url string, method string, requestData RequestData) []byte {
	apiKey := os.Getenv("API_KEY")
	headerKey := "X-API-Key"

	client := &http.Client{}
	body := requestData.Body
	queryParams := requestData.QueryParams
	req, _ := http.NewRequest(method, url, body)
	req.Header.Set(headerKey, apiKey)

	q := req.URL.Query()
	if queryParams != nil {
		for _, queryParam := range queryParams {
			q.Add(queryParam.Key, queryParam.Value)
		}
	}
	req.URL.RawQuery = q.Encode()

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
