package model

import "time"

type BillingAccount struct {
	ClerkUserID          string
	PlanCode             string
	SubscriptionStatus   string
	BillingInterval      string
	StripeCustomerID     string
	StripeSubscriptionID string
	CurrentPeriodEnd     *time.Time
	CancelAtPeriodEnd    bool
	FreeDownloadsLimit   int
	FreeDownloadsUsed    int
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (b BillingAccount) IsProActive(now time.Time) bool {
	if b.PlanCode != "pro" {
		return false
	}

	switch b.SubscriptionStatus {
	case "active", "trialing":
		if b.CurrentPeriodEnd == nil {
			return true
		}
		return now.Before(*b.CurrentPeriodEnd) || now.Equal(*b.CurrentPeriodEnd)
	default:
		return false
	}
}

func (b BillingAccount) FreeDownloadsRemaining() int {
	remaining := b.FreeDownloadsLimit - b.FreeDownloadsUsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (b BillingAccount) ToSummary(now time.Time) BillingSummary {
	isPro := b.IsProActive(now)
	remaining := b.FreeDownloadsRemaining()

	return BillingSummary{
		Plan:                   planOrDefault(b.PlanCode),
		SubscriptionStatus:     subscriptionStatusOrDefault(b.SubscriptionStatus),
		FreeDownloadsLimit:     b.FreeDownloadsLimit,
		FreeDownloadsUsed:      b.FreeDownloadsUsed,
		FreeDownloadsRemaining: remaining,
		CanDownload:            isPro || remaining > 0,
	}
}

func planOrDefault(plan string) string {
	if plan == "" {
		return "free"
	}
	return plan
}

func subscriptionStatusOrDefault(status string) string {
	if status == "" {
		return "inactive"
	}
	return status
}
