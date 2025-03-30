# mcps-go
![Go](https://img.shields.io/badge/Go-1.23-%2300ADD8?logo=go)
![Coverage](https://img.shields.io/badge/Coverage-0.0%25-red)
![License](https://img.shields.io/badge/license-MIT-blue)

# Introduction
[Introduction in japanese](https://www.endorphinbath.com/go-cline-mcp-server-date-calculation-web-search)

# Execution

## Requirements
Nothing. A binary file only.

# Development

## Requirements
- Go 1.23.5 or later
- [mcp-go](https://github.com/mark3labs/mcp-go)

## Build
```bash
go build main.go
```

## Build with Docker
```bash
docker build -t mcps-go .
```
Check MCP server
```bash
docker run -i --rm -e BRAVE_API_KEY=$BRAVE_API_KEY mcps-go brave_search
```

## Cline settings
Add the following into `cline_mcp_setting.json`
```json
{
  "mcpServers": {
    "arithmetic_calculator": {
      "command": "/home/nov/mcps-go/mcps-go",
      "args": [
        "arith_calc"
      ],
      "disabled": false,
      "autoApprove": []
    },
    "datetime_calculator": {
      "command": "/home/nov/mcps-go/mcps-go",
      "args": [
        "datetime_calc"
      ],
      "disabled": false,
      "autoApprove": []
    },
    "http_request": {
      "command": "/home/nov/mcps-go/mcps-go",
      "args": [
        "http_request"
      ],
      "disabled": false,
      "autoApprove": []
    },
    "brave_web_search": {
      "command": "/home/nov/mcps-go/mcps-go",
      "args": [
        "brave_web_search"
      ],
      "env": {
        "BRAVE_API_KEY": "YOUR_BRAVE_API_KEY"
      },
      "disabled": false,
      "autoApprove": []
    }
  }
}


```

# Contributing
Welcome.

# License
MIT License
