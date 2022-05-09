package reTGanslatorBot

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type fakeUpdater struct {
	updates []tgbotapi.Update
}

func (f *fakeUpdater) HandleUpdate(update tgbotapi.Update) error {
	f.updates = append(f.updates, update)
	return nil
}

func TestServer_updateHandler_happy_path(t *testing.T) {
	fu := &fakeUpdater{}

	srv := NewServer(fu, "12345")

	u := tgbotapi.Update{
		UpdateID: 1,
		Message: &tgbotapi.Message{
			MessageID: 123,
		},
	}

	payload, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("unexpected error during marshaling update payload: %v\n", err)
		return
	}

	req := httptest.NewRequest(http.MethodPost, "/webhook/12345", bytes.NewReader(payload))
	req.Header.Add("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected error code 200, got=%d", rr.Code)
	}

	if got := rr.Body.String(); got != "" {
		t.Errorf("expected empty response body, got=%v", got)
	}

	if len(fu.updates) != 1 {
		t.Errorf("expected updates number = %d, got %d", 1, len(fu.updates))
	}
}
