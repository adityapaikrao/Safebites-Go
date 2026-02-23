package handler

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.json
var openapiSpec []byte

// SwaggerUI serves the Swagger UI HTML page.
func SwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1"/>
  <title>SafeBites API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"/>
  <style>
    html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin: 0; background: #fafafa; }
    .topbar { display: none; }
  </style>
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
<script>
  window.onload = function() {
    window.ui = SwaggerUIBundle({
      url: "/docs/openapi.json",
      dom_id: "#swagger-ui",
      deepLinking: true,
      presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
      plugins: [SwaggerUIBundle.plugins.DownloadUrl],
      layout: "StandaloneLayout",
      tryItOutEnabled: true,
      supportedSubmitMethods: ["get","post","put","delete","patch","options","head"],
    })
  }
</script>
</body>
</html>`))
}

// OpenAPISpec serves the raw OpenAPI JSON spec.
func OpenAPISpec(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(openapiSpec)
}
