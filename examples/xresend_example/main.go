package main

import (
	"context"
	"os"
	"time"

	"github.com/seefs001/xox/x"
	"github.com/seefs001/xox/xenv"
	"github.com/seefs001/xox/xlog"
	"github.com/seefs001/xox/xresend"
)

func main() {
	xenv.Load()
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		xlog.Error("RESEND_API_KEY environment variable is not set")
		return
	}

	objEmail := os.Getenv("EMAIL")
	if objEmail == "" {
		xlog.Error("EMAIL environment variable is not set")
		return
	}

	sendEmail := x.Ternary(os.Getenv("SEND_EMAIL") != "", os.Getenv("SEND_EMAIL"), "onboarding@resend.dev")

	client, err := xresend.NewClient(apiKey)
	if err != nil {
		xlog.Error("Failed to create Resend client", "error", err)
		return
	}

	ctx := context.Background()

	emailReq := xresend.SendEmailRequest{
		From:    sendEmail,
		To:      []string{objEmail},
		Subject: "Hello from Resend!",
		HTML:    "<strong>It works!</strong>",
	}

	emailResp, err := client.SendEmail(ctx, emailReq)
	if err != nil {
		xlog.Error("Failed to send email", "error", err)
		return
	}

	xlog.Info("Email sent successfully", "ID", emailResp.ID)

	// Wait for a short time to allow the email to be processed
	time.Sleep(5 * time.Second)

	email, err := client.GetEmail(ctx, emailResp.ID)
	if err != nil {
		xlog.Warn("Failed to get email details", "error", err)
	} else {
		xlog.Info("Email details",
			"From", email.From,
			"To", email.To,
			"Subject", email.Subject,
			"Created At", email.CreatedAt.Time.Format(time.RFC3339),
			"Status", email.LastEvent,
		)
	}

	domains, err := client.ListDomains(ctx)
	if err != nil {
		xlog.Warn("Failed to list domains", "error", err)
	} else {
		xlog.Info("Domains:")
		for _, domain := range domains.Data {
			xlog.Info("Domain",
				"Name", domain.Name,
				"Status", domain.Status,
				"Created At", domain.CreatedAt.Time.Format(time.RFC3339),
			)
		}
	}
}
