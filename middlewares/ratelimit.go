package middlewares

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

type Client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	//mu              sync.Mutex
	//clients         = make(map[string]*Client)
	rateLimitPrefix = "rate_limit_"
	capacity        = 5 // max token
	refill          = 1 * time.Second
)

func getClientIp(c *gin.Context) string {
	ip := c.ClientIP()
	if ip == "" {
		ip = c.Request.RemoteAddr
	}
	return ip
}

//func getRateLimiter(ip string) *rate.Limiter {
//	mu.Lock()
//	defer mu.Unlock()
//	client, exists := clients[ip]
//	if !exists {
//		limiter := rate.NewLimiter(2, 5)
//		clients[ip] = &Client{limiter, time.Now()}
//		client = clients[ip]
//	}
//	client.lastSeen = time.Now()
//	return client.limiter
//}

type RateLimit struct {
	ctx context.Context
	rdb *redis.Client
}

func NewRateLimitService(rdb *redis.Client) *RateLimit {
	return &RateLimit{
		ctx: context.Background(),
		rdb: rdb,
	}
}

func (rl *RateLimit) rateLimit(c *gin.Context) bool {
	ip := getClientIp(c)
	key := rateLimitPrefix + ip
	res, _ := rl.rdb.HGetAll(rl.ctx, key).Result()
	tokens := float64(capacity)
	last := time.Now().Unix()

	if t, ok := res["tokens"]; ok {
		tokens, _ = strconv.ParseFloat(t, 64)
	}
	if l, ok := res["last"]; ok {
		last, _ = strconv.ParseInt(l, 10, 64)
	}

	// Làm dầy token
	now := time.Now().Unix()
	tokens += float64(now-last) / refill.Seconds()
	if tokens > float64(capacity) {
		tokens = float64(capacity)
	}

	if tokens < 1 {
		return false // quá giới hạn
	}
	tokens -= 1

	rl.rdb.HSet(rl.ctx, key, "tokens", tokens, "last", now)
	return true
}

//func NewRateLimiterMiddleware() gin.HandlerFunc {
//	return func(c *gin.Context) {
//		ip := getClientIp(c)
//		limiter := getRateLimiter(ip)
//		if !limiter.Allow() {
//			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
//				"status":  http.StatusTooManyRequests,
//				"message": "Quá nhiều request, hãy thử lại",
//			})
//			return
//		}
//		c.Next()
//	}
//}

func (rl *RateLimit) NewRateLimiterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		check := rl.rateLimit(c)
		if !check {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"status":  http.StatusTooManyRequests,
				"message": "Quá nhiều request, hãy thử lại",
			})
			return
		}
		c.Next()
	}
}
