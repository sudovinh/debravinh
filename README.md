# debravinh

[![Docker Build and Push Status](https://github.com/sudovinh/debravinh/actions/workflows/docker-build-push.yaml/badge.svg?branch=main)](https://github.com/sudovinh/debravinh/actions/workflows/docker-build-push.yaml)

Latest Image Version: sudovinh/debravinh:7fa47df-20240212030519

Repo for creating new images for debravinh.com

This Go web server application developed using the [Echo framework](https://echo.labstack.com/). The application serves as a redirector and landing social media page for Debra and Vinh (similar to linktree). It provides URL redirection for specific paths and displays a landing page with HTML content.

## Features

- URL Redirection: The server redirects specific paths to predefined URLs using a mapping defined in the `redirectMap` variable.
- Landing Page: The server displays a landing page with HTML content when accessing the root path (`/`).
- Static Files: Static assets such as CSS and images are served from the "web/assets" directory.
- Logging: The server utilizes logging functionality to log requests and errors to a file.
- Error Handling: It handles HTTP 404 errors by redirecting to the home page.

## Usage

1. Clone the repository:

```shell
git clone https://github.com/your-username/your-repository.git
```

2. Install the GO dependencies:

```shell
go mod download
```

3. Build and run the application:

```shell
go run main.go
```

4. Build and create executable (optional)

```shell
go build -o /app/debravinh
./debravinh
open browser and go to 0.0.0.0:8080
```
