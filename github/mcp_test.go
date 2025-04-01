package github

// ヘルパー関数のテスト
// func TestGetStringParam(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		args     map[string]interface{}
// 		key      string
// 		expected string
// 		ok       bool
// 	}{
// 		{
// 			name:     "存在するキー",
// 			args:     map[string]interface{}{"key": "value"},
// 			key:      "key",
// 			expected: "value",
// 			ok:       true,
// 		},
// 		{
// 			name:     "存在しないキー",
// 			args:     map[string]interface{}{"other": "value"},
// 			key:      "key",
// 			expected: "",
// 			ok:       false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, ok := getStringParam(tt.args, tt.key)
// 			if ok != tt.ok {
// 				t.Errorf("getStringParam() ok = %v, want %v", ok, tt.ok)
// 			}
// 			if got != tt.expected {
// 				t.Errorf("getStringParam() = %v, want %v", got, tt.expected)
// 			}
// 		})
// 	}
// }

// func TestGetRequiredStringParam(t *testing.T) {
// 	args := map[string]interface{}{"key": "value"}
// 	got := getRequiredStringParam(args, "key")
// 	if got != "value" {
// 		t.Errorf("getRequiredStringParam() = %v, want %v", got, "value")
// 	}

// 	// パニックのテスト
// 	defer func() {
// 		if r := recover(); r == nil {
// 			t.Errorf("getRequiredStringParam() did not panic for missing key")
// 		}
// 	}()
// 	getRequiredStringParam(args, "missing")
// }

// func TestGetNumberParam(t *testing.T) {
// 	tests := []struct {
// 		name       string
// 		args       map[string]interface{}
// 		key        string
// 		defaultVal int
// 		expected   int
// 	}{
// 		{
// 			name:       "存在するキー",
// 			args:       map[string]interface{}{"key": float64(10)},
// 			key:        "key",
// 			defaultVal: 5,
// 			expected:   10,
// 		},
// 		{
// 			name:       "存在しないキー",
// 			args:       map[string]interface{}{"other": float64(10)},
// 			key:        "key",
// 			defaultVal: 5,
// 			expected:   5,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got := getNumberParam(tt.args, tt.key, tt.defaultVal)
// 			if got != tt.expected {
// 				t.Errorf("getNumberParam() = %v, want %v", got, tt.expected)
// 			}
// 		})
// 	}
// }

// func TestGetBoolParam(t *testing.T) {
// 	tests := []struct {
// 		name       string
// 		args       map[string]interface{}
// 		key        string
// 		defaultVal bool
// 		expected   bool
// 	}{
// 		{
// 			name:       "存在するキー（true）",
// 			args:       map[string]interface{}{"key": true},
// 			key:        "key",
// 			defaultVal: false,
// 			expected:   true,
// 		},
// 		{
// 			name:       "存在するキー（false）",
// 			args:       map[string]interface{}{"key": false},
// 			key:        "key",
// 			defaultVal: true,
// 			expected:   false,
// 		},
// 		{
// 			name:       "存在しないキー",
// 			args:       map[string]interface{}{"other": true},
// 			key:        "key",
// 			defaultVal: true,
// 			expected:   true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got := getBoolParam(tt.args, tt.key, tt.defaultVal)
// 			if got != tt.expected {
// 				t.Errorf("getBoolParam() = %v, want %v", got, tt.expected)
// 			}
// 		})
// 	}
// }

// func TestReturnJSONResult(t *testing.T) {
// 	// このテストは簡略化します
// 	// 実際のmcp.CallToolResultの構造は複雑なため、
// 	// エラーが発生しないことだけを確認します
// 	result := map[string]interface{}{"key": "value"}
// 	_, err := returnJSONResult(result)
// 	if err != nil {
// 		t.Errorf("returnJSONResult() error = %v", err)
// 		return
// 	}

// 	// 成功すれば良しとします
// }

// func TestAddToOptions(t *testing.T) {
// 	tests := []struct {
// 		name          string
// 		options       map[string]interface{}
// 		args          map[string]interface{}
// 		key           string
// 		expectedValue interface{}
// 		expectedExist bool
// 	}{
// 		{
// 			name:          "キーが存在する場合",
// 			options:       map[string]interface{}{},
// 			args:          map[string]interface{}{"key": "value"},
// 			key:           "key",
// 			expectedValue: "value",
// 			expectedExist: true,
// 		},
// 		{
// 			name:          "キーが存在しない場合",
// 			options:       map[string]interface{}{},
// 			args:          map[string]interface{}{"other": "value"},
// 			key:           "key",
// 			expectedValue: nil,
// 			expectedExist: false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			addToOptions(tt.options, tt.args, tt.key)
// 			value, exists := tt.options[tt.key]
// 			if exists != tt.expectedExist {
// 				t.Errorf("addToOptions() key exists = %v, want %v", exists, tt.expectedExist)
// 			}
// 			if exists && value != tt.expectedValue {
// 				t.Errorf("addToOptions() value = %v, want %v", value, tt.expectedValue)
// 			}
// 		})
// 	}
// }

// // GitHubClient のテスト
// func TestNewGitHubClient(t *testing.T) {
// 	client := NewGitHubClient("test-token")
// 	if client == nil {
// 		t.Error("NewGitHubClient() returned nil")
// 		return // clientがnilの場合は早期リターン
// 	}
// 	if client.httpClient == nil {
// 		t.Error("NewGitHubClient() httpClient is nil")
// 	}
// }

// // モックサーバーを作成するヘルパー関数
// func setupMockServer(t *testing.T, statusCode int, response string) *httptest.Server {
// 	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// リクエストヘッダーの検証
// 		if r.Header.Get("Accept") != "application/vnd.github.v3+json" {
// 			t.Errorf("Expected Accept header to be 'application/vnd.github.v3+json', got %s", r.Header.Get("Accept"))
// 		}

// 		// POSTリクエストの場合はボディを検証
// 		if r.Method == "POST" || r.Method == "PATCH" || r.Method == "PUT" {
// 			if r.Header.Get("Content-Type") != "application/json" {
// 				t.Errorf("Expected Content-Type header to be 'application/json', got %s", r.Header.Get("Content-Type"))
// 			}

// 			body, err := io.ReadAll(r.Body)
// 			if err != nil {
// 				t.Fatalf("Failed to read request body: %v", err)
// 			}
// 			defer r.Body.Close()

// 			// ボディが空でないことを確認
// 			if len(body) == 0 && (r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH") {
// 				t.Error("Request body is empty for POST/PUT/PATCH request")
// 			}
// 		}

// 		w.Header().Set("Content-Type", "application/json")
// 		w.WriteHeader(statusCode)
// 		w.Write([]byte(response))
// 	}))
// }

// func TestDoRequest(t *testing.T) {
// 	tests := []struct {
// 		name       string
// 		method     string
// 		statusCode int
// 		response   string
// 		body       io.Reader
// 		wantErr    bool
// 	}{
// 		{
// 			name:       "成功するGETリクエスト",
// 			method:     "GET",
// 			statusCode: 200,
// 			response:   `{"success": true}`,
// 			body:       nil,
// 			wantErr:    false,
// 		},
// 		{
// 			name:       "成功するPOSTリクエスト",
// 			method:     "POST",
// 			statusCode: 201,
// 			response:   `{"created": true}`,
// 			body:       strings.NewReader(`{"key": "value"}`),
// 			wantErr:    false,
// 		},
// 		{
// 			name:       "エラーを返すリクエスト",
// 			method:     "GET",
// 			statusCode: 404,
// 			response:   `{"message": "Not found", "documentation_url": "https://docs.github.com"}`,
// 			body:       nil,
// 			wantErr:    true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			server := setupMockServer(t, tt.statusCode, tt.response)
// 			defer server.Close()

// 			client := NewGitHubClient("test-token")
// 			data, err := client.doRequest(tt.method, server.URL, tt.body)

// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("doRequest() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}

// 			if !tt.wantErr {
// 				var result map[string]interface{}
// 				if err := json.Unmarshal(data, &result); err != nil {
// 					t.Errorf("Failed to unmarshal response: %v", err)
// 				}
// 			}
// 		})
// 	}
// }

// func TestSearchRepositories(t *testing.T) {
// 	mockResponse := `{
// 		"total_count": 1,
// 		"incomplete_results": false,
// 		"items": [
// 			{
// 				"id": 123456,
// 				"name": "test-repo",
// 				"full_name": "test-owner/test-repo",
// 				"html_url": "https://github.com/test-owner/test-repo"
// 			}
// 		]
// 	}`

// 	server := setupMockServer(t, 200, mockResponse)
// 	defer server.Close()

// 	// テスト用のクライアントを作成し、URLを直接指定

// 	client := &GitHubClient{
// 		httpClient: &http.Client{},
// 		token:      "test-token",
// 	}

// 	// SearchRepositoriesメソッドを直接呼ばず、テスト用の関数を作成
// 	searchURL := fmt.Sprintf("%s/search/repositories?q=%s&page=%d&per_page=%d",
// 		server.URL, "test", 1, 10)
// 	data, err := client.doRequest("GET", searchURL, nil)
// 	if err != nil {
// 		t.Errorf("doRequest() error = %v", err)
// 		return
// 	}

// 	var result map[string]interface{}
// 	if err := json.Unmarshal(data, &result); err != nil {
// 		t.Errorf("Failed to unmarshal response: %v", err)
// 		return
// 	}

// 	// レスポンスの検証
// 	if result["total_count"].(float64) != 1 {
// 		t.Errorf("SearchRepositories() total_count = %v, want %v", result["total_count"], 1)
// 	}

// 	items := result["items"].([]interface{})
// 	if len(items) != 1 {
// 		t.Errorf("SearchRepositories() items length = %v, want %v", len(items), 1)
// 	}

// 	item := items[0].(map[string]interface{})
// 	if item["name"].(string) != "test-repo" {
// 		t.Errorf("SearchRepositories() item name = %v, want %v", item["name"], "test-repo")
// 	}
// }

// func TestGetFileContents(t *testing.T) {
// 	mockResponse := `{
// 		"name": "test.txt",
// 		"path": "test.txt",
// 		"sha": "abc123",
// 		"size": 10,
// 		"content": "SGVsbG8gV29ybGQ=",
// 		"encoding": "base64"
// 	}`

// 	server := setupMockServer(t, 200, mockResponse)
// 	defer server.Close()

// 	client := &GitHubClient{
// 		httpClient: &http.Client{},
// 		token:      "test-token",
// 	}

// 	// GetFileContentsメソッドを直接呼ばず、テスト用の関数を作成
// 	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s",
// 		server.URL, "test-owner", "test-repo", "test.txt")
// 	url += fmt.Sprintf("?ref=%s", "main")

// 	data, err := client.doRequest("GET", url, nil)
// 	if err != nil {
// 		t.Errorf("doRequest() error = %v", err)
// 		return
// 	}

// 	var result map[string]interface{}
// 	if err := json.Unmarshal(data, &result); err != nil {
// 		t.Errorf("Failed to unmarshal response: %v", err)
// 		return
// 	}

// 	// ファイルの内容をデコードする
// 	if content, ok := result["content"].(string); ok {
// 		decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(content, "\n", ""))
// 		if err != nil {
// 			t.Errorf("Failed to decode content: %v", err)
// 			return
// 		}
// 		result["decoded_content"] = string(decoded)
// 	}

// 	// レスポンスの検証
// 	if result["name"].(string) != "test.txt" {
// 		t.Errorf("GetFileContents() name = %v, want %v", result["name"], "test.txt")
// 	}

// 	if result["decoded_content"].(string) != "Hello World" {
// 		t.Errorf("GetFileContents() decoded_content = %v, want %v", result["decoded_content"], "Hello World")
// 	}
// }

// func TestCreateIssue(t *testing.T) {
// 	mockResponse := `{
// 		"id": 123456,
// 		"number": 1,
// 		"title": "Test Issue",
// 		"body": "This is a test issue",
// 		"state": "open",
// 		"html_url": "https://github.com/test-owner/test-repo/issues/1"
// 	}`

// 	server := setupMockServer(t, 201, mockResponse)
// 	defer server.Close()

// 	client := NewGitHubClient("test-token")

// 	// モックサーバーのURLを使用するためのテスト用関数
// 	createIssueWithCustomURL := func(client *GitHubClient, owner, repo string, options map[string]interface{}, baseURL string) (map[string]interface{}, error) {
// 		url := fmt.Sprintf("%s/repos/%s/%s/issues", baseURL, owner, repo)

// 		jsonBody, err := json.Marshal(options)
// 		if err != nil {
// 			return nil, err
// 		}

// 		data, err := client.doRequest("POST", url, strings.NewReader(string(jsonBody)))
// 		if err != nil {
// 			return nil, err
// 		}

// 		var result map[string]interface{}
// 		if err := json.Unmarshal(data, &result); err != nil {
// 			return nil, err
// 		}

// 		return result, nil
// 	}

// 	options := map[string]interface{}{
// 		"title": "Test Issue",
// 		"body":  "This is a test issue",
// 	}

// 	// カスタムURLを使用してCreateIssueの機能をテスト
// 	result, err := createIssueWithCustomURL(client, "test-owner", "test-repo", options, server.URL)
// 	if err != nil {
// 		t.Errorf("CreateIssue() error = %v", err)
// 		return
// 	}

// 	// レスポンスの検証
// 	if result["title"].(string) != "Test Issue" {
// 		t.Errorf("CreateIssue() title = %v, want %v", result["title"], "Test Issue")
// 	}

// 	if result["number"].(float64) != 1 {
// 		t.Errorf("CreateIssue() number = %v, want %v", result["number"], 1)
// 	}
// }

// func TestListIssues(t *testing.T) {
// 	mockResponse := `[
// 		{
// 			"id": 123456,
// 			"number": 1,
// 			"title": "Test Issue 1",
// 			"state": "open"
// 		},
// 		{
// 			"id": 123457,
// 			"number": 2,
// 			"title": "Test Issue 2",
// 			"state": "closed"
// 		}
// 	]`

// 	server := setupMockServer(t, 200, mockResponse)
// 	defer server.Close()

// 	client := &GitHubClient{
// 		httpClient: &http.Client{},
// 		token:      "test-token",
// 	}

// 	options := map[string]interface{}{
// 		"state": "all",
// 	}

// 	// ListIssuesメソッドを直接呼ばず、テスト用の関数を作成
// 	url := fmt.Sprintf("%s/repos/%s/%s/issues",
// 		server.URL, "test-owner", "test-repo")

// 	// クエリパラメータを追加
// 	queryParams := []string{}
// 	for k, v := range options {
// 		queryParams = append(queryParams, fmt.Sprintf("%s=%v", k, v))
// 	}
// 	if len(queryParams) > 0 {
// 		url += "?" + strings.Join(queryParams, "&")
// 	}

// 	data, err := client.doRequest("GET", url, nil)
// 	if err != nil {
// 		t.Errorf("doRequest() error = %v", err)
// 		return
// 	}

// 	var result []map[string]interface{}
// 	if err := json.Unmarshal(data, &result); err != nil {
// 		t.Errorf("Failed to unmarshal response: %v", err)
// 		return
// 	}

// 	// レスポンスの検証
// 	if len(result) != 2 {
// 		t.Errorf("ListIssues() result length = %v, want %v", len(result), 2)
// 	}

// 	if result[0]["title"].(string) != "Test Issue 1" {
// 		t.Errorf("ListIssues() first issue title = %v, want %v", result[0]["title"], "Test Issue 1")
// 	}

// 	if result[1]["state"].(string) != "closed" {
// 		t.Errorf("ListIssues() second issue state = %v, want %v", result[1]["state"], "closed")
// 	}
// }

// func TestUpdateIssue(t *testing.T) {
// 	mockResponse := `{
// 		"id": 123456,
// 		"number": 1,
// 		"title": "Updated Issue",
// 		"body": "This issue has been updated",
// 		"state": "closed"
// 	}`

// 	server := setupMockServer(t, 200, mockResponse)
// 	defer server.Close()

// 	client := &GitHubClient{
// 		httpClient: &http.Client{},
// 		token:      "test-token",
// 	}

// 	options := map[string]interface{}{
// 		"title": "Updated Issue",
// 		"body":  "This issue has been updated",
// 		"state": "closed",
// 	}

// 	// UpdateIssueメソッドを直接呼ばず、テスト用の関数を作成
// 	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d",
// 		server.URL, "test-owner", "test-repo", 1)

// 	jsonBody, err := json.Marshal(options)
// 	if err != nil {
// 		t.Errorf("Failed to marshal options: %v", err)
// 		return
// 	}

// 	data, err := client.doRequest("PATCH", url, strings.NewReader(string(jsonBody)))
// 	if err != nil {
// 		t.Errorf("doRequest() error = %v", err)
// 		return
// 	}

// 	var result map[string]interface{}
// 	if err := json.Unmarshal(data, &result); err != nil {
// 		t.Errorf("Failed to unmarshal response: %v", err)
// 		return
// 	}

// 	// レスポンスの検証
// 	if result["title"].(string) != "Updated Issue" {
// 		t.Errorf("UpdateIssue() title = %v, want %v", result["title"], "Updated Issue")
// 	}

// 	if result["state"].(string) != "closed" {
// 		t.Errorf("UpdateIssue() state = %v, want %v", result["state"], "closed")
// 	}
// }

// func TestAddIssueComment(t *testing.T) {
// 	mockResponse := `{
// 		"id": 123456,
// 		"body": "This is a test comment",
// 		"user": {
// 			"login": "test-user"
// 		},
// 		"created_at": "2023-01-01T00:00:00Z"
// 	}`

// 	server := setupMockServer(t, 201, mockResponse)
// 	defer server.Close()

// 	client := &GitHubClient{
// 		httpClient: &http.Client{},
// 		token:      "test-token",
// 	}

// 	// AddIssueCommentメソッドを直接呼ばず、テスト用の関数を作成
// 	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments",
// 		server.URL, "test-owner", "test-repo", 1)

// 	jsonBody, err := json.Marshal(map[string]string{"body": "This is a test comment"})
// 	if err != nil {
// 		t.Errorf("Failed to marshal body: %v", err)
// 		return
// 	}

// 	data, err := client.doRequest("POST", url, strings.NewReader(string(jsonBody)))
// 	if err != nil {
// 		t.Errorf("doRequest() error = %v", err)
// 		return
// 	}

// 	var result map[string]interface{}
// 	if err := json.Unmarshal(data, &result); err != nil {
// 		t.Errorf("Failed to unmarshal response: %v", err)
// 		return
// 	}

// 	// レスポンスの検証
// 	if result["body"].(string) != "This is a test comment" {
// 		t.Errorf("AddIssueComment() body = %v, want %v", result["body"], "This is a test comment")
// 	}
// }

// // GitHubError のテスト
// func TestGitHubError(t *testing.T) {
// 	err := &GitHubError{
// 		Message:          "Not Found",
// 		DocumentationURL: "https://docs.github.com",
// 		StatusCode:       404,
// 	}

// 	expected := "GitHub API Error: Not Found (Status: 404)"
// 	if err.Error() != expected {
// 		t.Errorf("GitHubError.Error() = %v, want %v", err.Error(), expected)
// 	}
// }
