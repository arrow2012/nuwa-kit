package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"time"

	"github.com/arrow2012/nuwa-kit/pkg/log"
	"github.com/arrow2012/nuwa-kit/pkg/metric"
	"github.com/arrow2012/nuwa-kit/pkg/options"
)

// SMTPSender implements Sender using direct SMTP
type SMTPSender struct {
	opts *options.EmailOptions
}

// NewSMTPSender creates a new SMTPSender
func NewSMTPSender(opts *options.EmailOptions) *SMTPSender {
	return &SMTPSender{opts: opts}
}

func (s *SMTPSender) Send(ctx context.Context, to, subject, body string) error {
	return SendSMTP(s.opts, to, subject, body)
}

func (s *SMTPSender) Close() {}

// SendSMTP is a helper to send email via SMTP, used by both Sync and Async senders
func SendSMTP(opts *options.EmailOptions, to, subject, body string) (err error) {
	start := time.Now()
	defer func() {
		status := "success"
		if err != nil {
			status = "failed"
		}
		// Use generic metric from nuwa-kit
		metric.ExternalAPIDuration.WithLabelValues("email", "send", status).Observe(time.Since(start).Seconds())
	}()

	if opts == nil || opts.Host == "" {
		log.Infof("[MockEmail] To: %s | Subject: %s | Body: %s", to, subject, body)
		return nil
	}

	addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	log.Infof("Sending email to %s via %s (SSL: %v)...", to, addr, opts.UseSSL)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: opts.SkipVerify,
		ServerName:         opts.Host,
	}

	var client *smtp.Client

	if opts.UseSSL {
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to dial SMTP (SSL): %w", err)
		}
		client, err = smtp.NewClient(conn, opts.Host)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
	} else {
		client, err = smtp.Dial(addr)
		if err != nil {
			return fmt.Errorf("failed to dial SMTP: %w", err)
		}
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(tlsConfig); err != nil {
				client.Quit()
				return fmt.Errorf("failed to start TLS: %w", err)
			}
		}
	}
	defer client.Quit()

	if opts.Username != "" && opts.Password != "" {
		auth := smtp.PlainAuth("", opts.Username, opts.Password, opts.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP Auth failed: %w", err)
		}
	}

	if err := client.Mail(opts.From); err != nil {
		return fmt.Errorf("SMTP Mail command failed: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("SMTP Rcpt command failed: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP Data command failed: %w", err)
	}

	msg := []byte("To: " + to + "\r\n" +
		"From: " + opts.From + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("failed to write message body: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close message writer: %w", err)
	}

	log.Infof("Email sent successfully to %s", to)
	return nil
}
