package service

import (
	"blog_api/src/model"
	"bytes"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
)

// EmailContent holds the content to send.
type EmailContent struct {
	Subject string
	Body    string
	IsHTML  bool
}

// SendEmail sends an email using the provided config, recipients, and content.
func SendEmail(conf model.EmailConf, to []string, content EmailContent) error {
	if !conf.Enable {
		return fmt.Errorf("email is disabled")
	}
	if conf.Host == "" {
		return fmt.Errorf("email host is required")
	}
	if conf.Port == 0 {
		return fmt.Errorf("email port is required")
	}
	if conf.UserName == "" {
		return fmt.Errorf("email username is required")
	}
	if len(to) == 0 {
		return fmt.Errorf("email recipients are required")
	}

	sender := conf.Sender
	if sender == "" {
		sender = conf.UserName
	}

	msg, err := buildEmailMessage(sender, to, content)
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	auth := smtp.PlainAuth("", conf.UserName, conf.Password, conf.Host)

	if conf.Port == 465 {
		return sendWithImplicitTLS(addr, conf.Host, sender, to, auth, msg)
	}
	return sendWithSTARTTLS(addr, conf.Host, sender, to, auth, msg)
}

func buildEmailMessage(sender string, to []string, content EmailContent) ([]byte, error) {
	if content.Subject == "" {
		return nil, fmt.Errorf("email subject is required")
	}
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("From: %s\r\n", sender))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", content.Subject))
	if content.IsHTML {
		buf.WriteString("MIME-Version: 1.0\r\n")
		buf.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	} else {
		buf.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	}
	buf.WriteString("\r\n")
	buf.WriteString(content.Body)
	return buf.Bytes(), nil
}

// doSend performs the SMTP transaction after a connection is established.
func doSend(client *smtp.Client, sender string, to []string, auth smtp.Auth, msg []byte) error {
	if auth != nil {
		if ok, _ := client.Extension("AUTH"); ok {
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("email auth failed: %w", err)
			}
		}
	}
	if err := client.Mail(sender); err != nil {
		return fmt.Errorf("email mail from failed: %w", err)
	}
	for _, rcpt := range to {
		if err := client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("email rcpt failed: %w", err)
		}
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("email data failed: %w", err)
	}
	if _, err := writer.Write(msg); err != nil {
		return fmt.Errorf("email write failed: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("email close failed: %w", err)
	}
	return client.Quit()
}

// sendWithImplicitTLS sends an email over a connection that is encrypted from the start (typically port 465).
func sendWithImplicitTLS(addr, host, sender string, to []string, auth smtp.Auth, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return fmt.Errorf("email tls dial failed: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("email client failed: %w", err)
	}
	defer client.Close()
	return doSend(client, sender, to, auth, msg)
}

// sendWithSTARTTLS sends an email over a plain text connection that may be upgraded to TLS.
func sendWithSTARTTLS(addr, host, sender string, to []string, auth smtp.Auth, msg []byte) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("email dial failed: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: host}); err != nil {
			return fmt.Errorf("email starttls failed: %w", err)
		}
	}
	return doSend(client, sender, to, auth, msg)
}
