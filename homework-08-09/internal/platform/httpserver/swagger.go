package httpserver

import (
	"fmt"
	"html"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterSwaggerRoutes(r *gin.Engine, title string, specification []byte) {
	r.GET("/swagger.yaml", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", specification)
	})
	r.GET("/swagger", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerUI(title)))
	})
}

func swaggerUI(title string) string {
	return fmt.Sprintf(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>%s Swagger</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function() {
      window.ui = SwaggerUIBundle({url: "/swagger.yaml", dom_id: "#swagger-ui"});
    };
  </script>
</body>
</html>`, html.EscapeString(title))
}
