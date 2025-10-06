package middlewares

import (
	"UserManagementVer/services"
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var (
	unAvailableType = []string{"approved"}
)

func AuthorizeJWT(jwtServce *services.JwtService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		authHeader = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if authHeader == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Token không thấy!",
			})
			c.Abort() // ngăn handler tiếp tục chạy
			return
		}
		token, err := jwtServce.ValidateToken(authHeader)
		tokenClaims, _ := jwtServce.ExtractCustomClaims(token.Raw)
		if token.Valid {
			claims := token.Claims.(jwt.Claims)
			log.Println(claims)
			if slices.Contains(unAvailableType, tokenClaims.Type) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"status":  http.StatusUnauthorized,
					"message": "Không có quyền truy cập",
				})
				c.Abort()
				return
			}
			c.Next()
		} else {
			log.Println(err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  http.StatusUnauthorized,
				"message": err.Error(),
			})
			c.Abort() // ngăn handler tiếp tục chạy
			return
		}
	}
}
