package controllers

import (
	"UserManagementVer/collections"
	"UserManagementVer/models"
	"UserManagementVer/utils"
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	MinSize = 1 * 1024 * 1024 // 1MB
	MaxSize = 5 * 1024 * 1024 // 5MB
)

type CreateAccount struct {
	Name      string    `json:"name,omitempty" validate:"required"`
	Email     string    `json:"email,omitempty" validate:"required,email"`
	Password  string    `json:"password,omitempty" validate:"required"`
	Phone     string    `json:"phone,omitempty" validate:"required,phoneVn"`
	Dob       time.Time `json:"dob,omitempty"`
	CreatedBy string    `json:"created_by,omitempty" validate:"required"`
}

type UpdateAccount struct {
	Name      *string    `json:"name,omitempty" validate:"required"`
	Password  *string    `json:"password,omitempty" validate:"required"`
	Phone     *string    `json:"phone,omitempty" validate:"required,phoneVn"`
	Dob       *time.Time `json:"dob,omitempty"`
	UpdatedBy *string    `json:"updated_by,omitempty" validate:"required"`
}

type AccountResponse struct {
	Email  string    `json:"email,omitempty"`
	Name   string    `json:"name,omitempty"`
	Phone  string    `json:"phone,omitempty"`
	Dob    time.Time `json:"dob,omitempty"`
	ImgUrl string    `json:"img_url,omitempty"`
}

type AccountController struct {
	accountCollection *collections.AccountCollection
}

func NewAccountController(accountCollection *collections.AccountCollection) *AccountController {
	return &AccountController{
		accountCollection: accountCollection,
	}
}

func (accountCon *AccountController) CreateAccount(c *gin.Context) {
	var CreateAccount CreateAccount
	if err := c.ShouldBindJSON(&CreateAccount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": err.Error(),
		})
		return
	}
	if err := utils.HandlerValidation(utils.Validator.Struct(CreateAccount)); len(err) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": err,
		})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, checkExisted := accountCon.accountCollection.Find(ctx, bson.M{"email": CreateAccount.Email})

	if !errors.Is(checkExisted, mongo.ErrNoDocuments) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": checkExisted.Error(),
		})
		return
	}

	CreateAccountModel := models.Account{
		Name:      CreateAccount.Name,
		Email:     CreateAccount.Email,
		Password:  CreateAccount.Password,
		Phone:     CreateAccount.Phone,
		Dob:       CreateAccount.Dob,
		CreatedBy: CreateAccount.CreatedBy,
	}

	err := accountCon.accountCollection.Create(ctx, CreateAccountModel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"status":    http.StatusCreated,
		"message":   "Tài khoản đã được tạo thành công",
		"timestamp": time.Now(),
	})
}

func (accountCon *AccountController) UpdateAccount(c *gin.Context) {
	var updateAccountRequest UpdateAccount
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)
	if err := c.ShouldBindJSON(&updateAccountRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": err.Error(),
		})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	oldAccount, err := accountCon.accountCollection.GetAccountById(ctx, objectId)
	if errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  http.StatusNotFound,
			"message": err.Error(),
		})
		return
	}

	updateAccountRequest = updateAccountRequest.handlerUpdateAccountRequest(oldAccount)
	if err := utils.HandlerValidation(utils.Validator.Struct(updateAccountRequest)); len(err) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": err,
		})
		return
	}
	filter := bson.M{
		"_id": objectId,
	}
	update := bson.M{
		"$set": bson.M{
			"phone":      updateAccountRequest.Phone,
			"dob":        updateAccountRequest.Dob,
			"updated_by": updateAccountRequest.UpdatedBy,
			"updated_at": time.Now(),
			"name":       updateAccountRequest.Name,
		},
	}
	err = accountCon.accountCollection.Update(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":    http.StatusOK,
		"timestamp": time.Now(),
		"message":   "Cập nhật tài khoản thành công",
	})
}

func (accountCon *AccountController) FindAccountById(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	accountRe, err := accountCon.accountCollection.GetAccountById(ctx, objectId)
	if errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  http.StatusNotFound,
			"message": err.Error(),
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":    http.StatusOK,
		"timestamp": time.Now(),
		"message":   "Tìm thấy!",
		"data":      accountRe,
	})
}

func (accountCon *AccountController) SoftDelete(c *gin.Context) {
	mp := make(map[string]string)
	if err := c.ShouldBindJSON(&mp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": err.Error(),
		})
		return
	}
	id := mp["id"]
	deleteBy := mp["deleted_by"]
	objectId, _ := primitive.ObjectIDFromHex(id)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	existedAccount, checkExisted := accountCon.accountCollection.GetAccountById(ctx, objectId)
	if errors.Is(checkExisted, mongo.ErrNoDocuments) {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  http.StatusNotFound,
			"message": "Không thấy tài khoản",
		})
		return
	}
	if !existedAccount.DeletedAt.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Tài khoản này đã bị xóa",
		})
		return
	}
	update := bson.M{
		"$set": bson.M{
			"deleted_at": time.Now(),
			"deleted_by": deleteBy,
		},
	}
	err := accountCon.accountCollection.Update(ctx, bson.M{"_id": objectId}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":    http.StatusOK,
		"timestamp": time.Now(),
		"message":   "Đã xóa tài khoản!",
	})
}

func (accountCon *AccountController) RestoreAccount(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	existedAccount, checkExisted := accountCon.accountCollection.GetAccountById(ctx, objectId)
	if errors.Is(checkExisted, mongo.ErrNoDocuments) {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  http.StatusNotFound,
			"message": "Không thấy tài khoản",
		})
		return
	}
	if existedAccount.DeletedAt.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Tài khoản này chưa bị xóa",
		})
		return
	}
	update := bson.M{
		"$set": bson.M{
			"deleted_at": nil,
			"deleted_by": nil,
		},
	}
	err := accountCon.accountCollection.Update(ctx, bson.M{"_id": objectId}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":    http.StatusOK,
		"timestamp": time.Now(),
		"message":   "Đã khôi phục tài khoản!",
	})
}

func (accountCon *AccountController) SearchAccount(c *gin.Context) {
	keyword := c.Query("keyword")
	keyword = regexp.QuoteMeta(keyword)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	filter := bson.M{
		"$and": []bson.M{
			{
				"$or": []bson.M{
					{"name": bson.M{"$regex": keyword, "$options": "i"}},
					{"email": bson.M{"$regex": keyword, "$options": "i"}},
				},
			},
			{
				"$or": []bson.M{
					{"deleted_at": nil},
					{"deleted_at": bson.M{"$exists": false}},
				},
			},
		},
	}
	accounts, err := accountCon.accountCollection.FindAll(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}
	if len(accounts) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  http.StatusNotFound,
			"message": "Không tìm thấy!",
		})
		return
	}
	accountRes := []AccountResponse{}
	for _, account := range accounts {
		accountRes = append(accountRes, AccountResponse{
			Name:   account.Name,
			Phone:  account.Phone,
			ImgUrl: account.ImageUrl,
			Dob:    account.Dob,
			Email:  account.Email,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"status":    http.StatusOK,
		"timestamp": time.Now(),
		"message":   "Tìm thấy!",
		"data":      accountRes,
	})
}
func (ac *AccountController) UploadImage(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)

	// Lấy file từ request
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	if file.Size < MinSize || file.Size > MaxSize {
		c.JSON(400, gin.H{"error": "File phải từ 1MB đến 5MB"})
		return
	}
	//Check valid file
	if err := utils.ChechValidFile(file); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := utils.CheckValidMiMe(file); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Đảm bảo thư mục uploads tồn tại
	if err := ensureUploadDir("uploads"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot create upload folder"})
		return
	}

	// Sinh UUID + tên file
	u := uuid.New().String()
	ext := filepath.Ext(file.Filename)
	baseName := strings.TrimSuffix(file.Filename, ext)
	fileName := fmt.Sprintf("%s_%s%s", u, baseName, ext)
	filePath := filepath.Join("uploads", fileName)

	// Lưu file vào server
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": int(http.StatusInternalServerError),
			"error":  "Cannot save file",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := ac.accountCollection.Update(ctx, bson.M{"_id": objectId}, bson.M{"$set": bson.M{
		"image_url":  filePath,
		"updated_at": time.Now(),
	}}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": int(http.StatusInternalServerError),
			"error":  "Cannot update DB",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Upload success",
		"path":    filePath,
	})
}

func (ac *AccountController) GetAvatar(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	account, err := ac.accountCollection.GetAccountById(ctx, objectId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  http.StatusNotFound,
			"message": "Không tìm thấy",
		})
		return
	}
	fmt.Println(account.ImageUrl)
	filepath := strings.ReplaceAll(account.ImageUrl, "\\", "/")
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"status": 404, "message": "Ảnh không tồn tại"})
		return
	}
	// 2. Trả file ảnh về
	c.File(filepath)
}

func (ac *AccountController) UpdateTimeToLiveHardDelete(c *gin.Context) {
	var ttl map[string]int
	if err := c.ShouldBindJSON(&ttl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": int(http.StatusBadRequest),
			"error":  err.Error(),
		})
		return
	}

	ac.accountCollection.DeleteIndex("deleted_at_1")
	if err := ac.accountCollection.UpdateIndex(ttl["ttl"], "deleted_at"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": int(http.StatusInternalServerError),
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    int(http.StatusOK),
		"timestamp": time.Now(),
		"message":   "Thành công",
	})
}

func (ac *AccountController) DownloadAccountsExcel(c *gin.Context) {
	// Tạo file Excel
	f := excelize.NewFile()
	sheet := "Sheet1"
	_ = f.SetSheetName(f.GetSheetName(0), sheet)

	styleTitle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Size:  14,        // font size lớn hơn header
			Color: "#000000", // chữ đen
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
	})

	styleHeader, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,                   // solid
			Color:   []string{"#ADD8E6"}, // xanh dương nhạt
		},
		Font: &excelize.Font{
			Bold:  true,
			Color: "#000000", // chữ đen
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})

	styleColDob, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})

	// Gán nội dung và style cho A1:D1
	_ = f.MergeCell(sheet, "A1", "D1")
	_ = f.SetCellValue(sheet, "A1", "Thống kê số lượng sinh viên theo ngày")
	_ = f.SetCellStyle(sheet, "A1", "D1", styleTitle)

	// Header
	headers := []string{"Dob", "Name", "Email", "Phone"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		_ = f.SetCellValue(sheet, cell, h)
	}

	// Áp dụng cho header (ví dụ từ A1 đến D1)
	_ = f.SetCellStyle(sheet, "A2", "D2", styleHeader)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	accounts, err := ac.accountCollection.FindAll(ctx, bson.M{})

	sort.Slice(accounts, func(i, j int) bool {
		return accounts[i].Dob.After(accounts[j].Dob)
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": int(http.StatusInternalServerError),
			"error":  err.Error(),
		})
	}

	mergeCol := make(map[string]int)
	widths := make(map[string]float64)
	keys := []string{}

	for _, v := range accounts {
		key := v.Dob.Format("2006-01-02")
		if _, exists := mergeCol[key]; !exists {
			keys = append(keys, key)
		}
		mergeCol[key]++
	}

	//Find max length of each columns
	for _, v := range accounts {
		oldVaA := float64(widths["A"])
		widths["A"] = math.Max(oldVaA, float64(len(v.Dob.Format("2006-01-02"))))
		oldVaB := float64(widths["B"])
		widths["B"] = math.Max(oldVaB, float64(len(v.Name)))
		oldVaC := float64(widths["C"])
		widths["C"] = math.Max(oldVaC, float64(len(v.Email)))
		oldVaD := float64(widths["D"])
		widths["D"] = math.Max(oldVaD, float64(len(v.Phone)))
	}

	//Set style col dob
	_ = f.SetCellStyle(sheet, "A3", fmt.Sprintf("A%d", len(accounts)+2), styleColDob)

	// Ghi dữ liệu
	for row, acc := range accounts {
		_ = f.SetCellValue(sheet, "A"+fmt.Sprintf("%d", row+3), acc.Dob.Format("2006-01-02"))
		_ = f.SetCellValue(sheet, "B"+fmt.Sprintf("%d", row+3), acc.Name)
		_ = f.SetCellValue(sheet, "C"+fmt.Sprintf("%d", row+3), acc.Email)
		_ = f.SetCellValue(sheet, "D"+fmt.Sprintf("%d", row+3), acc.Phone)
	}

	index := 3
	for _, key := range keys {
		v := mergeCol[key]
		if v > 1 {
			_ = f.MergeCell(sheet, fmt.Sprintf("A%d", index), fmt.Sprintf("A%d", index+v-1))
		}
		index += v
	}

	for k, width := range widths {
		_ = f.SetColWidth(sheet, k, k, width*1.3)
	}

	// Statistics
	_ = f.SetCellValue(sheet, "F2", "Dob")
	_ = f.SetCellValue(sheet, "G2", "Số lượng")
	_ = f.SetCellStyle(sheet, "F2", "G2", styleHeader)
	index = 0
	for k, v := range mergeCol {
		_ = f.SetCellValue(sheet, "F"+fmt.Sprintf("%d", index+3), k)
		_ = f.SetCellValue(sheet, "G"+fmt.Sprintf("%d", index+3), v)
		index++
	}

	_ = f.SetColWidth(sheet, "F", "F", widths["A"]*1.2)

	// Trả file về client
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="accounts_%s.xlsx"`, time.Now().Format("2006_01_02_15_04_05")))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Expires", "0")

	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to export excel"})
		return
	}

}

func ensureUploadDir(fileName string) error {
	uploadDir := fileName
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		return os.MkdirAll(uploadDir, os.ModePerm)
	}
	return nil
}

func (updateAccountRequest UpdateAccount) handlerUpdateAccountRequest(oldAccount models.Account) UpdateAccount {
	if updateAccountRequest.Name == nil {
		updateAccountRequest.Name = &oldAccount.Name
	}
	if updateAccountRequest.Dob == nil {
		updateAccountRequest.Dob = &oldAccount.Dob
	}
	if updateAccountRequest.Password == nil {
		updateAccountRequest.Password = &oldAccount.Password
	} else {
		hashPass, _ := utils.HashPassword(*updateAccountRequest.Password)
		updateAccountRequest.Password = &hashPass
	}
	if updateAccountRequest.Phone == nil {
		updateAccountRequest.Phone = &oldAccount.Phone
	}
	return updateAccountRequest
}
