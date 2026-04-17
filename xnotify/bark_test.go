package xnotify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseBarkTargetDefaultHost(t *testing.T) {
	baseURL, deviceKeys, err := parseBarkTarget("device123")
	if err != nil {
		t.Fatalf("parseBarkTarget err: %v", err)
	}
	if baseURL != barkDefaultBaseURL {
		t.Fatalf("baseURL mismatch, got %s", baseURL)
	}
	if len(deviceKeys) != 1 || deviceKeys[0] != "device123" {
		t.Fatalf("deviceKeys mismatch, got %#v", deviceKeys)
	}
}

func TestParseBarkTargetCustomHost(t *testing.T) {
	baseURL, deviceKeys, err := parseBarkTarget("bark.example.com/device123")
	if err != nil {
		t.Fatalf("parseBarkTarget err: %v", err)
	}
	if baseURL != "https://bark.example.com" {
		t.Fatalf("baseURL mismatch, got %s", baseURL)
	}
	if len(deviceKeys) != 1 || deviceKeys[0] != "device123" {
		t.Fatalf("deviceKeys mismatch, got %#v", deviceKeys)
	}
}

func TestParseBarkTargetExplicitURL(t *testing.T) {
	baseURL, deviceKeys, err := parseBarkTarget("http://127.0.0.1:8080/bark/device123")
	if err != nil {
		t.Fatalf("parseBarkTarget err: %v", err)
	}
	if baseURL != "http://127.0.0.1:8080/bark" {
		t.Fatalf("baseURL mismatch, got %s", baseURL)
	}
	if len(deviceKeys) != 1 || deviceKeys[0] != "device123" {
		t.Fatalf("deviceKeys mismatch, got %#v", deviceKeys)
	}
}

func TestParseBarkTargetMultipleKeys(t *testing.T) {
	baseURL, deviceKeys, err := parseBarkTarget("bark.example.com/device123,device456")
	if err != nil {
		t.Fatalf("parseBarkTarget err: %v", err)
	}
	if baseURL != "https://bark.example.com" {
		t.Fatalf("baseURL mismatch, got %s", baseURL)
	}
	if len(deviceKeys) != 2 || deviceKeys[0] != "device123" || deviceKeys[1] != "device456" {
		t.Fatalf("deviceKeys mismatch, got %#v", deviceKeys)
	}
}

func TestBuildBarkPayload(t *testing.T) {
	badge := 3
	payload := buildBarkPayload("device123", "hello", NotifyOptions{
		Title:      "title",
		Subtitle:   "subtitle",
		URL:        "https://example.com",
		Group:      "ops",
		Sound:      "alarm",
		Icon:       "https://example.com/icon.png",
		Level:      "timeSensitive",
		Volume:     "8",
		Copy:       "copy-me",
		Action:     "none",
		Ciphertext: "cipher",
		Badge:      &badge,
		Call:       true,
		AutoCopy:   true,
		IsArchive:  true,
	})

	if payload.DeviceKey != "device123" || payload.Body != "hello" {
		t.Fatalf("payload base fields mismatch: %#v", payload)
	}
	if payload.Title != "title" || payload.Subtitle != "subtitle" {
		t.Fatalf("payload title fields mismatch: %#v", payload)
	}
	if payload.URL != "https://example.com" || payload.Group != "ops" || payload.Sound != "alarm" {
		t.Fatalf("payload routing fields mismatch: %#v", payload)
	}
	if payload.Icon != "https://example.com/icon.png" || payload.Level != "timeSensitive" || payload.Volume != "8" {
		t.Fatalf("payload alert fields mismatch: %#v", payload)
	}
	if payload.Badge == nil || *payload.Badge != 3 {
		t.Fatalf("payload badge mismatch: %#v", payload.Badge)
	}
	if payload.Call != "1" || payload.AutoCopy != "1" || payload.IsArchive != "1" {
		t.Fatalf("payload flags mismatch: %#v", payload)
	}
	if payload.Copy != "copy-me" || payload.Action != "none" || payload.Ciphertext != "cipher" {
		t.Fatalf("payload extra fields mismatch: %#v", payload)
	}
}

func TestSendBarkWithOptions(t *testing.T) {
	var gotPath string
	var gotPayload barkPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Fatalf("decode request body err: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200,"message":"success"}`))
	}))
	defer server.Close()

	err := SendBarkWithOptions(context.Background(), server.URL+"/device123", "hello bark",
		NotifyOptions{
			Title:    "title",
			Group:    "ops",
			Sound:    "alarm",
			AutoCopy: true,
		},
	)
	if err != nil {
		t.Fatalf("SendBarkWithOptions err: %v", err)
	}

	if gotPath != "/push" {
		t.Fatalf("request path mismatch, got %s", gotPath)
	}
	if gotPayload.DeviceKey != "device123" {
		t.Fatalf("device key mismatch, got %s", gotPayload.DeviceKey)
	}
	if gotPayload.Body != "hello bark" || gotPayload.Title != "title" {
		t.Fatalf("message fields mismatch, got %#v", gotPayload)
	}
	if gotPayload.Group != "ops" || gotPayload.Sound != "alarm" || gotPayload.AutoCopy != "1" {
		t.Fatalf("optional fields mismatch, got %#v", gotPayload)
	}
}

func TestSendBarkWithOptionsMultipleKeys(t *testing.T) {
	var gotKeys []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload barkPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request body err: %v", err)
		}
		gotKeys = append(gotKeys, payload.DeviceKey)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200,"message":"success"}`))
	}))
	defer server.Close()

	err := SendBarkWithOptions(context.Background(), server.URL+"/device123,device456", "hello bark", NotifyOptions{})
	if err != nil {
		t.Fatalf("SendBarkWithOptions err: %v", err)
	}

	if len(gotKeys) != 2 || gotKeys[0] != "device123" || gotKeys[1] != "device456" {
		t.Fatalf("request count or keys mismatch, got %#v", gotKeys)
	}
}

func TestSendBarkWithOptionsMultipleKeysPartialFailure(t *testing.T) {
	var gotKeys []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload barkPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request body err: %v", err)
		}
		gotKeys = append(gotKeys, payload.DeviceKey)
		w.Header().Set("Content-Type", "application/json")
		if payload.DeviceKey == "device456" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"message":"bad key"}`))
			return
		}
		_, _ = w.Write([]byte(`{"code":200,"message":"success"}`))
	}))
	defer server.Close()

	err := SendBarkWithOptions(context.Background(), server.URL+"/device123,device456", "hello bark", NotifyOptions{})
	if err == nil {
		t.Fatalf("expected partial failure error")
	}
	if len(gotKeys) != 2 {
		t.Fatalf("request count mismatch, got %#v", gotKeys)
	}
	if !strings.Contains(err.Error(), "device456") || !strings.Contains(err.Error(), "bad key") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNotifyWithOptionsRejectsUnsupportedBarkMessageType(t *testing.T) {
	err := NotifyWithOptions(
		context.Background(),
		"bark://device123",
		"hello bark",
		WithMessageType(MessageTypeMarkdown),
	)
	if err == nil {
		t.Fatalf("expected error for unsupported bark message type")
	}
}
