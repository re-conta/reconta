package email

import (
	"fmt"
	"html"
	"strings"
)

// Cores extraídas de web/src/styles/main.css (--color-brand-* e --color-ink-*),
// para que os e-mails mantenham a identidade visual do site.
const (
	colorBrand500 = "#ff9c43"
	colorBrand600 = "#f2751f"
	colorCoral500 = "#e8496b"
	colorInk900   = "#1c1712"
	colorInk700   = "#453b32"
	colorInk500   = "#7c6d5c"
	colorInk200   = "#e6ded5"
	colorInk100   = "#f4f0ec"
	colorInk50    = "#fbf9f7"

	appName    = "ReConta"
	appTagline = "Suas finanças, organizadas"
)

// Button descreve um botão de call-to-action opcional no corpo do e-mail.
type Button struct {
	Text string
	URL  string
}

// Message descreve o conteúdo de um e-mail transacional, independente do
// canal de envio. Render() produz as versões HTML e texto simples a partir
// dos mesmos dados, mantendo um layout único e consistente para todos os
// e-mails do site.
type Message struct {
	// Preheader é o texto de pré-visualização mostrado por clientes de
	// e-mail antes de abrir a mensagem (ex.: na lista de e-mails do Gmail).
	Preheader string
	// Heading é o título em destaque no topo do corpo do e-mail.
	Heading string
	// Paragraphs são os parágrafos do corpo, renderizados em ordem.
	Paragraphs []string
	// Button, se preenchido, renderiza um botão de ação em destaque.
	Button *Button
	// Footnote é um texto pequeno e discreto após o botão (ex.: aviso de
	// segurança ou instrução alternativa).
	Footnote string
}

// Render gera as versões HTML e texto simples da mensagem, usando o mesmo
// layout de marca (cabeçalho com logo, cartão de conteúdo, rodapé) em todos
// os e-mails do site.
func (m Message) Render() (htmlBody, textBody string) {
	return m.renderHTML(), m.renderText()
}

func (m Message) renderHTML() string {
	var body strings.Builder

	body.WriteString(fmt.Sprintf(`<h1 style="margin:0 0 20px;font-family:'Nunito',Arial,sans-serif;font-size:22px;line-height:1.3;font-weight:800;color:%s;">%s</h1>`, colorInk900, html.EscapeString(m.Heading)))

	for _, p := range m.Paragraphs {
		body.WriteString(fmt.Sprintf(`<p style="margin:0 0 16px;font-family:'Nunito',Arial,sans-serif;font-size:15px;line-height:1.6;color:%s;">%s</p>`, colorInk700, html.EscapeString(p)))
	}

	if m.Button != nil {
		body.WriteString(fmt.Sprintf(`
		<table role="presentation" cellpadding="0" cellspacing="0" border="0" style="margin:28px 0 8px;">
			<tr>
				<td align="center" bgcolor="%s" style="border-radius:999px;background:linear-gradient(135deg,%s,%s);">
					<a href="%s" target="_blank" style="display:inline-block;padding:13px 32px;font-family:'Nunito',Arial,sans-serif;font-size:15px;font-weight:800;color:#ffffff;text-decoration:none;border-radius:999px;">%s</a>
				</td>
			</tr>
		</table>`, colorBrand600, colorBrand500, colorBrand600, m.Button.URL, html.EscapeString(m.Button.Text)))
	}

	if m.Footnote != "" {
		body.WriteString(fmt.Sprintf(`<p style="margin:24px 0 0;font-family:'Nunito',Arial,sans-serif;font-size:13px;line-height:1.6;color:%s;">%s</p>`, colorInk500, html.EscapeString(m.Footnote)))
	}

	return wrapLayout(m.Preheader, body.String())
}

func (m Message) renderText() string {
	var b strings.Builder
	b.WriteString(appName)
	b.WriteString(" — ")
	b.WriteString(appTagline)
	b.WriteString("\n\n")
	b.WriteString(m.Heading)
	b.WriteString("\n\n")

	for _, p := range m.Paragraphs {
		b.WriteString(p)
		b.WriteString("\n\n")
	}

	if m.Button != nil {
		b.WriteString(m.Button.Text)
		b.WriteString(": ")
		b.WriteString(m.Button.URL)
		b.WriteString("\n\n")
	}

	if m.Footnote != "" {
		b.WriteString(m.Footnote)
		b.WriteString("\n\n")
	}

	b.WriteString("--\n")
	b.WriteString(appName + " · reconta.app")

	return strings.TrimSpace(b.String()) + "\n"
}

// wrapLayout monta o esqueleto de e-mail (cabeçalho com marca, cartão de
// conteúdo, rodapé) compartilhado por todas as mensagens do site. Usa
// tabelas e CSS inline, com media query para telas estreitas, garantindo
// compatibilidade com a maioria dos clientes de e-mail (Gmail, Apple Mail,
// Outlook) mantendo o layout responsivo.
func wrapLayout(preheader, bodyHTML string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="pt-BR">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<meta name="color-scheme" content="light">
<meta name="supported-color-schemes" content="light">
<title>%s</title>
<!--[if mso]>
<noscript>
<xml>
<o:OfficeDocumentSettings>
<o:PixelsPerInch>96</o:PixelsPerInch>
</o:OfficeDocumentSettings>
</xml>
</noscript>
<![endif]-->
<style>
  body, table, td, a { -webkit-text-size-adjust:100%%; -ms-text-size-adjust:100%%; }
  table, td { mso-table-lspace:0pt; mso-table-rspace:0pt; }
  img { -ms-interpolation-mode:bicubic; border:0; outline:none; text-decoration:none; }
  body { margin:0; padding:0; width:100%% !important; background-color:%s; }
  a { color:%s; }
  @media screen and (max-width:600px) {
    .email-container { width:100%% !important; }
    .email-padding { padding-left:20px !important; padding-right:20px !important; }
  }
</style>
</head>
<body style="margin:0;padding:0;background-color:%s;">
  <div style="display:none;max-height:0;overflow:hidden;opacity:0;mso-hide:all;">%s</div>
  <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="background-color:%s;">
    <tr>
      <td align="center" style="padding:32px 16px;">
        <table role="presentation" class="email-container" width="600" cellpadding="0" cellspacing="0" border="0" style="width:600px;max-width:600px;">
          <tr>
            <td align="center" style="padding-bottom:24px;">
              <table role="presentation" cellpadding="0" cellspacing="0" border="0">
                <tr>
                  <td style="vertical-align:middle;padding-right:8px;">
                    <table role="presentation" cellpadding="0" cellspacing="0" border="0" width="32" height="32" style="width:32px;height:32px;border-radius:9px;background:linear-gradient(135deg,%s,%s);">
                      <tr><td align="center" style="font-family:'Nunito',Arial,sans-serif;font-size:16px;font-weight:800;color:#ffffff;">R</td></tr>
                    </table>
                  </td>
                  <td style="vertical-align:middle;font-family:'Nunito',Arial,sans-serif;font-size:20px;font-weight:800;color:%s;">%s</td>
                </tr>
              </table>
            </td>
          </tr>
          <tr>
            <td class="email-padding" style="background-color:#ffffff;border:1px solid %s;border-radius:20px;padding:36px;">
              %s
            </td>
          </tr>
          <tr>
            <td align="center" style="padding:24px 16px 0;font-family:'Nunito',Arial,sans-serif;font-size:12px;line-height:1.6;color:%s;">
              %s · <a href="https://reconta.app" target="_blank" style="color:%s;text-decoration:none;">reconta.app</a><br>
              Você recebeu este e-mail porque possui uma conta no %s.
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`,
		html.EscapeString(preheader),
		colorInk50,
		colorBrand600,
		colorInk50,
		html.EscapeString(preheader),
		colorInk50,
		colorBrand500, colorBrand600,
		colorInk900, appName,
		colorInk200,
		bodyHTML,
		colorInk500,
		appName,
		colorBrand600,
		appName,
	)
}
