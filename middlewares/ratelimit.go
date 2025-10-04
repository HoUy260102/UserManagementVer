package middlewares

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/khaaleoo/gin-rate-limiter/core"
)

func NewRateLimiterMiddleware() gin.HandlerFunc {
	// Cấu hình rate limiter
	rateLimiterOption := core.RateLimiterOption{
		Limit: 2,               // số request tối đa mỗi window
		Burst: 5,               // số request được phép gửi ngay lập tức
		Len:   1 * time.Second, // window 1 giây
	}

	// Tạo middleware từ thư viện
	rateLimiterMiddleware := core.RequireRateLimiter(core.RateLimiter{
		RateLimiterType: core.IPRateLimiter, // áp dụng theo IP
		Key:             "iplimiter_maximum_requests_for_ip_test",
		Option:          rateLimiterOption,
	})

	return func(c *gin.Context) {
		// Gọi middleware của thư viện
		rateLimiterMiddleware(c)

		// Nếu request bị chặn, thư viện thường gọi c.Abort(), nên không cần c.Next()
		if c.IsAborted() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"status":  http.StatusTooManyRequests,
				"message": "Quá nhiều request",
			})
		}

		// Nếu được phép, tiếp tục tới handler
		c.Next()
	}
}
