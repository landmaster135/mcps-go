# go_dev_template
![Go](https://img.shields.io/badge/Go-1.22-%2300ADD8?logo=go)
![Coverage](https://img.shields.io/badge/Coverage-42.3%25-yellow)
![License](https://img.shields.io/badge/license-MIT-blue)

# Cloud Functions

## Entry points
- any_function

## Example
Request
```go
{
  "any": "ANY"
}
```
Example of response
```
{"data": "any_data"}
```

# Features
-

# Usage
- Go 1.22.2

# Deployment
Pack the functions to deploy.
```bash
cd mypkg
go mod init example.com/mypkg
go mod tidy
```

Deploy the function in the directory: `example.com/mypkg`.
```bash
gcloud functions deploy any-function \
  --gen2 \
  --runtime=go122 \
  --region={MY_REGION} \
  --source=. \
  --entry-point={ANY_FUNCTION} \
  --trigger-http \
  --allow-unauthenticated \
  --timeout=180s \
```

Set env-vars.
```bash
gcloud functions deploy any-function \
  --set-env-vars SCRIPT_MANAGER_API_CLIENT_ID=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,SCRIPT_MANAGER_API_CLIENT_SECRET=BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB,SCRIPT_MANAGER_API_ENDPOINT=https://CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC
```

# Development

## Tools
Refer `go.mod` and etc..

## Installing
go version
```bash
go version
go list -m -u all
go mod init example.com/mymodule
go mod tidy
```

## Environment settings
Ses environment variables
```bash
export ANY_ENV='any_value'
```
Sets if deployed on cloud function
```bash
export SCRIPT_MANAGER_API_CLIENT_ID='AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA'
export SCRIPT_MANAGER_API_CLIENT_SECRET='BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB'
export SCRIPT_MANAGER_API_ENDPOINT='https://CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC'
```

## Memory investigation
Add following statements into each function to investigate the states of its memory.

## Test
Install libraries for test at first.
```bash
go get github.com/stretchr/testify/assert
go get go.uber.org/mock/mockgen@latest
```

### Generate test codes and activate `testify/assert`
Then, if testing in VSCode, Eclipse theia and so on, we can auto-generate test codes.
1. In the source file in the editor, press `Ctrl + Shift + P` to open command palette.
2. Select `GO: Install/Update Tools`
3. If you use VSCode, install this extension: `Go` by golang.
4. Add following field and values into `settings.json`
```json
"go.generateTestsFlags": [
		"-template",
		"testify"
	],
```

5. Open command palette again, then, select `GO: Generate Unit Tests for File.`
6. The test file for the current source file is generated!
7. Move test files to `example.com/mymodule/tests` directory.
8. Add import statements.
```go
mypkg "mymodule/mypkg"
// mocks "mymodule/mocks"
```

### Generate mock files and use those mocks
1. Add following statement into each test function in the test file.
```go
  ctrl := gomock.NewController(t)
  defer ctrl.Finish()
```
2. Install mock library: `go get go.uber.org/mock/mockgen@latest`
3. Generate mock file with like this command: `mockgen -source main.go -destination mocks/main.go -package mocks`
4. The mock file is generated!
5. Add test data into that generated template codes

### Test
Test with following command.
```bash
go test -v -coverpkg=./mypkg ./...
```

## Build
```bash
go build main.go
```

# Contributing
Contributions to the project are welcome. Please fork the repository and submit a pull request with your changes.

# License
MIT License
