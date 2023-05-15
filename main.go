package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/labstack/echo"
)

func main() {
	e := echo.New()

	// Create the reverse proxy
	proxyURL, _ := url.Parse("https://www.zola.com/wedding/debraandvinh")
	proxy := httputil.NewSingleHostReverseProxy(proxyURL)

	// Define the route to redirect
	redirectRoute := "/wedding"

	e.Any(redirectRoute, func(c echo.Context) error {
		req := c.Request()
		res := c.Response().Writer

		req.URL = proxyURL

		proxy.ServeHTTP(res, req)
		return nil
	})

	// Add the landing page route
	e.GET("/", func(c echo.Context) error {
		// Read the contents of the index.html file
		html, err := ReadFile("index.html")
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
		}

		// Set the content type as HTML
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)

		// Send the HTML response
		return c.HTML(http.StatusOK, html)
	})

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
