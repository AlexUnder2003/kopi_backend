package services

import (
	"crypto/tls"
	"fmt"
	"mime"
	"net/smtp"

	"KopiBackend/internal/config"

	"go.uber.org/zap"
)

type SMTPService struct {
	appConfig *config.AppConfig
	logger    *zap.SugaredLogger
}

func NewSMTPService(appConfig *config.AppConfig, logger *zap.SugaredLogger) *SMTPService {
	return &SMTPService{appConfig: appConfig, logger: logger}
}

func (s *SMTPService) Send(to, subject, body string) error {
	from := s.appConfig.SMTPFrom

	addr := fmt.Sprintf("%s:%s", s.appConfig.SMTPHost, s.appConfig.SMTPPort)

	tlsConn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName: s.appConfig.SMTPHost,
	})
	if err != nil {
		s.logger.Errorf("Failed to dial SMTP TLS %s: %v", addr, err)
		return err
	}
	defer tlsConn.Close()

	client, err := smtp.NewClient(tlsConn, s.appConfig.SMTPHost)
	if err != nil {
		return err
	}
	defer client.Close()

	auth := smtp.PlainAuth("", s.appConfig.SMTPUser, s.appConfig.SMTPPassword, s.appConfig.SMTPHost)
	if err := client.Auth(auth); err != nil {
		s.logger.Errorf("SMTP auth failed: %v", err)
		return err
	}

	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	encodedSubject := mime.QEncoding.Encode("utf-8", subject)

	msg := []byte(
		"From: " + from + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + encodedSubject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
			"\r\n" +
			body + "\r\n",
	)
	if _, err := w.Write(msg); err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	if err := client.Quit(); err != nil {
		s.logger.Errorf("Failed to send email to %s: %v", to, err)
		return err
	}

	return nil
}
