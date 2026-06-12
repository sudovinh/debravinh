package main

import (
	"context"
	"embed"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:embed web/views web/assets web/robots.txt
var webFS embed.FS

var redirectMap = map[string]string{
	"/wedding": "https://www.zola.com/wedding/debraandvinh",
	// Add more redirects as needed
}

// contentSecurityPolicy allows same-origin content plus Google Fonts; the site has no JavaScript.
const contentSecurityPolicy = "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com; img-src 'self'; script-src 'none'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'"

// newServer builds the Echo instance with all middleware and routes registered.
func newServer() *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.BodyLimit("1M"))
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "0", // deprecated header; CSP supersedes it
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000,
		ContentSecurityPolicy: contentSecurityPolicy,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}))
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
			return next(c)
		}
	})

	indexHTML := mustReadPage("web/views/index.html")
	aboutHTML := mustReadPage("web/views/aboutus.html")

	e.GET("/", func(c echo.Context) error {
		return c.HTMLBlob(http.StatusOK, indexHTML)
	})
	e.GET("/aboutus", func(c echo.Context) error {
		return c.HTMLBlob(http.StatusOK, aboutHTML)
	})

	e.StaticFS("/assets", echo.MustSubFS(webFS, "web/assets"))
	e.FileFS("/robots.txt", "web/robots.txt", webFS)

	for path, redirectURL := range redirectMap {
		e.GET(path, func(c echo.Context) error {
			return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
		})
	}

	// Send unknown paths back to the landing page instead of a bare 404.
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		var he *echo.HTTPError
		if errors.As(err, &he) && he.Code == http.StatusNotFound {
			if err := c.Redirect(http.StatusTemporaryRedirect, "/"); err == nil {
				return
			}
		}
		e.DefaultHTTPErrorHandler(err, c)
	}

	return e
}

// mustReadPage loads an embedded page once at startup so requests never touch the filesystem.
func mustReadPage(path string) []byte {
	page, err := webFS.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read embedded page %s: %v", path, err)
	}
	return page
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	e := newServer()
	e.Server.ReadTimeout = 10 * time.Second
	e.Server.ReadHeaderTimeout = 5 * time.Second
	e.Server.WriteTimeout = 10 * time.Second
	e.Server.IdleTimeout = 60 * time.Second

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := e.Start(":" + port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			e.Logger.Fatal(err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		e.Logger.Fatal(err)
	}
}
