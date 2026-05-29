package smtp

import (
	"context"
	"fmt"

	"lockbox/internal/http/response"

	"gopkg.in/gomail.v2"
)

type MailService interface {
	SendVerifyLink(ctx context.Context, email string, link string) error
	SendOTP(ctx context.Context, email string, otp string) error
}

type mailService struct {
	cfg *SMTPConfig
}

func NewMailService(cfg *SMTPConfig) MailService {
	return &mailService{cfg: cfg}
}

func (m *mailService) SendVerifyLink(ctx context.Context, email string, link string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

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

func (m *mailService) SendOTP(ctx context.Context, email string, otp string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", m.cfg.From)
	msg.SetHeader("To", email)
	msg.SetHeader("Subject", "Your Lockbox OTP code")
	msg.SetBody("text/plain", fmt.Sprintf("Your Lockbox OTP code is %s. This code expires in 1 minute.", otp))

	if err := m.cfg.Dialer.DialAndSend(msg); err != nil {
		return response.InternalServerError("failed to send otp email")
	}

	return nil
}
