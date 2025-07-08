# Web Page Analyzer

This web application is to analyzes web pages and provides detailed information about the HTML content. The application includes both a REST API backend (Golang) and a web frontend (HTML, CSS, and JavaScript).

## Features

Following details will be extracted from the HTML page.

- HTML version
- Page title
- Heading level count (h1-h6)
- Link count by type (internal/external)
- Inaccessible link count
- Presence of login form


## Requirements

- Go 1.23 or newer version
- Docker installed on your machine (for docker deployment)

## Technologies 

- Go
- HTML, CSS, JavaScript
- Docker
- Go Standard Library : net/http, golang.org/x/net/html, encoding/json, testing, etc
- Mock Generation - go.uber.org/mock & mockgen (command line tool to generate mocks)
- Http Testing - net/http/httptest
- Git, Make 

## Dependencies

- `golang.org/x/net/html` - HTML parsing
- `go.uber.org/mock` - To generate mock for testing

Execute below command to install dependencies
   ```bash
   go mod tidy
   ```


## Quick Start

1. Start the application:
   ```bash
   go run main.go
   ```

2. Open your browser and go to: `http://localhost:8080`

3. Enter any URL (e.g., `https://example.com`) and click "Analyze" to see results


## API Endpoints

This Golang application serves both the frontend and API endpoints:

- **Web UI**: `http://localhost:8080/`


### REST API Backend

#### POST /api/analyze

Analyzes a web page and returns detailed information.

**Request:**
```json
{
  "url": "https://example.com"
}
```

**Response:**
```json
{
  "html_version": "HTML5",
  "title": "Example Domain",
  "headings": {
    "h1": 1,
    "h2": 0,
    "h3": 0,
    "h4": 0,
    "h5": 0,
    "h6": 0
  },
  "links": {
    "internal": 0,
    "external": 1,
    "inaccessible": 0
  },
  "has_login_form": false
}
```

## Direct Backend API Access

1. **Start the server:**
   ```bash
   go run main.go
   ```

2. **Test with curl:**
   ```bash
   curl -X POST http://localhost:8080/api/analyze \
     -H "Content-Type: application/json" \
     -d '{"url": "https://example.com"}'
   ```

## Docker Deployment

**Running web pages analyzer from docker hub**
   ```bash
   docker run -d -p 8080:8080 --name web-pages-analyzer-app namalsanjaya/web-pages-analyzer:v1.0.0
   ```

## Makefile Commands

This Makefile provides commands for building, testing, generating mocks, docker deployment and running the Web Pages Analyzer service.

#### `make run` - Build and run the web service
#### `make build` - Build the web service binary
#### `make test` - Run all unit tests
#### `make test-coverage` - Run all unit tests with coverage report
#### `make gen-mocks` - Generate mocks for testing
#### `make deps` - Download and install Go dependencies
#### `make tidy` - Clean up go.mod and go.sum files
#### `make docker-build` - Build docker image
#### `make docker-run` - Run docker image
#### `make docker-run-prebuilt` - Run prebuilt docker image from dockerhub

## Future Improvements
- Implement multiple goroutines to parallelly get the HTML version, title & other fields since they are independent.
- Add Redis cache for frequently analyzed web pages
- Process large HTML documents in chunks to reduce memory usage
