package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// nginxNotFoundPage is nginx's built-in 404 body with server_tokens off.
const nginxNotFoundPage = "<html>\r\n" +
	"<head><title>404 Not Found</title></head>\r\n" +
	"<body>\r\n" +
	"<center><h1>404 Not Found</h1></center>\r\n" +
	"<hr><center>nginx</center>\r\n" +
	"</body>\r\n" +
	"</html>\r\n"

// NotFound answers like a stock nginx so a probe that never learned the secret
// path cannot tell this listener from any other web server.
func NotFound(c *gin.Context) {
	h := c.Writer.Header()
	for name := range h {
		h.Del(name)
	}
	h.Set("Server", "nginx")
	h.Set("Content-Type", "text/html")
	h.Set("Content-Length", strconv.Itoa(len(nginxNotFoundPage)))
	c.Abort()
	c.Writer.WriteHeader(http.StatusNotFound)
	// Flush now: an unwritten 404 lets gin append its own "404 page not found".
	c.Writer.WriteHeaderNow()
	if c.Request.Method != http.MethodHead {
		_, _ = c.Writer.WriteString(nginxNotFoundPage)
	}
}
