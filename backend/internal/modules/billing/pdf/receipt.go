package pdf

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/signintech/gopdf"
)

// RenderReceipt produces a compact A4 payment receipt (about 1/2 page used).
// Layout:
//   • Dark header band with company + "PAYMENT RECEIPT" + receipt number.
//   • Customer + invoice reference block.
//   • Amount paid (large, prominent).
//   • Payment method + reference + paid-at line.
//   • Footer with GSTIN.
//
// Compared to the invoice PDF, this one omits line items and tax breakdown —
// those live on the invoice. The receipt just confirms the payment happened.
func (r *GopdfRenderer) RenderReceipt(in ReceiptInput) ([]byte, error) {
	doc := gopdf.GoPdf{}
	doc.Start(gopdf.Config{PageSize: gopdf.Rect{W: pageWidth, H: pageHeight}})
	doc.AddPage()
	if err := loadFonts(&doc); err != nil {
		return nil, err
	}

	// ── header band ──
	doc.SetFillColor(0x10, 0x18, 0x28)
	doc.RectFromUpperLeftWithStyle(0, 0, pageWidth, 80, "F")
	doc.SetTextColor(255, 255, 255)
	must(doc.SetFont("regular-bold", "", 18))
	doc.SetXY(margin, 24)
	must(doc.Cell(nil, ifEmpty(in.CompanyName, "Your Company")))
	must(doc.SetFont("regular", "", 9))
	doc.SetXY(margin, 46)
	if in.CompanyGSTIN != "" {
		must(doc.Cell(nil, "GSTIN: "+in.CompanyGSTIN))
	}
	if in.CompanyAddress != "" {
		doc.SetXY(margin, 58)
		must(doc.Cell(nil, truncate(in.CompanyAddress, 90)))
	}

	must(doc.SetFont("regular-bold", "", 16))
	titleW, _ := doc.MeasureTextWidth("PAYMENT RECEIPT")
	doc.SetXY(pageWidth-margin-titleW, 24)
	must(doc.Cell(nil, "PAYMENT RECEIPT"))
	must(doc.SetFont("regular", "", 9))
	rcpStr := "#" + in.ReceiptNumber
	rw, _ := doc.MeasureTextWidth(rcpStr)
	doc.SetXY(pageWidth-margin-rw, 46)
	must(doc.Cell(nil, rcpStr))

	doc.SetTextColor(textColor, textColor, textColor)
	y := 110.0

	// ── customer + invoice reference ──
	must(doc.SetFont("regular-bold", "", 9))
	doc.SetTextColor(subtleColor, subtleColor, subtleColor)
	doc.SetXY(margin, y)
	must(doc.Cell(nil, "RECEIVED FROM"))
	doc.SetXY(pageWidth/2, y)
	must(doc.Cell(nil, "AGAINST INVOICE"))
	y += 14
	must(doc.SetFont("regular-bold", "", 13))
	doc.SetTextColor(textColor, textColor, textColor)
	doc.SetXY(margin, y)
	must(doc.Cell(nil, ifEmpty(in.CustomerName, "—")))
	doc.SetXY(pageWidth/2, y)
	must(doc.Cell(nil, in.InvoiceNumber))
	y += 18
	must(doc.SetFont("regular", "", 10))
	if in.CustomerEmail != "" {
		doc.SetTextColor(subtleColor, subtleColor, subtleColor)
		doc.SetXY(margin, y)
		must(doc.Cell(nil, in.CustomerEmail))
		y += 14
	}
	y += 12

	// ── amount paid box ──
	doc.SetFillColor(0x18, 0xA0, 0x58)
	doc.RectFromUpperLeftWithStyle(margin, y, pageWidth-2*margin, 90, "F")
	doc.SetTextColor(255, 255, 255)
	must(doc.SetFont("regular", "", 11))
	doc.SetXY(margin+20, y+16)
	must(doc.Cell(nil, "AMOUNT RECEIVED"))
	must(doc.SetFont("regular-bold", "", 36))
	amt := money(in.Currency, in.AmountCents)
	doc.SetXY(margin+20, y+34)
	must(doc.Cell(nil, amt))
	must(doc.SetFont("regular", "", 10))
	doc.SetXY(margin+20, y+76)
	must(doc.Cell(nil, "Paid on "+in.PaidAt.Format("02 Jan 2006, 15:04 MST")))
	y += 110

	doc.SetTextColor(textColor, textColor, textColor)

	// ── method + reference table ──
	must(doc.SetFont("regular-bold", "", 9))
	doc.SetTextColor(subtleColor, subtleColor, subtleColor)
	rows := [][2]string{
		{"PAYMENT METHOD", methodLabel(in.Method)},
	}
	if in.Reference != "" {
		rows = append(rows, [2]string{methodReferenceLabel(in.Method), in.Reference})
	}
	rows = append(rows, [2]string{"RECEIPT NUMBER", in.ReceiptNumber})
	for _, row := range rows {
		must(doc.SetFont("regular", "", 8))
		doc.SetTextColor(subtleColor, subtleColor, subtleColor)
		doc.SetXY(margin, y)
		must(doc.Cell(nil, row[0]))
		must(doc.SetFont("regular-bold", "", 11))
		doc.SetTextColor(textColor, textColor, textColor)
		doc.SetXY(margin+180, y-2)
		must(doc.Cell(nil, row[1]))
		y += 22
	}

	// ── footer ──
	footerY := pageHeight - margin - 60
	doc.SetStrokeColor(gridColor, gridColor, gridColor)
	doc.Line(margin, footerY, pageWidth-margin, footerY)
	footerY += 8
	must(doc.SetFont("regular", "", 8))
	doc.SetTextColor(subtleColor, subtleColor, subtleColor)
	doc.SetXY(margin, footerY)
	must(doc.Cell(nil, "Thank you for your payment."))
	if in.CompanyGSTIN != "" {
		doc.SetXY(margin, footerY+14)
		must(doc.Cell(nil, "GSTIN: "+in.CompanyGSTIN))
	}
	doc.SetXY(margin, pageHeight-margin+4)
	must(doc.SetFont("regular", "", 7))
	must(doc.Cell(nil, "Computer-generated receipt — no signature required."))

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		return nil, fmt.Errorf("receipt pdf write: %w", err)
	}
	return buf.Bytes(), nil
}

func methodLabel(method string) string {
	switch strings.ToLower(method) {
	case "cash":
		return "Cash"
	case "bank_transfer":
		return "Bank transfer"
	case "cheque":
		return "Cheque"
	case "gateway":
		return "Online payment gateway"
	default:
		return method
	}
}

func methodReferenceLabel(method string) string {
	switch strings.ToLower(method) {
	case "bank_transfer":
		return "TRANSACTION REF"
	case "cheque":
		return "CHEQUE NUMBER"
	case "gateway":
		return "GATEWAY REF"
	default:
		return "REFERENCE"
	}
}
