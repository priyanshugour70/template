package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/your-org/your-service/internal/modules/billing/pricing"
	apperr "github.com/your-org/your-service/internal/pkg/errors"
	"github.com/your-org/your-service/internal/pkg/mail"
)

// CycleReport summarises one daily tick — what changed, what failed. The
// worker logs this so we can verify the cron is doing its job.
type CycleReport struct {
	TrialsExpired   int   `json:"trialsExpired"`
	InvoicesIssued  int   `json:"invoicesIssued"`
	Errors          []string `json:"errors,omitempty"`
}

// ExpireTrialsBefore flips every subscription whose trial_ends_at is past `t`
// from status='trial' to status='expired'. Idempotent — running it twice with
// the same `t` is a no-op on the second call.
//
// We don't email the customer here; the worker emits a notification via the
// billing.cycle.tick consumer once the count > 0.
func (s *Service) ExpireTrialsBefore(ctx context.Context, t time.Time) (int, error) {
	res := s.repo.DB().WithContext(ctx).
		Model(&Subscription{}).
		Where("status = 'trial' AND trial_ends_at IS NOT NULL AND trial_ends_at <= ?", t).
		Updates(map[string]interface{}{
			"status":   "expired",
			"ended_at": t,
		})
	if res.Error != nil {
		return 0, apperr.New(apperr.CodeInternal, "expire trials failed", res.Error)
	}
	if res.RowsAffected > 0 {
		s.log.Info("billing cycle: trials expired",
			zap.Int64("rows", res.RowsAffected),
			zap.Time("cutoff", t))
	}
	return int(res.RowsAffected), nil
}

// RunBillingCycle is the daily worker entry point. It first expires due
// trials, then rolls every active subscription forward whose
// next_billing_at <= t — mints the next-cycle invoice, advances period
// pointers, and emits an EmailInvoiceIssued.
//
// Each subscription is processed in its own transaction so a failure on one
// org doesn't block the others. The report carries error strings, not whole
// errors — the worker logs the report and continues.
func (s *Service) RunBillingCycle(ctx context.Context, t time.Time) (*CycleReport, error) {
	rep := &CycleReport{}

	// 1. Trial expiry.
	n, err := s.ExpireTrialsBefore(ctx, t)
	if err != nil {
		rep.Errors = append(rep.Errors, err.Error())
	}
	rep.TrialsExpired = n

	// 2. Find every subscription that needs a new invoice now.
	due := []Subscription{}
	if err := s.repo.DB().WithContext(ctx).
		Where(`status = 'active'
		       AND next_billing_at IS NOT NULL
		       AND next_billing_at <= ?
		       AND (cancel_at IS NULL OR cancel_at > ?)`,
			t, t).
		Order("next_billing_at ASC").
		Find(&due).Error; err != nil {
		rep.Errors = append(rep.Errors, "load due subscriptions: "+err.Error())
		return rep, nil
	}

	for i := range due {
		sub := due[i]
		if err := s.rollSubscriptionForward(ctx, &sub, t); err != nil {
			s.log.Warn("billing cycle: subscription roll failed",
				zap.String("subscription", sub.ID.String()),
				zap.String("plan", sub.PlanCode),
				zap.Error(err))
			rep.Errors = append(rep.Errors,
				fmt.Sprintf("sub %s: %v", sub.ID, err))
			continue
		}
		rep.InvoicesIssued++
	}

	s.log.Info("billing cycle done",
		zap.Int("trials_expired", rep.TrialsExpired),
		zap.Int("invoices_issued", rep.InvoicesIssued),
		zap.Int("errors", len(rep.Errors)))
	return rep, nil
}

// rollSubscriptionForward re-prices the subscription against the current
// catalog, mints the next-cycle invoice, advances period pointers, and
// notifies. Wrapped in one transaction per subscription.
func (s *Service) rollSubscriptionForward(ctx context.Context, sub *Subscription, t time.Time) error {
	// Re-price using the current catalog so a feature whose price changed
	// upstream gets billed correctly next cycle. The user's selection comes
	// from the subscription's snapshotted features list.
	var featureKeys []string
	if len(sub.Features) > 0 {
		_ = json.Unmarshal(sub.Features, &featureKeys)
	}
	if len(featureKeys) == 0 {
		// Snapshot is empty — treat as a one-off plan that shouldn't auto-roll.
		return fmt.Errorf("subscription has no feature snapshot")
	}

	// User count from limits.users.max (set at activation time).
	userCount := 1
	if len(sub.Limits) > 0 {
		var m map[string]json.Number
		if err := json.Unmarshal(sub.Limits, &m); err == nil {
			if v, ok := m["users.max"]; ok {
				if n, err := v.Int64(); err == nil && n > 0 {
					userCount = int(n)
				}
			}
		}
	}

	tax, catalog, err := s.loadCatalogAndTax(ctx)
	if err != nil {
		return err
	}
	customerState := sub.BillingState
	if customerState == "" {
		customerState = tax.HomeState
	}
	quote, err := s.runPricing(ctx, catalog, tax, featureKeys, userCount, customerState)
	if err != nil {
		return err
	}

	// Mint the invoice + advance the subscription atomically.
	invoiceNum, err := s.repo.NextInvoiceNumber(ctx, sub.TenantID, t.Year())
	if err != nil {
		return err
	}

	newPeriodStart := t
	newPeriodEnd := t.Add(subscriptionCycleDays * 24 * time.Hour)
	due := t.Add(invoiceDueDays * 24 * time.Hour)
	taxCents := quote.CGSTCents + quote.SGSTCents + quote.IGSTCents

	var issuedInv Invoice

	err = s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		lineItemsJSON, _ := json.Marshal(quote.Lines)
		inv := Invoice{
			TenantID:        sub.TenantID,
			OrganizationID:  sub.OrganizationID,
			SubscriptionID:  &sub.ID,
			Number:          invoiceNum,
			Status:          "open",
			Currency:        quote.Currency,
			SubtotalCents:   quote.SubtotalCents,
			TaxCents:        taxCents,
			TotalCents:      quote.TotalCents,
			AmountDueCents:  quote.TotalCents,
			AmountPaidCents: 0,
			LineItems:       lineItemsJSON,
			PeriodStart:     &newPeriodStart,
			PeriodEnd:       &newPeriodEnd,
			IssuedAt:        t,
			DueAt:           &due,
			HSNSAC:          tax.DefaultHSNSAC,
			PlaceOfSupply:   customerState,
			CGSTCents:       quote.CGSTCents,
			SGSTCents:       quote.SGSTCents,
			IGSTCents:       quote.IGSTCents,
			Metadata:        []byte(`{"source":"cycle"}`),
		}
		if err := tx.Create(&inv).Error; err != nil {
			return fmt.Errorf("create invoice: %w", err)
		}
		lineRows := make([]InvoiceLine, 0, len(quote.Lines))
		for _, l := range quote.Lines {
			lineRows = append(lineRows, InvoiceLine{
				InvoiceID:          inv.ID,
				FeatureKey:         l.FeatureKey,
				Description:        l.Description,
				HSNSAC:             l.HSNSAC,
				Quantity:           l.Quantity,
				UnitPriceCents:     l.UnitPriceCents,
				TaxableAmountCents: l.TaxableAmountCents,
				CGSTCents:          l.Tax.CGSTCents,
				SGSTCents:          l.Tax.SGSTCents,
				IGSTCents:          l.Tax.IGSTCents,
				TotalCents:         l.TotalCents,
				SortOrder:          l.SortOrder,
				Metadata:           []byte("{}"),
			})
		}
		if len(lineRows) > 0 {
			if err := tx.Create(&lineRows).Error; err != nil {
				return fmt.Errorf("create invoice lines: %w", err)
			}
		}

		// Advance the subscription's billing cursor.
		patch := map[string]interface{}{
			"current_period_start": newPeriodStart,
			"current_period_end":   newPeriodEnd,
			"next_billing_at":      newPeriodEnd,
			"last_billed_at":       t,
			"unit_price_cents":     quote.SubtotalCents,
			"tax_cents":            taxCents,
			"total_cents":          quote.TotalCents,
		}
		if err := tx.Model(&Subscription{}).Where("id = ?", sub.ID).Updates(patch).Error; err != nil {
			return fmt.Errorf("advance subscription: %w", err)
		}

		issuedInv = inv
		return nil
	})
	if err != nil {
		return err
	}

	// Bust the org's resolved feature cache so any limit changes go live.
	s.invalidateCache(ctx, sub.OrganizationID)

	// Best-effort: render the invoice PDF + email it. Same pattern as
	// RecordPayment — failures here don't roll back the invoice.
	pdfBytes, pdfErr := s.renderInvoicePDF(ctx, &issuedInv)
	if pdfErr != nil {
		s.log.Warn("cycle invoice pdf render failed",
			zap.String("invoice", issuedInv.Number), zap.Error(pdfErr))
	}
	if pdfErr == nil && s.s3 != nil {
		key := s.s3.Key("invoices", issuedInv.TenantID.String(), issuedInv.Number+".pdf")
		if upErr := s.s3.PutBytes(ctx, key, "application/pdf", pdfBytes); upErr == nil {
			_ = s.repo.UpdateInvoice(ctx, issuedInv.ID, map[string]interface{}{"pdf_storage_key": key})
		}
	}

	recipient := s.resolveBillingEmail(ctx, &issuedInv, sub)
	if recipient != "" {
		amtFmt := formatMoney(issuedInv.Currency, issuedInv.TotalCents)
		dueFmt := ""
		if issuedInv.DueAt != nil {
			dueFmt = issuedInv.DueAt.Format("02 Jan 2006")
		}
		body := mail.EmailInvoiceIssued(issuedInv.Number, dueFmt, amtFmt, "")
		var mailErr error
		if pdfErr == nil && len(pdfBytes) > 0 {
			mailErr = s.mailer.SendWithAttachments(recipient,
				"New invoice "+issuedInv.Number, body,
				mail.Attachment{
					Filename: "invoice-" + issuedInv.Number + ".pdf",
					MIMEType: "application/pdf",
					Data:     pdfBytes,
				})
		} else {
			mailErr = s.mailer.Send(recipient, "New invoice "+issuedInv.Number, body)
		}
		if mailErr != nil {
			s.log.Warn("cycle invoice email send failed",
				zap.String("to", recipient), zap.Error(mailErr))
		}
	}
	return nil
}

// pricingHasUsableCatalog is a tiny shim that exists so unit tests can run
// against an in-memory catalog without touching the DB. Kept here for the
// rare case where the cycle runs while the catalog is empty.
func pricingHasUsableCatalog(catalog map[string]pricing.Feature) bool {
	return len(catalog) > 0
}

// UUIDFromInvoice is a small adapter for tests that need to spin a fake
// invoice without going through the full activation path.
func UUIDFromInvoice(inv *Invoice) uuid.UUID { return inv.ID }
