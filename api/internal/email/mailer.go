// Package email envia e-mails transacionais (lembretes de contas fixas) via
// SMTP puro (net/smtp), sem depender de provedores externos.
package email

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
)

// Mailer envia e-mails usando as credenciais SMTP configuradas por variáveis
// de ambiente. Se SMTP_HOST não estiver definido, Send é um no-op (apenas loga).
type Mailer struct {
	host string
	port string
	user string
	pass string
	from string
}

// NewFromEnv monta um Mailer a partir das variáveis SMTP_HOST, SMTP_PORT,
// SMTP_USER, SMTP_PASS e SMTP_FROM.
func NewFromEnv() *Mailer {
	return &Mailer{
		host: os.Getenv("SMTP_HOST"),
		port: getEnv("SMTP_PORT", "587"),
		user: os.Getenv("SMTP_USER"),
		pass: os.Getenv("SMTP_PASS"),
		from: getEnv("SMTP_FROM", os.Getenv("SMTP_USER")),
	}
}

// Enabled indica se o envio de e-mail está configurado.
func (m *Mailer) Enabled() bool {
	return m.host != ""
}

// Send envia um e-mail com corpo em texto simples. Se o SMTP não estiver
// configurado, apenas registra a intenção em log e retorna nil.
func (m *Mailer) Send(to, subject, body string) error {
	if !m.Enabled() {
		log.Printf("e-mail não enviado (SMTP não configurado): para=%s assunto=%q", to, subject)
		return nil
	}

	addr := fmt.Sprintf("%s:%s", m.host, m.port)
	var auth smtp.Auth
	if m.user != "" {
		auth = smtp.PlainAuth("", m.user, m.pass, m.host)
	}

	msg := buildMessage(m.from, to, subject, body)
	if err := smtp.SendMail(addr, auth, m.from, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("enviando e-mail via smtp: %w", err)
	}
	return nil
}

func buildMessage(from, to, subject, body string) string {
	var b strings.Builder
	b.WriteString("From: ")
	b.WriteString(from)
	b.WriteString("\r\n")
	b.WriteString("To: ")
	b.WriteString(to)
	b.WriteString("\r\n")
	b.WriteString("Subject: ")
	b.WriteString(subject)
	b.WriteString("\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	b.WriteString("\r\n")
	b.WriteString(body)
	return b.String()
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
