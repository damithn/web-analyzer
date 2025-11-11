# Web Analyzer

## Overview
**Web Analyzer** is a simple yet powerful web application built using **Go (Golang)** that analyzes a given web page URL and extracts key details including:
- HTML version
- Page title
- Headings structure (H1–H6)
- Internal, external, and inaccessible links
- Detection of login form presence

This tool demonstrates clean architecture, best practices in Go development, and frontend-backend integration and production-grade micro web tools.

---

## Prerequisites
Before running this project, ensure you have the following installed:

- [Go (version 1.25 or later)](https://go.dev/dl/)
- [Git](https://git-scm.com/downloads)
- Any modern web browser (e.g., Chrome, Firefox, Safari)
- Optional (for development): [VS Code](https://code.visualstudio.com/) or any preferred IDE

---
## Technologies Used

### **Frontend (FE)**
- **HTML5**, **CSS3**, and **Vanilla JavaScript**
- Handles user input and displays analyzed results dynamically
- Provides input validation and error feedback for invalid URLs

- Frontend URL:
http://localhost:8080/ (served directly by Go)

### **Backend (BE)**
- **Go (Golang)** — used for building the server and business logic
- Packages:
  - `net/http` — for HTTP server and requests
  - `encoding/json` — for JSON encoding/decoding
  - `golang.org/x/net/html` — for HTML parsing and analysis
 
- Architecture:
  - handler → HTTP request/response handling
  - service → Business logic (analyzers for HTML version, title, links, forms)
  - main.go → Server setup, routing, graceful shutdown

### **DevOps**
- **GitHub** — for version control and repository hosting
- **Go Modules** — for dependency management
- **DOCKER**: For containerized deployment

## API Documentation (local)
- POST /analyze
- Request:
{
  "url": "https://example.com"
}

Response:
{
  "htmlVersion": "HTML5",
  "pageTitle": "Example Domain",
  "headings": {"h1": 1, "h2": 0},
  "links": {
    "total": 5,
    "internal": 4,
    "external": 1,
    "inaccessible": 0
  },
  "containsLoginForm": true
}

## Setup & Installation
Follow these steps to set up and run the Web Analyzer project:

- Clone the repository :
  git clone https://github.com/<your-username>/web-analyzer.git
  cd web-analyzer
- Initialize Go module :
  go mod init web-analyzer
  go mod tidy
- Run the server :
  go run main.go
- The server will start at: http://localhost:8080
- Access the Web UI : http://localhost:8080

## Usage

-Step 1: Enter any webpage URL (e.g. https://example.com) in the input field.

-Step 2: Click “Analyze” — the backend fetches and processes the page.

-Step 3: Results displayed include:

 - HTML Version: Identifies document type (HTML5, XHTML, etc.)
 - Page Title: Extracts content within <title> tag
 - Headings: Counts h1 -> h6 elements
 - Links: Counts internal, external, and inaccessible links
 - Login Form Detection: Checks for forms with password/username fields


## Testing
The Web Analyzer project includes a set of unit and intergration tests to validate the correctness,performance amd reliability of the system components.
Testing ensures that, all major components(service,cache,and handler layers) behave as expected.

  - How to run tests with coverage:
    - `go test ./... -v -cover`

  - This will: 
    1. Run all test files recursively
    2. Disply verbose output
    3. Show code coverage summary

  - If want to genarate html coverage report:
    - `go test ./... -coverprofile=coverage.out`
    - `go tool cover -html=coverage.out`

  - When test run successfully, you should see output similar to :

    - === RUN   TestAnalyzeLinks_InternalAndExternal
    - --- PASS: TestAnalyzeLinks_InternalAndExternal (0.05s)

    - PASS
    - coverage: 87.3% of statements
    - ok  	web-analyzer/service	0.092s

 ## Challenges
Use Concurrency for Link Analysis
  - Problem : 
    Link checking (AnalyzeLinks()) is done sequentially — every http.Get(link) waits for the previous one to finish.

  - Solution : 
    Use Go routines + a bounded worker pool to check links concurrently but safely (avoid overwhelming network or CPU).

  - Improvements : 
    - Parallelizes link checking → faster response times.
    - Rate-limited concurrency via sem channel prevents overload.
    - Thread-safe updates using sync.Mutex.
    - Non-blocking — the main routine waits for all workers using WaitGroup.

## Future Enhancements
- Archtectural Enhancements
  
  a. Modular Service Architecture
  - Current: Monolithic structure (handler + service + web in one project)
  - Future:
            - Split into modules:
            - analyzer-service (core analysis logic)
            - frontend-service (UI)
            - api-gateway (manages requests, rate limiting, logging)
            - Deploy them as independent Docker containers orchestrated via Kubernetes.

  b. Message Queue Integration (Optional)
  - If analysis becomes CPU-heavy (e.g., parsing large pages):
  - Offload it to background workers using Kafka or RabbitMQ.
  - The frontend triggers an async request → analysis runs in background → results fetched later.

- Performance & Scalability Enhancements

  a. Concurrent Analysis
  - Use Goroutines to parallelize:
     - Link checks
     - Heading parsing
     - Login form detection
     - This can dramatically reduce analysis time for large pages.

  b. Caching Layer
     - Add Redis or In-memory cache to store previously analyzed results.
     - Key: URL → Cached analysis result + timestamp
     - Implement TTL (e.g., 1 hour)

- Security & Reliability
  
  a. Input Validation & Sanitization
     - Strictly validate user URLs to prevent SSRF (Server Side Request Forgery).
     - Limit allowed domains or apply regex validation.

- Middleware layer Enhancements
     - Add Rate limit for the middleware. then we can prevent the abuse of the analyzer service.
     - Add Request Validation for the middleware. move URL validation logic out of the handler.
     - Add Recovery implementation for the middleware . then we can prevent server crashes.
     - Add CORS to the middleware.

- Testing, Quality & Maintainability

  a. Expand Test Coverage
   - Add more unit + integration tests
   - Use mock servers for HTTP requests.
   - Set up GitHub Actions CI to run tests + lint checks automatically.

- Frontend Enhancements

  a. Richer UI/UX
   - Migrate static HTML → React / Vue / Svelte
     
## System Architecture diagram

![System Architecture](https://github.com/damithn/web-analyzer/blob/main/images/Web_Analyser_flow.jpg)
 
Author : https://www.linkedin.com/in/damith-samarakoon-02aba754/
