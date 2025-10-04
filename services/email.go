package services

import (
	"fmt"
	"net/smtp"
)

type EmailService struct {
	Host string
	Port int
	User string
	Pass string
}

func NewEmailService(host, user, pass string, port int) *EmailService {
	return &EmailService{
		Host: host,
		Port: port,
		User: user,
		Pass: pass,
	}
}

func (e *EmailService) SendNewDeviceAlert(to string, deviceId string, loginTime string) error {
	from := e.User
	password := e.Pass

	// danh sách người nhận
	recipients := []string{to}

	// nội dung email
	subject := "Cảnh báo đăng nhập từ thiết bị mới"
	body := fmt.Sprintf("Tài khoản của bạn vừa đăng nhập từ thiết bị lạ (DeviceID: %s) vào lúc %s.\n\nNếu không phải bạn, vui lòng đổi mật khẩu ngay.", deviceId, loginTime)

	msg := []byte(
		"From: " + from + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"\r\n" +
			body + "\r\n")

	// auth
	auth := smtp.PlainAuth("", from, password, e.Host)

	// gửi
	err := smtp.SendMail(fmt.Sprintf("%s:%d", e.Host, e.Port), auth, from, recipients, msg)
	if err != nil {
		return err
	}

	return nil
}
