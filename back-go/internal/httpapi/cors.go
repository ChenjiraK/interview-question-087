package httpapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const allowedOrigin = "https://interview-question-087.onrender.com"

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Add("Vary", "Origin")
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}
		if origin != allowedOrigin {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Header("Access-Control-Allow-Origin", allowedOrigin)
		c.Header("Access-Control-Expose-Headers", "Location")
		if c.Request.Method == http.MethodOptions && c.GetHeader("Access-Control-Request-Method") != "" {
			c.Writer.Header().Add("Vary", "Access-Control-Request-Method")
			c.Writer.Header().Add("Vary", "Access-Control-Request-Headers")
			method := c.GetHeader("Access-Control-Request-Method")
			if method != http.MethodGet && method != http.MethodPost {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			for _, header := range strings.Split(c.GetHeader("Access-Control-Request-Headers"), ",") {
				switch strings.ToLower(strings.TrimSpace(header)) {
				case "", "content-type", "authorization":
				default:
					c.AbortWithStatus(http.StatusForbidden)
					return
				}
			}
			c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
