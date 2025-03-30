# mcps-go
![Go](https://img.shields.io/badge/Go-1.22-%2300ADD8?logo=go)
![Coverage](https://img.shields.io/badge/Coverage-42.3%25-yellow)
![License](https://img.shields.io/badge/license-MIT-blue)

# Usage
- Go 1.22.2

## Build
```bash
go build main.go
```

## Build with Docker
```bash
docker build -t mcps-go .
```

```bash

docker run -i --rm -e BRAVE_API_KEY=$BRAVE_API_KEY --name my-mcps-go -t mcps-go brave_search
```

## Cline settings
Add the following into `cline_mcp_setting.json`
```json
{
  "mcpServers": {
    "arithmetic_calculator": {
      "command": "/home/nov/mcps-go-for-claude/mcps-go",
      "args": [
        "arith_calc"
      ],
      "disabled": false,
      "autoApprove": []
    },
    "datetime_calculator": {
      "command": "/home/nov/mcps-go-for-claude/mcps-go",
      "args": [
        "datetime_calc"
      ],
      "disabled": false,
      "autoApprove": []
    },
    "http_request": {
      "command": "/home/nov/mcps-go-for-claude/mcps-go",
      "args": [
        "http_request"
      ],
      "disabled": false,
      "autoApprove": []
    }
  }
}

```

# Contributing
Contributions to the project are welcome. Please fork the repository and submit a pull request with your changes.

# License
MIT License
