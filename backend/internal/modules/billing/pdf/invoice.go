package pdf

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/signintech/gopdf"
)

//go:embed fonts/Geist-Regular.ttf
var geistRegular []byte

// GopdfRenderer renders billing PDFs using the pure-Go signintech/gopdf
// library. It ships built-in Helvetica fonts so the binary stays self-contained.
type GopdfRenderer struct{}

func NewRenderer() *GopdfRenderer { return &GopdfRenderer{} }

// Margins and palette — adjusting these is the cheapest way to retune the
// look without touching the row-by-row drawing code.
const (
	pageWidth     = 595.0  // A4 portrait width in points (gopdf uses points)
	pageHeight    = 842.0
	margin        = 36.0   // 0.5 inch all-around
	headerBg      = 0x14   // dark slate grey hex byte
	gridColor     = 0xDD
	textColor     = 0x18   // near-black
	subtleColor   = 0x66
	successColor  = 0x18A058 // green for paid stamps
	warnColor     = 0xCC6600 // amber for open invoices
)

// RenderInvoice produces an A4 tax invoice. Layout summary:
//   • Header band with company name, GSTIN, "TAX INVOICE" + status pill.
//   • Two-column block: "From" (us) | "Bill to" (customer).
//   • Meta row: invoice number, issued, due, period, place of supply.
//   • Line-items table with HSN/SAC, qty, rate, taxable, GST split.
//   • Totals box on the right + amount-in-words.
//   • Footer: payment terms + bank details.
//
// The page is sized to fit the standard amount of line items (≤ 12) on one
// page. Future improvement: paginate when more lines are present.
func (r *GopdfRenderer) RenderInvoice(in InvoiceInput) ([]byte, error) {
	doc := gopdf.GoPdf{}
	doc.Start(gopdf.Config{PageSize: gopdf.Rect{W: pageWidth, H: pageHeight}})
	doc.AddPage()
	if err := loadFonts(&doc); err != nil {
		return nil, err
	}

	yCursor := margin

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

	// Right-aligned invoice title + status pill.
	must(doc.SetFont("regular-bold", "", 16))
	titleW, _ := doc.MeasureTextWidth("TAX INVOICE")
	doc.SetXY(pageWidth-margin-titleW, 24)
	must(doc.Cell(nil, "TAX INVOICE"))
	statusLabel, statusR, statusG, statusB := statusPill(in.Status)
	must(doc.SetFont("regular-bold", "", 9))
	pillW, _ := doc.MeasureTextWidth(statusLabel)
	pillW += 16
	doc.SetFillColor(statusR, statusG, statusB)
	doc.RectFromUpperLeftWithStyle(pageWidth-margin-pillW, 46, pillW, 16, "F")
	doc.SetTextColor(255, 255, 255)
	doc.SetXY(pageWidth-margin-pillW+8, 49)
	must(doc.Cell(nil, statusLabel))

	doc.SetTextColor(textColor, textColor, textColor)
	yCursor = 100

	// ── From / Bill to columns ──
	colW := (pageWidth - 2*margin - 16) / 2
	leftX := margin
	rightX := margin + colW + 16
	must(doc.SetFont("regular-bold", "", 9))
	doc.SetXY(leftX, yCursor)
	must(doc.Cell(nil, "FROM"))
	doc.SetXY(rightX, yCursor)
	must(doc.Cell(nil, "BILL TO"))
	yCursor += 14

	yLeft := yCursor
	yRight := yCursor
	must(doc.SetFont("regular-bold", "", 11))
	doc.SetXY(leftX, yLeft)
	must(doc.Cell(nil, in.CompanyName))
	yLeft += 14
	must(doc.SetFont("regular", "", 9))
	if in.CompanyAddress != "" {
		yLeft = drawWrapped(&doc, leftX, yLeft, colW, in.CompanyAddress)
	}
	if in.CompanyGSTIN != "" {
		doc.SetXY(leftX, yLeft)
		must(doc.Cell(nil, "GSTIN: "+in.CompanyGSTIN))
		yLeft += 12
	}

	must(doc.SetFont("regular-bold", "", 11))
	doc.SetXY(rightX, yRight)
	must(doc.Cell(nil, ifEmpty(in.CustomerName, "—")))
	yRight += 14
	must(doc.SetFont("regular", "", 9))
	if in.CustomerEmail != "" {
		doc.SetXY(rightX, yRight)
		must(doc.Cell(nil, in.CustomerEmail))
		yRight += 12
	}
	if in.CustomerAddress != "" {
		yRight = drawWrapped(&doc, rightX, yRight, colW, in.CustomerAddress)
	}
	if in.CustomerGSTIN != "" {
		doc.SetXY(rightX, yRight)
		must(doc.Cell(nil, "GSTIN: "+in.CustomerGSTIN))
		yRight += 12
	}

	yCursor = maxF(yLeft, yRight) + 12

	// ── meta strip ──
	doc.SetFillColor(0xF5, 0xF7, 0xFA)
	doc.RectFromUpperLeftWithStyle(margin, yCursor, pageWidth-2*margin, 48, "F")
	must(doc.SetFont("regular", "", 8))
	doc.SetTextColor(subtleColor, subtleColor, subtleColor)
	metaY := yCursor + 8
	cellW := (pageWidth - 2*margin) / 5
	metaCells := [][2]string{
		{"INVOICE #", in.InvoiceNumber},
		{"ISSUED", in.IssuedAt.Format("02 Jan 2006")},
		{"DUE", in.DueAt.Format("02 Jan 2006")},
		{"PERIOD", in.PeriodStart.Format("02 Jan") + " — " + in.PeriodEnd.Format("02 Jan 2006")},
		{"PLACE OF SUPPLY", ifEmpty(in.PlaceOfSupply, "—")},
	}
	for i, c := range metaCells {
		x := margin + float64(i)*cellW + 10
		must(doc.SetFont("regular", "", 7))
		doc.SetTextColor(subtleColor, subtleColor, subtleColor)
		doc.SetXY(x, metaY)
		must(doc.Cell(nil, c[0]))
		must(doc.SetFont("regular-bold", "", 9))
		doc.SetTextColor(textColor, textColor, textColor)
		doc.SetXY(x, metaY+14)
		must(doc.Cell(nil, c[1]))
	}
	yCursor += 60

	// ── line items table ──
	yCursor = drawLineItems(&doc, yCursor, in)

	// ── totals box (right-aligned) ──
	yCursor += 8
	totalsW := 220.0
	totalsX := pageWidth - margin - totalsW
	totalsY := yCursor

	must(doc.SetFont("regular", "", 9))
	doc.SetTextColor(textColor, textColor, textColor)
	yCursor = drawTotalsRow(&doc, totalsX, totalsW, yCursor, "Subtotal", money(in.Currency, in.SubtotalCents))
	if in.DiscountCents > 0 {
		yCursor = drawTotalsRow(&doc, totalsX, totalsW, yCursor, "Discount", "-"+money(in.Currency, in.DiscountCents))
	}
	if in.IntraState {
		yCursor = drawTotalsRow(&doc, totalsX, totalsW, yCursor, "CGST", money(in.Currency, in.CGSTCents))
		yCursor = drawTotalsRow(&doc, totalsX, totalsW, yCursor, "SGST", money(in.Currency, in.SGSTCents))
	} else {
		yCursor = drawTotalsRow(&doc, totalsX, totalsW, yCursor, "IGST", money(in.Currency, in.IGSTCents))
	}
	// Grand total — bold, bigger, with a line above.
	doc.SetStrokeColor(gridColor, gridColor, gridColor)
	doc.Line(totalsX, yCursor+2, totalsX+totalsW, yCursor+2)
	yCursor += 6
	must(doc.SetFont("regular-bold", "", 12))
	doc.SetXY(totalsX, yCursor)
	must(doc.Cell(nil, "Total"))
	totW, _ := doc.MeasureTextWidth(money(in.Currency, in.TotalCents))
	doc.SetXY(totalsX+totalsW-totW, yCursor)
	must(doc.Cell(nil, money(in.Currency, in.TotalCents)))
	yCursor += 18
	if in.AmountPaidCents > 0 && in.AmountPaidCents < in.TotalCents {
		must(doc.SetFont("regular", "", 9))
		yCursor = drawTotalsRow(&doc, totalsX, totalsW, yCursor, "Paid", money(in.Currency, in.AmountPaidCents))
		must(doc.SetFont("regular-bold", "", 10))
		yCursor = drawTotalsRow(&doc, totalsX, totalsW, yCursor, "Balance due", money(in.Currency, in.AmountDueCents))
	}

	// Amount-in-words below totals, left side.
	must(doc.SetFont("regular", "", 9))
	doc.SetTextColor(subtleColor, subtleColor, subtleColor)
	doc.SetXY(margin, totalsY)
	must(doc.Cell(nil, "Amount in words:"))
	must(doc.SetFont("regular-bold", "", 10))
	doc.SetTextColor(textColor, textColor, textColor)
	doc.SetXY(margin, totalsY+14)
	words := indianRupeeWords(in.TotalCents)
	drawWrapped(&doc, margin, totalsY+14, pageWidth-2*margin-totalsW-16, words)

	// ── footer: payment terms + bank details ──
	footerY := pageHeight - margin - 80
	doc.SetStrokeColor(gridColor, gridColor, gridColor)
	doc.Line(margin, footerY, pageWidth-margin, footerY)
	footerY += 8
	must(doc.SetFont("regular-bold", "", 8))
	doc.SetTextColor(subtleColor, subtleColor, subtleColor)
	doc.SetXY(margin, footerY)
	must(doc.Cell(nil, "PAYMENT TERMS"))
	doc.SetXY(margin+220, footerY)
	must(doc.Cell(nil, "BANK DETAILS"))
	footerY += 12

	must(doc.SetFont("regular", "", 9))
	doc.SetTextColor(textColor, textColor, textColor)
	drawWrapped(&doc, margin, footerY, 200, ifEmpty(in.InvoiceTerms, "Payment due as per due date."))
	bankLines := buildBankLines(in)
	for i, line := range bankLines {
		doc.SetXY(margin+220, footerY+float64(i)*12)
		must(doc.Cell(nil, line))
	}

	must(doc.SetFont("regular", "", 7))
	doc.SetTextColor(subtleColor, subtleColor, subtleColor)
	doc.SetXY(margin, pageHeight-margin+4)
	must(doc.Cell(nil, "Computer-generated invoice — no signature required."))

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		return nil, fmt.Errorf("pdf write: %w", err)
	}
	return buf.Bytes(), nil
}

// drawLineItems draws the table starting at yStart and returns the new Y.
func drawLineItems(doc *gopdf.GoPdf, yStart float64, in InvoiceInput) float64 {
	rowH := 24.0
	headerH := 18.0
	contentW := pageWidth - 2*margin

	// Column widths sum to contentW. Description gets the rest.
	// description, hsn, qty, rate, taxable, tax, total
	// Bug fix: the TOTAL column had width 0, so its header and values rendered
	// on top of the tax column — that's why the last column read "CGSTTSGSI"
	// with double rupee signs in every row.
	cols := []float64{0, 50, 30, 65, 70, 70, 65}
	used := 0.0
	for i, w := range cols {
		if i == 0 {
			continue
		}
		used += w
	}
	cols[0] = contentW - used // description takes leftover

	// header
	doc.SetFillColor(0xF5, 0xF7, 0xFA)
	doc.RectFromUpperLeftWithStyle(margin, yStart, contentW, headerH, "F")
	must(doc.SetFont("regular-bold", "", 8))
	doc.SetTextColor(subtleColor, subtleColor, subtleColor)
	headers := []string{"DESCRIPTION", "HSN/SAC", "QTY", "RATE", "TAXABLE", taxColLabel(in.IntraState), "TOTAL"}
	x := margin
	for i, h := range headers {
		doc.SetXY(x+6, yStart+5)
		if i >= 2 { // numeric columns right-align
			textW, _ := doc.MeasureTextWidth(h)
			doc.SetXY(x+cols[i]-textW-6, yStart+5)
		}
		must(doc.Cell(nil, h))
		x += cols[i]
	}

	// rows
	doc.SetStrokeColor(gridColor, gridColor, gridColor)
	y := yStart + headerH
	for _, line := range in.Lines {
		// Stripe alt rows for readability.
		if (int(y/rowH))%2 == 0 {
			doc.SetFillColor(0xFC, 0xFC, 0xFD)
			doc.RectFromUpperLeftWithStyle(margin, y, contentW, rowH, "F")
		}
		must(doc.SetFont("regular", "", 9))
		doc.SetTextColor(textColor, textColor, textColor)
		x = margin
		// description
		doc.SetXY(x+6, y+7)
		must(doc.Cell(nil, truncate(line.Description, 50)))
		x += cols[0]
		// HSN
		doc.SetXY(x+6, y+7)
		must(doc.Cell(nil, line.HSNSAC))
		x += cols[1]
		// qty
		writeNum(doc, x, cols[2], y+7, fmt.Sprintf("%d", line.Quantity))
		x += cols[2]
		// rate (unit)
		writeNum(doc, x, cols[3], y+7, money(in.Currency, line.UnitPriceCents))
		x += cols[3]
		// taxable
		writeNum(doc, x, cols[4], y+7, money(in.Currency, line.TaxableAmountCents))
		x += cols[4]
		// tax value
		taxAmt := line.IGSTCents
		if in.IntraState {
			taxAmt = line.CGSTCents + line.SGSTCents
		}
		writeNum(doc, x, cols[5], y+7, money(in.Currency, taxAmt))
		x += cols[5]
		// total
		writeNum(doc, x, cols[6], y+7, money(in.Currency, line.TotalCents))
		// row separator
		doc.Line(margin, y+rowH, margin+contentW, y+rowH)
		y += rowH
	}
	return y
}

func drawTotalsRow(doc *gopdf.GoPdf, x, w, y float64, label, value string) float64 {
	doc.SetXY(x, y)
	must(doc.Cell(nil, label))
	vw, _ := doc.MeasureTextWidth(value)
	doc.SetXY(x+w-vw, y)
	must(doc.Cell(nil, value))
	return y + 16
}

// drawWrapped draws text wrapping at the given max width and returns the new Y.
func drawWrapped(doc *gopdf.GoPdf, x, y, maxW float64, text string) float64 {
	for _, line := range wrapLines(doc, text, maxW) {
		doc.SetXY(x, y)
		must(doc.Cell(nil, line))
		y += 12
	}
	return y
}

func wrapLines(doc *gopdf.GoPdf, text string, maxW float64) []string {
	var out []string
	for _, hard := range strings.Split(text, "\n") {
		words := strings.Fields(hard)
		current := ""
		for _, w := range words {
			candidate := current
			if candidate != "" {
				candidate += " "
			}
			candidate += w
			if width, _ := doc.MeasureTextWidth(candidate); width > maxW && current != "" {
				out = append(out, current)
				current = w
				continue
			}
			current = candidate
		}
		if current != "" {
			out = append(out, current)
		}
	}
	if len(out) == 0 {
		out = []string{""}
	}
	return out
}

func writeNum(doc *gopdf.GoPdf, x, w, y float64, s string) {
	tw, _ := doc.MeasureTextWidth(s)
	doc.SetXY(x+w-tw-6, y)
	must(doc.Cell(nil, s))
}

func taxColLabel(intraState bool) string {
	if intraState {
		return "CGST+SGST"
	}
	return "IGST"
}

func statusPill(status string) (string, uint8, uint8, uint8) {
	switch strings.ToLower(status) {
	case "paid":
		return "PAID", 0x18, 0xA0, 0x58
	case "void", "voided":
		return "VOID", 0x88, 0x88, 0x88
	case "refunded":
		return "REFUNDED", 0x66, 0x66, 0xAA
	default:
		return "DUE", 0xCC, 0x66, 0x00
	}
}

func buildBankLines(in InvoiceInput) []string {
	out := []string{}
	if in.BankName != "" {
		out = append(out, "Bank: "+in.BankName)
	}
	if in.BankAccountName != "" {
		out = append(out, "Account name: "+in.BankAccountName)
	}
	if in.BankAccount != "" {
		out = append(out, "A/C No: "+in.BankAccount)
	}
	if in.BankIFSC != "" {
		out = append(out, "IFSC: "+in.BankIFSC)
	}
	if len(out) == 0 {
		out = []string{"Bank details on request."}
	}
	return out
}

// money formats cents into a currency string. Supports INR rupee symbol;
// other currencies get the 3-letter code.
func money(currency string, cents int64) string {
	prefix := currency + " "
	if strings.EqualFold(currency, "INR") {
		prefix = "₹"
	}
	neg := cents < 0
	if neg {
		cents = -cents
	}
	rupees := cents / 100
	paise := cents % 100
	// Indian grouping: 1,23,45,678.90 — easier as a custom formatter.
	rs := indianGroup(rupees)
	out := fmt.Sprintf("%s%s.%02d", prefix, rs, paise)
	if neg {
		out = "-" + out
	}
	return out
}

func indianGroup(n int64) string {
	if n == 0 {
		return "0"
	}
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	last3 := s[len(s)-3:]
	rest := s[:len(s)-3]
	var grouped []string
	for len(rest) > 2 {
		grouped = append([]string{rest[len(rest)-2:]}, grouped...)
		rest = rest[:len(rest)-2]
	}
	if rest != "" {
		grouped = append([]string{rest}, grouped...)
	}
	return strings.Join(grouped, ",") + "," + last3
}

// indianRupeeWords converts cents to "Rupees X and Paise Y only" Indian-style.
// Caps at 999,99,99,999.99 (nine hundred ninety-nine crore).
func indianRupeeWords(cents int64) string {
	rupees := cents / 100
	paise := cents % 100
	words := numberToWords(rupees)
	out := "Rupees " + words
	if paise > 0 {
		out += " and Paise " + numberToWords(paise)
	}
	return out + " only."
}

var ones = []string{
	"Zero", "One", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight", "Nine",
	"Ten", "Eleven", "Twelve", "Thirteen", "Fourteen", "Fifteen",
	"Sixteen", "Seventeen", "Eighteen", "Nineteen",
}
var tens = []string{"", "", "Twenty", "Thirty", "Forty", "Fifty", "Sixty", "Seventy", "Eighty", "Ninety"}

func numberToWords(n int64) string {
	if n < 20 {
		return ones[n]
	}
	if n < 100 {
		if n%10 == 0 {
			return tens[n/10]
		}
		return tens[n/10] + " " + ones[n%10]
	}
	if n < 1000 {
		return ones[n/100] + " Hundred " + maybe(numberToWords(n%100))
	}
	if n < 100000 { // < 1 lakh
		return numberToWords(n/1000) + " Thousand " + maybe(numberToWords(n%1000))
	}
	if n < 10000000 { // < 1 crore
		return numberToWords(n/100000) + " Lakh " + maybe(numberToWords(n%100000))
	}
	return numberToWords(n/10000000) + " Crore " + maybe(numberToWords(n%10000000))
}

func maybe(s string) string {
	s = strings.TrimSpace(s)
	if s == "Zero" {
		return ""
	}
	return s
}

func ifEmpty(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func maxF(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func must(err error) {
	if err != nil {
		// gopdf operations on a valid doc shouldn't fail; panic surfaces bugs early
		// in tests + the recover in the gin handler turns them into 500s.
		panic(err)
	}
}

// loadFonts registers the embedded Geist Regular TTF under two semantic names
// so the rest of the renderer code can call SetFont("regular-bold", …) for
// emphasis without us shipping a second weight. Geist Regular is a sans-serif
// that ships with the binary via go:embed — no external font file or system
// dependency needed in production.
func loadFonts(doc *gopdf.GoPdf) error {
	if len(geistRegular) == 0 {
		return fmt.Errorf("embedded font missing: did go:embed fail?")
	}
	if err := doc.AddTTFFontData("regular", geistRegular); err != nil {
		return fmt.Errorf("font load regular: %w", err)
	}
	if err := doc.AddTTFFontData("regular-bold", geistRegular); err != nil {
		return fmt.Errorf("font load regular-bold: %w", err)
	}
	return nil
}

// Time helper used by tests.
var nowFn = time.Now
