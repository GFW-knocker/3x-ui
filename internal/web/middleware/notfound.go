package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NotFound answers exactly like a stock net/http ServeMux miss, so a probe that
// never learned the secret path sees one of the commonest servers on the net.
func NotFound(c *gin.Context) {
	h := c.Writer.Header()
	for name := range h {
		h.Del(name)
	}
	c.Abort()
	http.NotFound(c.Writer, c.Request)
}
