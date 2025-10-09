package middlewares

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type Client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	mu      sync.Mutex
	clients = make(map[string]*Client)
)

func getClientIp(c *gin.Context) string {
	ip := c.ClientIP()
	if ip == "" {
		ip = c.Request.RemoteAddr
	}
	return ip
}

func getRateLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()
	client, exists := clients[ip]
	if !exists {
		limiter := rate.NewLimiter(5, 10)
		clients[ip] = &Client{limiter, time.Now()}
		client = clients[ip]
	}
	client.lastSeen = time.Now()
	return client.limiter
}

func NewRateLimiterMiddleware() gin.HandlerFunc {
	// Cấu hình rate limiter
	return func(c *gin.Context) {
		ip := getClientIp(c)
		limiter := getRateLimiter(ip)
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"status":  http.StatusTooManyRequests,
				"message": "Quá nhiều request, hãy thử lại",
			})
			return
		}
		c.Next()
	}
}
