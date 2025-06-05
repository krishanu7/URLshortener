# URL Shortener

A scalable and feature-rich URL shortener built in Go. This project follows a phased implementation approach—from basic setup to advanced system design and deployment.

---

## 📌 Table of Contents

1. [Project Structure](#project-structure)  
2. [Phases](#phases)  
   - [Phase 1: Setup and Basic API](#phase-1-setup-and-basic-api)  
   - [Phase 2: Complete API Endpoints](#phase-2-complete-api-endpoints)  
   - [Phase 3: Add Caching](#phase-3-add-caching)  
   - [Phase 4: Add Frontend and Redirects](#phase-4-add-frontend-and-redirects)  
   - [Phase 5: Add Advanced Features](#phase-5-add-advanced-features)  
3. [Testing & Optimization](#step-6-test-and-optimize)  
4. [Deployment & Scaling](#step-7-deploy-and-scale)  
5. [Advanced System Design](#step-8-advanced-system-design-learning)  
6. [Iterate and Expand](#step-9-iterate-and-expand)  
7. [Resources](#resources)  

---

## Project Structure

```
urlshortener/
├── cmd/
│   └── api/
│       └── main.go              # Entry point for the API server
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration (e.g., environment variables, database settings)
│   ├── handlers/
│   │   └── url.go               # HTTP handlers for URL endpoints
│   ├── models/
│   │   └── url.go               # URL data model and structs
│   ├── repository/
│   │   └── url_repository.go    # Database operations (CRUD)
│   └── util/
│       └── shortcode.go         # Short code generation logic
├── docker-compose.yml           # Docker configuration for PostgreSQL and API
├── Dockerfile                   # Dockerfile for the Go application
├── .env                         # Environment variables (e.g., DB connection string)
├── go.mod                       # Go module file
├── go.sum                       # Go module dependencies
└── README.md                    # Project documentation
```

---

## Phases

### Phase 1: Setup and Basic API

- **Initialize Go Project**  
  ```bash
  go mod init urlshortener
  ```

- **Install Dependencies**  
  ```bash
  go get github.com/gin-gonic/gin
  go get github.com/jmoiron/sqlx
  go get github.com/go-playground/validator/v10
  ```

- **Create HTTP Server**  
  Implement a basic Gin server that responds to `POST /shorten`.

- **Set Up SQLite**  
  Define a simple schema:
  ```sql
  CREATE TABLE urls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    original_url TEXT NOT NULL,
    short_code TEXT NOT NULL UNIQUE,
    access_count INTEGER DEFAULT 0
  );
  ```

- **Short Code Generation**  
  Use a 6-character base62 random generator.

- **URL Validation**  
  Use `net/url` and `validator` to ensure valid URLs.

---

### Phase 2: Complete API Endpoints

- Implement:
  - `POST /shorten` – create short URL
  - `GET /shorten/{shortCode}` – get original URL
  - `GET /stats/{shortCode}` – get access statistics

- Add proper error handling for:
  - Invalid URLs
  - Duplicate/custom short codes
  - Non-existent records

- Write unit tests using Go’s `testing` package.

---

### Phase 3: Add Caching

- **Integrate Redis**
  - Store shortCode → originalURL mappings
  - Reduce DB load on high-traffic endpoints

- **Update Access Count**
  - Track in Redis, periodically flush to DB

---

### Phase 4: Add Frontend and Redirects

- **HTML Frontend**
  - Simple form to input a URL and receive short URL

- **Redirect Endpoint**
  - `GET /{shortCode}` → redirect to original URL

- **Browser Testing**
  - Manually test end-to-end flow

---

### Phase 5: Add Advanced Features

- **Rate Limiting**
  - Use `golang.org/x/time/rate` to limit API calls per IP/user

- **Structured Logging**
  - Add `logrus` for logging API usage and errors

- **Monitoring (Optional)**
  - Integrate Prometheus for performance metrics

- **Security Enhancements**
  - Sanitize and validate inputs to prevent XSS/malicious redirects

---

## Step 6: Test and Optimize

- **Unit Tests**
  - Validate short code generation, URL validation, and API logic

- **Integration Tests**
  - Full flow: shorten → redirect → stats

- **Performance Testing**
  - Use `wrk` or `ab` (ApacheBench) to simulate high load

- **Profiling**
  - Use `pprof` to find bottlenecks

---

## Step 7: Deploy and Scale

- **Deployment Platforms**
  - Heroku, AWS EC2, DigitalOcean, etc.

- **Containerization**
  - Dockerfile for consistent builds:
    ```Dockerfile
    FROM golang:1.21
    WORKDIR /app
    COPY . .
    RUN go build -o urlshortener
    CMD ["./urlshortener"]
    ```

- **Scaling**
  - Add load balancer (e.g., Nginx)
  - Scale horizontally using multiple app instances

- **Database Scaling**
  - Explore replication, read/write separation, or sharding

---

## Step 8: Advanced System Design Learning

- **Distributed Systems**
  - Use services like DynamoDB or CockroachDB for global scale

- **Consistent Hashing**
  - Shard short codes across multiple backend services

- **Bloom Filters**
  - Prevent short code collisions efficiently

- **Advanced Analytics**
  - Track clicks by device, location (store in InfluxDB, visualize in Grafana)

---

## Step 9: Iterate and Expand

- **Authentication**
  - OAuth2 or email/password for managing URLs

- **Custom Short Codes**
  - Let users define their own codes

- **Expiration Logic**
  - Auto-delete expired URLs

- **Analytics Dashboard**
  - Charts and graphs using Chart.js or D3.js

---

## Resources

- 🔍 GitHub: Search for **“Go URL shortener”**
- 📖 Tutorials:  
  - *Building a URL Shortener in Go* – [Medium](https://medium.com)  
  - *Dev.to Go Projects* – [Dev.to](https://dev.to)  
- 📘 System Design Primer:  
  [https://github.com/donnemartin/system-design-primer](https://github.com/donnemartin/system-design-primer)

---

## 📚 Learning Goals by Step

| Step | Goal |
|------|------|
| 1–2  | Build a REST API, learn Go web dev and SQL |
| 3    | Learn caching and Redis |
| 4    | Frontend integration and redirects |
| 5    | Logging, rate limiting, and security |
| 6    | Testing and performance optimization |
| 7    | DevOps, Docker, deployment |
| 8    | System design at scale |
| 9    | Product features and user management |

---

Happy building! 🚀  
Feel free to fork and expand this project.
