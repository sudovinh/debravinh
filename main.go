package main

import (
	"context"
	"embed"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:embed web/views web/assets web/robots.txt
var webFS embed.FS

var redirectMap = map[string]string{
	"/levi": "https://digital.fidelity.com/prgw/digital/familygifting/mlgLandingPage?uuid=0ec1277a30464ccbbc25f69b04ca429a",
	// Add more redirects as needed
}

// contentSecurityPolicy allows same-origin content plus Google Fonts; the site has no JavaScript.
const contentSecurityPolicy = "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com; img-src 'self'; script-src 'none'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'"

// newServer builds the Echo instance with all middleware and routes registered.
func newServer() *echo.Echo {
	e := echo.New()
	// Only reachable through the App Platform proxy, which sets X-Forwarded-Proto.
	// Echo >= 4.16 ignores that header unless the peer is a private IP, which the
	// proxy's isn't, so the Secure middleware would silently drop HSTS.
	e.SchemeExtractor = echo.LegacySchemeExtractor()
	e.HideBanner = true

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.BodyLimit("1M"))
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		// JPEGs are already compressed; gzipping them wastes CPU for nothing.
		Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Request().URL.Path, "/assets/imgs/")
		},
	}))
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if strings.HasPrefix(c.Request().URL.Path, "/assets/") {
				c.Response().Header().Set("Cache-Control", "public, max-age=86400")
			}
			return next(c)
		}
	})
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
	// Own http.Server: e.Start would install e itself as the handler,
	// bypassing headAsGet.
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           headAsGet(e),
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			e.Logger.Fatal(err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		e.Logger.Fatal(err)
	}
}

// headAsGet answers HEAD like GET. Echo only routes registered methods, so
// uptime checks and link previews got 404/405. Echo routes a GET copy of the
// request; w still belongs to the HEAD request, so net/http drops the body.
func headAsGet(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			r = r.Clone(r.Context())
			r.Method = http.MethodGet
		}
		h.ServeHTTP(w, r)
	})
}
