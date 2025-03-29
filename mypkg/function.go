package mypkg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	functions.HTTP("AnyCloudFunction", AnyCloudFunction)
}

func AnyFunction(token string) (string, error) {

  // TODO: Write your functions...

	ctx := context.Background()

	results := fmt.Sprintf("any result %v", ctx)
  err := fmt.Errorf("any exception")
	return results, err
}

func RecordLog(scriptName string, functionName string, isRecording bool) string {
	SMapiClientID, ok := os.LookupEnv("SCRIPT_MANAGER_API_CLIENT_ID")
	if !ok {
		fmt.Println("SCRIPT_MANAGER_API_CLIENT_ID is not set")
	}
	SMapiClientSecret, ok := os.LookupEnv("SCRIPT_MANAGER_API_CLIENT_SECRET")
	if !ok {
		fmt.Println("SCRIPT_MANAGER_API_CLIENT_SECRET is not set")
	}
	SMapiEndpoint, ok := os.LookupEnv("SCRIPT_MANAGER_API_ENDPOINT")
	if !ok {
		fmt.Println("SCRIPT_MANAGER_API_ENDPOINT is not set")
	}

	if !isRecording {
		return ""
	}

	type ClientInfo struct {
		ID     string `json:"id"`
		Secret string `json:"secret"`
	}
	postData := struct {
		Path       string     `json:"path"`
		Method     string     `json:"method"`
		ClientInfo ClientInfo `json:"client_info"`
	}{
		Path:   "/log",
		Method: "POST",
		ClientInfo: ClientInfo{
			ID:     SMapiClientID,
			Secret: SMapiClientSecret,
		},
	}

	endpoint := fmt.Sprintf("%s?scriptname=%s&function-name=%s", SMapiEndpoint, scriptName, functionName)
	jsonPostData, err := json.Marshal(postData)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	req, err := http.NewRequest(
		"POST",
		endpoint,
		bytes.NewBuffer([]byte(jsonPostData)),
	)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)
	fmt.Println(resp.StatusCode)
	fmt.Println(string(responseBody))

	return string(responseBody)
}

// helloHTTP is an HTTP Cloud Function with a request parameter.
func AnyCloudFunction(w http.ResponseWriter, r *http.Request) {
	const myScriptName = "github-info-getter-with-go"
	const myFunctionName = "github-info-getter-with-go"

	type ReqBody struct {
		Token string `json:"token"`
	}
	var rb ReqBody
  fmt.Println(r.Body)
	if err := json.NewDecoder(r.Body).Decode(&rb); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
  fmt.Println(rb)
  fmt.Println(rb.Token)

	results, err := AnyFunction(rb.Token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println(results)

	resObj := struct {
		Data interface{} `json:"data"`
	}{
		Data: results,
	}
	json.NewEncoder(w).Encode(resObj)

	RecordLog(myScriptName, myFunctionName, true)
}
