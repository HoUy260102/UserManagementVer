package routers

import (
	"github.com/gin-gonic/gin"
)

type Router interface {
	Register(v *gin.RouterGroup)
}

func RegisterRouters(v *gin.RouterGroup, routers ...Router) {
	for _, router := range routers {
		router.Register(v)
	}
}
