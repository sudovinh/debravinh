package main

import (
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var redirectMap = map[string]string{
	"/wedding": "https://www.zola.com/wedding/debraandvinh",
	// Add more redirects as needed
}

func main() {
	e := echo.New()

	// Set up logging
	e.Use(middleware.Logger())
	logFile, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	e.Logger.SetOutput(logFile)

	// Register static file handler
	e.Static("/assets", "web/assets")

	// Add the landing page route
	e.GET("/", func(c echo.Context) error {
		// Read the contents of the index.html file
		html, err := ReadFile("web/views/index.html")
		if err != nil {
			e.Logger.Errorf("Failed to read index.html: %s", err.Error())
			return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
		}

		// Set the content type as HTML
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)

		// Send the HTML response
		return c.HTML(http.StatusOK, html)
	})

	// Add the redirect routes
	for path, redirectURL := range redirectMap {
		path := path               // Create a new variable to capture the loop value correctly
		redirectURL := redirectURL // Create a new variable to capture the loop value correctly

		e.GET(path, func(c echo.Context) error {
			e.Logger.Infof("Redirecting %s to %s", path, redirectURL)
			return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
		})
	}

	// Redirect 404 errors to "/dv"
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
		}
		if code == http.StatusNotFound {
			c.Redirect(http.StatusTemporaryRedirect, "/")
			return
		}
		e.DefaultHTTPErrorHandler(err, c)
	}

	// Start the server
	e.Logger.Infof("Server started on :8080")
	e.Start(":8080")
}

// ReadFile reads the contents of a file and returns it as a string
func ReadFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
