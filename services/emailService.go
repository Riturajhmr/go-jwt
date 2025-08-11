package services

import (
	"fmt"
	"net/smtp"
	"os"
	"time"

	"github.com/google/uuid"
)

type EmailService struct {
	fromEmail    string
	fromPassword string
	smtpHost     string
	smtpPort     string
}

func NewEmailService() *EmailService {
	return &EmailService{
		fromEmail:    os.Getenv("EMAIL_FROM"),
		fromPassword: os.Getenv("EMAIL_PASSWORD"),
		smtpHost:     os.Getenv("SMTP_HOST"),
		smtpPort:     os.Getenv("SMTP_PORT"),
	}
}

func (s *EmailService) SendVerificationEmail(toEmail string, verifyToken string) error {
	auth := smtp.PlainAuth("", s.fromEmail, s.fromPassword, s.smtpHost)
	
	// Create verification link
	verifyLink := fmt.Sprintf("http://localhost:8000/verify-email?token=%s", verifyToken)
	
	// Email content
	subject := "Email Verification"
	body := fmt.Sprintf(`
		<html>
			<body>
				<h2>Email Verification</h2>
				<p>Please click the link below to verify your email address:</p>
				<p><a href="%s">Verify Email</a></p>
				<p>This link will expire in 24 hours.</p>
				<p>If you did not request this verification, please ignore this email.</p>
			</body>
		</html>
	`, verifyLink)

	msg := fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", toEmail, subject, body)

	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)
	return smtp.SendMail(addr, auth, s.fromEmail, []string{toEmail}, []byte(msg))
}

func GenerateVerificationToken() string {
	return uuid.New().String()
}

func GetVerificationExpiryTime() time.Time {
	return time.Now().Add(24 * time.Hour)
} 