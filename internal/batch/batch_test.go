package batch

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"

	"github.com/portal26/event_receiver_service/internal/config"
	"github.com/portal26/event_receiver_service/internal/model"
)

type mockS3Client struct {
	s3.Client
}

func (m *mockS3Client) PutObject(ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return &s3.PutObjectOutput{}, nil
}

func TestBatchProcessor_AddEvent(t *testing.T) {
	cfg := &config.Config{
		BatchSizeMB:         1,
		BatchTimeoutSeconds: 5,
	}
	bp := NewBatchProcessor(cfg)
	bp.client = &mockS3Client{}

	event := model.Event{
		EventTimestamp: time.Now(),
		Body:           "test event",
	}

	bp.AddEvent(event)
	time.Sleep(100 * time.Millisecond)

	select {
	case e := <-bp.events:
		assert.Equal(t, event, e)
	default:
		t.Fatal("Event not added to channel")
	}
}
