package mail

import (
	"fmt"
	"html"
	"strings"
)

// Shared layout tokens (inline CSS for email clients).
const (
	fontSans    = "-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,'Helvetica Neue',Arial,sans-serif"
	colorInk    = "#18181b"
	colorBody   = "#3f3f46"
	colorMuted  = "#71717a"
	colorFaint  = "#a1a1aa"
	colorBorder = "#e4e4e7"
	bgPage      = "#f4f4f5"
	bgButton    = "#18181b"
	fgButton    = "#fafafa"

	// BrandName is the product name used in transactional templates.
	// Override per project — keep short so subject lines stay clean.
	BrandName = "App"
)

func transactionalShell(preheader string, cardInnerHTML string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head>
<body style="margin:0;padding:0;background-color:%s;">
<span style="display:none!important;visibility:hidden;opacity:0;height:0;width:0;overflow:hidden;">%s</span>
<table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="background-color:%s;padding:40px 16px;">
<tr><td align="center">
<table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="max-width:560px;background-color:#ffffff;border-radius:12px;overflow:hidden;box-shadow:0 1px 3px rgba(0,0,0,0.06);">
%s
</table>
</td></tr>
</table>
</body>
</html>`,
		bgPage, html.EscapeString(preheader), bgPage, cardInnerHTML)
}

func brandAndHeading(kicker, title string) string {
	return fmt.Sprintf(
		`<tr><td style="padding:32px 40px 8px 40px;font-family:%s;">
<p style="margin:0;font-size:13px;font-weight:600;letter-spacing:0.02em;color:%s;text-transform:uppercase;">%s</p>
<h1 style="margin:12px 0 0 0;font-size:22px;font-weight:600;line-height:1.3;color:%s;">%s</h1>
</td></tr>`,
		fontSans, colorMuted, html.EscapeString(kicker), colorInk, html.EscapeString(title))
}

func proseParagraph(text string) string {
	return fmt.Sprintf(
		`<tr><td style="padding:8px 40px 0 40px;font-family:%s;font-size:16px;line-height:1.6;color:%s;">
<p style="margin:0 0 16px 0;">%s</p>
</td></tr>`,
		fontSans, colorBody, html.EscapeString(text))
}

func primaryButton(href, label string) string {
	return fmt.Sprintf(
		`<tr><td style="padding:8px 40px 28px 40px;">
<table role="presentation" cellspacing="0" cellpadding="0"><tr><td style="border-radius:8px;background-color:%s;">
<a href="%s" target="_blank" rel="noopener noreferrer" style="display:inline-block;padding:14px 28px;font-family:%s;font-size:15px;font-weight:600;color:%s;text-decoration:none;border-radius:8px;">%s</a>
</td></tr></table>
</td></tr>`,
		bgButton, html.EscapeString(href), fontSans, fgButton, html.EscapeString(label))
}

func optionalRoleRow(roleName string) string {
	roleName = strings.TrimSpace(roleName)
	if roleName == "" {
		return ""
	}
	return fmt.Sprintf(
		`<tr><td style="padding:0 40px 24px 40px;font-family:%s;font-size:15px;line-height:1.5;color:%s;">Your role: <strong style="color:%s;">%s</strong></td></tr>`,
		fontSans, colorBody, colorInk, html.EscapeString(roleName))
}

func expiryAndLinkFallback(expiryLine string, link string) string {
	return fmt.Sprintf(
		`<tr><td style="padding:0 40px 32px 40px;font-family:%s;font-size:13px;line-height:1.5;color:%s;">
<p style="margin:0 0 8px 0;">%s</p>
<p style="margin:0 0 16px 0;">If the button doesn’t work, paste this URL into your browser:</p>
<p style="margin:0;word-break:break-all;font-size:12px;color:%s;">%s</p>
</td></tr>`,
		fontSans, colorMuted, expiryLine, colorFaint, html.EscapeString(link))
}

func cardFooter(note string) string {
	return fmt.Sprintf(
		`<tr><td style="padding:20px 40px;border-top:1px solid %s;font-family:%s;font-size:12px;line-height:1.5;color:%s;">
<p style="margin:0;">%s</p>
</td></tr>`,
		colorBorder, fontSans, colorFaint, html.EscapeString(note))
}

// EmailSimpleNotification returns HTML for a generic in-app notification email.
func EmailSimpleNotification(title, message string) string {
	inner := strings.Builder{}
	inner.WriteString(brandAndHeading(BrandName, title))
	if message != "" {
		inner.WriteString(proseParagraph(message))
	}
	inner.WriteString(cardFooter("This is an automated notification."))
	return transactionalShell(title, inner.String())
}

// EmailInvitation returns HTML for the team-invite flow.
func EmailInvitation(link, expiresAtUTC, roleName string) string {
	inner := strings.Builder{}
	inner.WriteString(brandAndHeading(BrandName, "You're invited"))
	inner.WriteString(proseParagraph("Someone on your team invited you. Use the button below to create your account and set a password."))
	inner.WriteString(optionalRoleRow(roleName))
	inner.WriteString(primaryButton(link, "Accept invitation"))
	expiryHTML := fmt.Sprintf(
		`This secure link expires on <strong style="color:#52525b;">%s</strong> (UTC).`,
		html.EscapeString(expiresAtUTC))
	inner.WriteString(expiryAndLinkFallback(expiryHTML, link))
	inner.WriteString(cardFooter("If you didn't expect this invitation, ignore this email."))
	return transactionalShell("Accept your invitation.", inner.String())
}

// EmailPasswordReset returns HTML for the forgot-password flow.
func EmailPasswordReset(resetLink string) string {
	inner := strings.Builder{}
	inner.WriteString(brandAndHeading(BrandName, "Reset your password"))
	inner.WriteString(proseParagraph("We received a request to reset the password for your account. If you made this request, use the button below. If not, ignore this email."))
	inner.WriteString(primaryButton(resetLink, "Reset password"))
	expiryHTML := `This link expires in <strong style="color:#52525b;">one hour</strong>.`
	inner.WriteString(expiryAndLinkFallback(expiryHTML, resetLink))
	inner.WriteString(cardFooter("If you didn't request a password reset, no action is needed."))
	return transactionalShell("Reset your password.", inner.String())
}

// EmailWelcome — first sign-in landed; sent on tenant creation / first invite
// acceptance. Lightweight, no buttons.
func EmailWelcome(displayName string) string {
	greeting := "Welcome aboard"
	if strings.TrimSpace(displayName) != "" {
		greeting = "Welcome, " + displayName
	}
	inner := strings.Builder{}
	inner.WriteString(brandAndHeading(BrandName, greeting))
	inner.WriteString(proseParagraph("Your account is ready. Sign in any time to manage your workspace, invite teammates, and configure billing."))
	inner.WriteString(cardFooter("Questions? Reply to this email and we'll help."))
	return transactionalShell("Welcome to "+BrandName, inner.String())
}

// EmailPaymentReceipt — sent when a payment is recorded against an invoice.
// The receipt PDF is the attachment; this is the cover email.
func EmailPaymentReceipt(receiptNumber, invoiceNumber, amountFormatted, method, paidAt string) string {
	inner := strings.Builder{}
	inner.WriteString(brandAndHeading(BrandName, "Payment received"))
	inner.WriteString(proseParagraph(fmt.Sprintf(
		"Thanks — we received your payment of %s on %s against invoice %s. Receipt %s is attached as a PDF for your records.",
		amountFormatted, paidAt, invoiceNumber, receiptNumber)))
	inner.WriteString(fmt.Sprintf(
		`<tr><td style="padding:0 40px 24px 40px;font-family:%s;font-size:14px;color:%s;">
<table role="presentation" cellspacing="0" cellpadding="0" width="100%%" style="border:1px solid %s;border-radius:8px;">
<tr><td style="padding:14px 16px;border-bottom:1px solid %s;"><strong style="color:%s;">Receipt #</strong> <span style="color:%s;">%s</span></td></tr>
<tr><td style="padding:14px 16px;border-bottom:1px solid %s;"><strong style="color:%s;">Invoice</strong> <span style="color:%s;">%s</span></td></tr>
<tr><td style="padding:14px 16px;border-bottom:1px solid %s;"><strong style="color:%s;">Amount</strong> <span style="color:%s;">%s</span></td></tr>
<tr><td style="padding:14px 16px;"><strong style="color:%s;">Method</strong> <span style="color:%s;">%s</span></td></tr>
</table>
</td></tr>`,
		fontSans, colorBody, colorBorder,
		colorBorder, colorInk, colorBody, html.EscapeString(receiptNumber),
		colorBorder, colorInk, colorBody, html.EscapeString(invoiceNumber),
		colorBorder, colorInk, colorBody, html.EscapeString(amountFormatted),
		colorInk, colorBody, html.EscapeString(method)))
	inner.WriteString(cardFooter("Reply to this email if anything looks off."))
	return transactionalShell("Receipt "+receiptNumber, inner.String())
}

// EmailInvoiceIssued — sent when a new cycle invoice is generated.
func EmailInvoiceIssued(invoiceNumber, dueDate, amountFormatted, viewLink string) string {
	inner := strings.Builder{}
	inner.WriteString(brandAndHeading(BrandName, "New invoice: "+invoiceNumber))
	inner.WriteString(proseParagraph(fmt.Sprintf(
		"Your latest invoice for %s is due on %s. The full PDF is attached. You can also view and pay it online.",
		amountFormatted, dueDate)))
	if viewLink != "" {
		inner.WriteString(primaryButton(viewLink, "View invoice"))
	}
	inner.WriteString(cardFooter("Need a copy or have questions? Reply to this email."))
	return transactionalShell("Invoice "+invoiceNumber+" — "+amountFormatted, inner.String())
}
