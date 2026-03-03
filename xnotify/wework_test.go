package xnotify

import "testing"

func TestBuildWeComTextPayload(t *testing.T) {
	payload := buildWeComTextPayload("hello", []string{"13800001111", "@all"})

	if payload["msgtype"] != wecomMsgTypeText {
		t.Fatalf("msgtype mismatch, got %v", payload["msgtype"])
	}

	text, ok := payload["text"].(map[string]any)
	if !ok {
		t.Fatalf("text payload type mismatch, got %T", payload["text"])
	}
	if text["content"] != "hello" {
		t.Fatalf("content mismatch, got %v", text["content"])
	}

	mobiles, ok := text["mentioned_mobile_list"].([]string)
	if !ok {
		t.Fatalf("mentioned_mobile_list type mismatch, got %T", text["mentioned_mobile_list"])
	}
	if len(mobiles) != 2 || mobiles[0] != "13800001111" || mobiles[1] != "@all" {
		t.Fatalf("mentioned_mobile_list mismatch, got %#v", mobiles)
	}
}

func TestBuildWeComTextPayloadWithoutMentions(t *testing.T) {
	payload := buildWeComTextPayload("hello", nil)
	text := payload["text"].(map[string]any)
	if _, ok := text["mentioned_mobile_list"]; ok {
		t.Fatalf("unexpected mentioned_mobile_list in payload: %#v", text)
	}
}

func TestBuildWeComMarkdownPayload(t *testing.T) {
	payload := buildWeComMarkdownPayload("**bold**")
	if payload["msgtype"] != wecomMsgTypeMarkdown {
		t.Fatalf("msgtype mismatch, got %v", payload["msgtype"])
	}

	markdown, ok := payload["markdown"].(map[string]any)
	if !ok {
		t.Fatalf("markdown payload type mismatch, got %T", payload["markdown"])
	}
	if markdown["content"] != "**bold**" {
		t.Fatalf("markdown content mismatch, got %v", markdown["content"])
	}
}

func TestBuildWeComMarkdownV2Payload(t *testing.T) {
	payload := buildWeComMarkdownV2Payload("# title")
	if payload["msgtype"] != wecomMsgTypeMarkdownV2 {
		t.Fatalf("msgtype mismatch, got %v", payload["msgtype"])
	}

	markdownV2, ok := payload["markdown_v2"].(map[string]any)
	if !ok {
		t.Fatalf("markdown_v2 payload type mismatch, got %T", payload["markdown_v2"])
	}
	if markdownV2["content"] != "# title" {
		t.Fatalf("markdown_v2 content mismatch, got %v", markdownV2["content"])
	}
}
