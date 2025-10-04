package controllers

import (
	"UserManagementVer/collections"
	"UserManagementVer/configs"
	"UserManagementVer/models"
	"UserManagementVer/services"
	"UserManagementVer/utils"
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AuthController struct {
	sessionCollection *collections.SessionCollection
	accountCollection *collections.AccountCollection
	emailService      *services.EmailService
	jwtService        *services.JwtService
}
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func NewAuthController(sessionController *collections.SessionCollection, accountController *collections.AccountCollection, emailService *services.EmailService, jwtService *services.JwtService) *AuthController {
	return &AuthController{sessionCollection: sessionController, accountCollection: accountController, emailService: emailService, jwtService: jwtService}
}

var MaxDevice int = 1

func (auth *AuthController) Login(c *gin.Context) {
	var loginRequest LoginRequest
	deviceId := c.GetHeader("Device-Id")

	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": int(http.StatusBadRequest),
			"error":  err.Error(),
		})
		return
	}

	if err := utils.HandlerValidation(utils.Validator.Struct(loginRequest)); len(err) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": int(http.StatusBadRequest),
			"error":  err,
		})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	account, err := auth.accountCollection.Find(ctx, bson.M{"email": loginRequest.Email})
	if errors.Is(err, mongo.ErrNoDocuments) || !utils.CheckPassword(account.Password, loginRequest.Password) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":   http.StatusBadRequest,
			"messsage": "tài khoản hoặc mật khẩu không chính xác",
		})
		return
	}
	//Lấy danh sách các deviceId cùng đăng nhập với user
	filer := bson.M{
		"user_id":        account.Id,
		"trusted_device": true,
		"device_id": bson.M{
			"$ne": deviceId,
		},
	}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})
	loginAccounts, _ := auth.sessionCollection.Find(ctx, filer, opts)

	countAccount := len(loginAccounts)
	if countAccount >= MaxDevice {
		//Gửi mail
		oldestAccount, _ := auth.accountCollection.GetAccountById(ctx, loginAccounts[0].UserId)
		auth.emailService.SendNewDeviceAlert(oldestAccount.Email, deviceId, time.Now().Format("2006-01-02"))
		approvedToken, _, _ := auth.jwtService.GenerateJwt(account.Email, configs.AppConfig.Jwt.JwtAprrovedTokenExpirationTime)

		_, err := auth.sessionCollection.FindAndUpdate(ctx, models.Session{
			ExpiresAt:     time.Time{},
			IsRevoked:     false,
			TrustedDevice: false,
			CreatedAt:     time.Time{},
			UserId:        account.Id,
			RefreshToken:  "",
			DeviceId:      deviceId,
			ApprovedToken: approvedToken,
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Đã đăng nhập nhiều hơn số thiết bị đã cho phép vui lòng check email",
		})
		return
	}

	accessToken, accessTokenClaims, err := auth.jwtService.GenerateJwt(account.Email, configs.AppConfig.Jwt.JwtAccessTokenExpirationTime)
	if accessToken == "" || err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  int(http.StatusInternalServerError),
			"message": "Không thể sinh được token",
		})
		return
	}

	refreshToken, refreshTokenClaims, err := auth.jwtService.GenerateJwt(account.Email, configs.AppConfig.Jwt.JwtRefreshTokenExpirationTime)
	if refreshToken == "" || err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  int(http.StatusInternalServerError),
			"message": "Không thể sinh được token",
		})
		return
	}

	sessionRes, _ := auth.sessionCollection.FindAndUpdate(ctx, models.Session{
		ExpiresAt:     refreshTokenClaims.RegisteredClaims.ExpiresAt.Time,
		IsRevoked:     false,
		TrustedDevice: true,
		CreatedAt:     refreshTokenClaims.RegisteredClaims.IssuedAt.Time,
		UserId:        account.Id,
		RefreshToken:  refreshToken,
		DeviceId:      deviceId,
		ApprovedToken: "",
	})

	c.JSON(http.StatusOK, bson.M{
		"status":    int(http.StatusOK),
		"message":   "Login account successfully",
		"timestamp": time.Now(),
		"data": bson.M{
			"session_id":               sessionRes.Id,
			"access_token":             accessToken,
			"refresh_token":            refreshToken,
			"access_token_expired_at":  accessTokenClaims.ExpiresAt.Time,
			"refresh_token_expired_at": refreshTokenClaims.ExpiresAt.Time,
		},
	})
}

func (auth *AuthController) ConfirmLogin(c *gin.Context) {
	confirm := c.Query("confirm")
	approvedToken := c.Query("approved_token")

	_, err := auth.jwtService.ValidateToken(approvedToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  http.StatusUnauthorized,
			"message": err.Error(),
		})
		return
	}

	approvedClaims, err := auth.jwtService.ExtractCustomClaims(approvedToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  http.StatusUnauthorized,
			"message": err.Error(),
		})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if confirm == "true" {
		filter := bson.M{
			"approved_token": approvedToken,
		}
		existsSession, checkExists := auth.sessionCollection.FindOne(ctx, filter)
		if errors.Is(checkExists, mongo.ErrNoDocuments) {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Approved token không tồn tại!",
			})
			return
		}
		filter = bson.M{
			"user_id":        existsSession.UserId,
			"trusted_device": true,
			"device_id": bson.M{
				"$ne": existsSession.DeviceId,
			},
		}
		opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})

		sessions, _ := auth.sessionCollection.Find(ctx, filter, opts)

		if len(sessions) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": http.StatusInternalServerError,
			})
			return
		}
		oldestAccount := sessions[0]

		refreshToken, refreshTokenClaims, err := auth.jwtService.GenerateJwt(approvedClaims.Email, configs.AppConfig.Jwt.JwtRefreshTokenExpirationTime)

		if refreshToken == "" || err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": err.Error(),
			})
			return
		}

		accountRes, err := auth.sessionCollection.FindAndUpdate(ctx, models.Session{
			UserId:        existsSession.UserId,
			DeviceId:      existsSession.DeviceId,
			CreatedAt:     refreshTokenClaims.RegisteredClaims.IssuedAt.Time,
			ExpiresAt:     refreshTokenClaims.RegisteredClaims.ExpiresAt.Time,
			IsRevoked:     false,
			TrustedDevice: true,
			RefreshToken:  refreshToken,
			ApprovedToken: "",
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": err.Error(),
			})
			return
		}
		//Xóa session cũ nhất
		filter = bson.M{
			"_id": oldestAccount.Id,
		}
		err = auth.sessionCollection.DeleteSession(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": err.Error(),
			})
			return
		}
		accessToken, accessTokenClaims, errJwt := auth.jwtService.GenerateJwt(approvedClaims.Email, configs.AppConfig.Jwt.JwtAccessTokenExpirationTime)
		if accessToken == "" || errJwt != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  int(http.StatusInternalServerError),
				"message": "Không thể sinh được token",
			})
			return
		}
		c.JSON(http.StatusOK, bson.M{
			"status":    http.StatusOK,
			"message":   "Login account successfully",
			"timestamp": time.Now(),
			"data": bson.M{
				"session_id":               existsSession.Id,
				"access_token":             accessToken,
				"refresh_token":            accountRes.RefreshToken,
				"access_token_expires_at":  accessTokenClaims.ExpiresAt.Time,
				"refresh_token_expires_at": accountRes.ExpiresAt,
			},
		})
	} else {

		//Xóa session bị từ chối
		filter := bson.M{
			"approved_token": approvedToken,
		}
		err = auth.sessionCollection.DeleteSession(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Từ chối đăng nhập",
		})
	}
}
