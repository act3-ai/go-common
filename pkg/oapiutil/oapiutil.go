// Package oapiutil implements helper functions for utilizing OpenAPI specifications.
package oapiutil

import (
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"gitlab.com/act3-ai/asce/go-common/pkg/httputil"
)

// SwaggerSpecHandler creates an [http.Handler] to serve the Swagger UI for a raw OpenAPI specification.
func SwaggerSpecHandler(loadSpec func() ([]byte, error)) http.Handler {
	return httputil.RootHandler(func(w http.ResponseWriter, r *http.Request) error {
		spec, err := loadSpec()
		if err != nil {
			return httputil.NewHTTPError(err, 500, "Fetching OpenAPI spec")
		}

		_, err = w.Write([]byte(heredoc.Docf(`
			<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="utf-8" />
				<meta name="viewport" content="width=device-width, initial-scale=1" />
				<meta name="description" content="SwaggerUI" />
				<title>SwaggerUI</title>
				<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css" />
			</head>
			<body>
				<div id="swagger-ui"></div>
				<script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js" crossorigin></script>
				<script crossorigin>
				window.onload = () => {
					window.ui = SwaggerUIBundle({
						spec: %s,
						dom_id: '#swagger-ui',
						docExpansion: false,
						defaultModelRendering: 'model',
					});
				};
				</script>
			</body>
			</html>`, string(spec))))
		if err != nil {
			return httputil.NewHTTPError(err, 500, "Writing HTML")
		}

		return nil
	})
}

// SwaggerURLHandler creates an [http.Handler] to serve the Swagger UI for a OpenAPI specification referenced by URL.
func SwaggerURLHandler(specPath string) http.Handler {
	return httputil.RootHandler(func(w http.ResponseWriter, r *http.Request) error {
		specURL := r.URL.Host + specPath
		_, err := w.Write([]byte(heredoc.Docf(`
			<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="utf-8" />
				<meta name="viewport" content="width=device-width, initial-scale=1" />
				<meta name="description" content="SwaggerUI" />
				<title>SwaggerUI</title>
				<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css" />
			</head>
			<body>
				<div id="swagger-ui"></div>
				<script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js" crossorigin></script>
				<script crossorigin>
				window.onload = () => {
					window.ui = SwaggerUIBundle({
						url: '%s',
						dom_id: '#swagger-ui',
						docExpansion: false,
						defaultModelRendering: 'model',
					});
				};
				</script>
			</body>
			</html>`, specURL)))
		if err != nil {
			return httputil.NewHTTPError(err, 500, "Writing HTML")
		}

		return nil
	})
}

// SpecHandler creates an [http.Handler] to serve an OpenAPI specification.
func SpecHandler(loadSpec func() ([]byte, error)) http.Handler {
	return httputil.RootHandler(func(w http.ResponseWriter, _ *http.Request) error {
		spec, err := loadSpec()
		if err != nil {
			return httputil.NewHTTPError(err, 500, "Fetching OpenAPI spec")
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(spec)
		if err != nil {
			return httputil.NewHTTPError(err, 500, "Writing OpenAPI spec")
		}

		return nil
	})
}
