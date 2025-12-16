package xnotify

import (
	"context"
	"testing"
)

func TestParseBotIDLarkWithMention(t *testing.T) {
	target, err := parseBotID("lark://bot123@u1,u2")
	if err != nil {
		t.Fatalf("parseBotID err: %v", err)
	}
	if target.scheme != LarkScheme {
		t.Fatalf("scheme mismatch, got %s", target.scheme)
	}
	if target.botID != "bot123" {
		t.Fatalf("bot id mismatch, got %s", target.botID)
	}
	if len(target.mentions) != 2 || target.mentions[0] != "u1" || target.mentions[1] != "u2" {
		t.Fatalf("mentions mismatch, got %#v", target.mentions)
	}
}

func TestParseBotIDWeWork(t *testing.T) {
	target, err := parseBotID("wework://key123")
	if err != nil {
		t.Fatalf("parseBotID err: %v", err)
	}
	if target.scheme != WeWorkScheme {
		t.Fatalf("scheme mismatch, got %s", target.scheme)
	}
	if target.botID != "key123" {
		t.Fatalf("bot id mismatch, got %s", target.botID)
	}
	if len(target.mentions) != 0 {
		t.Fatalf("expected no mentions, got %#v", target.mentions)
	}
}

func TestParseBotIDWithQueryMention(t *testing.T) {
	Notify(context.Background(), "wework://9sxxx5@11600000032", "hello")

	Notify(context.Background(), "lark://2xxxxx97@刀刀", "hello")
}

func TestParseBotIDInvalid(t *testing.T) {
	if _, err := parseBotID("foo://id"); err == nil {
		t.Fatalf("expected error for invalid scheme")
	}
	if _, err := parseBotID("lark://"); err == nil {
		t.Fatalf("expected error for missing bot id")
	}
}

func TestNotifyUnsupportedProvider(t *testing.T) {
	err := Notify(context.Background(), "foo://id", "hello")
	if err == nil {
		t.Fatalf("expected error for unsupported provider")
	}
}

func TestNotifyRoutesToSender(t *testing.T) {
	fs := &fakeSender{}
	if err := RegisterProvider("fake", fs); err != nil {
		t.Fatalf("register provider err: %v", err)
	}

	err := Notify(context.Background(), "fake://abc@u1,u1,u2?mention=u2,u3", "hello")
	if err != nil {
		t.Fatalf("notify err: %v", err)
	}

	if fs.botID != "abc" {
		t.Fatalf("bot id mismatch, got %s", fs.botID)
	}
	if fs.message != "hello" {
		t.Fatalf("message mismatch, got %s", fs.message)
	}
	wantMentions := []string{"u1", "u2", "u3"}
	if len(fs.mentions) != len(wantMentions) {
		t.Fatalf("mentions length mismatch, got %#v", fs.mentions)
	}
	for i, m := range wantMentions {
		if fs.mentions[i] != m {
			t.Fatalf("mention mismatch at %d, got %s", i, fs.mentions[i])
		}
	}
}

type fakeSender struct {
	botID    string
	message  string
	mentions []string
}

func (f *fakeSender) Send(_ context.Context, botID string, message string, mentions []string) error {
	f.botID = botID
	f.message = message
	f.mentions = mentions
	return nil
}
