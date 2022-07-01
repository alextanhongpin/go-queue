package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/alextanhongpin/go-queue/background"
	"github.com/hibiken/asynq"
)

func main() {
	bg := background.New()
	defer bg.Close()

	usecase := NewEmailUsecase(bg)

	// Register a task based on the request name.
	if err := background.RegisterTask(bg, DeliverEmailRequest{}, usecase.DeliverEmail); err != nil {
		log.Fatalf("failed to register task: %v", err)
	}

	ctx := context.Background()

	// This will send the email immediately.
	if err := usecase.DeliverEmail(ctx, DeliverEmailRequest{
		Email:    "john.appleseed@mail.com",
		Template: "send immediately",
	}); err != nil {
		log.Fatalf("failed to deliver email: %v", err)
	}

	// This will call the `DeliverEmail` in the background.
	// Both operations calls the usecase `DeliverEmail`.
	if err := usecase.DeliverEmailBackground(ctx, DeliverEmailRequest{
		Email:    "john.appleseed@mail.com",
		Template: "send in background",
	}); err != nil {
		log.Fatalf("failed to deliver email in background: %v", err)
	}

	if err := bg.Start(); err != nil {
		log.Fatalf("failed to start server: %s", err)
	}
}

type EmailUsecase struct {
	bg *background.Background
}

func NewEmailUsecase(bg *background.Background) *EmailUsecase {
	return &EmailUsecase{
		bg: bg,
	}
}

type DeliverEmailRequest struct {
	Email    string
	Template string
}

// DeliverEmail sends a template to the recipient.
func (uc *EmailUsecase) DeliverEmail(ctx context.Context, req DeliverEmailRequest) error {
	fmt.Printf("delivering email: %+v", req)

	return nil
}

// DeliverEmailBackground will call the `DeliverEmail` method in the background.
func (uc *EmailUsecase) DeliverEmailBackground(ctx context.Context, req DeliverEmailRequest) error {
	return uc.bg.Enqueue(req, asynq.ProcessIn(10*time.Second))
}
