package background_test

import (
	"context"
	"testing"

	"github.com/alextanhongpin/go-queue/background"
)

func TestBackgroundMock(t *testing.T) {
	bg := background.NewMock()

	type deliverEmailRequest struct {
		Email string
	}

	deliverEmail := func(ctx context.Context, req deliverEmailRequest) error {
		if exp, got := "john.doe@appleseed.com", req.Email; exp != got {
			t.Logf("expected %s, got %s", exp, got)
		}
		t.Logf("deliverEmail: %+v\n", req)

		return nil
	}

	if err := background.RegisterTask(bg, deliverEmail); err != nil {
		t.Fatalf("failed to register background task: %v", err)
	}

	if err := bg.Enqueue(deliverEmailRequest{
		Email: "john.doe@appleseed.com",
	}); err != nil {
		t.Fatalf("failed to deliver email: %v", err)
	}
}
