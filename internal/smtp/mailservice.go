package smtp

import (
	"fmt"

	"lockbox/internal/http/response"

	"gopkg.in/gomail.v2"
)

type MailService interface {
	SendVerifyLink(email string, link string) error
	SendOTP(email string, otp string) error
}

type mailService struct {
	cfg *SMTPConfig
}

func NewMailService(cfg *SMTPConfig) MailService {
	return &mailService{cfg: cfg}
}

func (m *mailService) SendVerifyLink(email string, link string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.cfg.From)
	msg.SetHeader("To", email)
	msg.SetHeader("Subject", "Verify your Lockbox email")
	msg.SetBody("text/plain", fmt.Sprintf("Verify your email address by opening this link:\n\n%s\n\nThis link expires in 10 minutes.", link))

	if err := m.cfg.Dialer.DialAndSend(msg); err != nil {
		return response.InternalServerError("failed to send verification email")
	}

	return nil
}

func (m *mailService) SendOTP(email string, otp string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.cfg.From)
	msg.SetHeader("To", email)
	msg.SetHeader("Subject", "Your Lockbox OTP code")

	body := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<body style="font-family: Arial, sans-serif;">
			<h2>Lockbox OTP Verification</h2>
			<p>Your OTP code is:</p>

			<div style="
				font-size: 32px;
				font-weight: bold;
				letter-spacing: 8px;
				margin: 20px 0;
			">
				%s
			</div>

			<p>This code expires in <strong>1 minute</strong>.</p>
			<p>If you did not request this code, you can safely ignore this email.</p>
		</body>
		</html>
	`, otp)

	msg.SetBody("text/html", body)

	if err := m.cfg.Dialer.DialAndSend(msg); err != nil {
		return response.InternalServerError("failed to send otp email")
	}

	return nil
}
