package server

import (
	"atlas/internal/nbt"
	"strings"
	"testing"
)

func plainComponent(tag nbt.Tag) string {
	if text, ok := tag.Text(); ok {
		return text
	}
	if o, ok := tag.Object(); ok {
		text, _ := o.Text("text")
		if extra, ok := o["extra"].Elements(); ok {
			for _, child := range extra {
				text += plainComponent(child)
			}
		}
		return text
	}
	return ""
}

func chatPlainText(t *testing.T, payload []byte) string {
	t.Helper()
	tag, consumed, err := nbt.ParseNetwork(payload)
	if err != nil || len(payload) != consumed+1 || payload[consumed] != 0 {
		t.Fatalf("invalid system chat packet: %v", err)
	}
	return plainComponent(tag)
}

func TestColoredTextUsesVanillaComponentStylesAndResets(t *testing.T) {
	spans := legacySpans("§aGreen §lBold §r白色 §#A1b2C3RGB §x§1§2§3§4§5§6Hex")
	if len(spans) != 5 || spans[0].Color != "green" || !spans[1].Bold || spans[2].Bold || spans[2].Color != "white" || spans[3].Color != "#A1b2C3" || spans[4].Color != "#123456" {
		t.Fatalf("wrong legacy styles: %+v", spans)
	}
	if got := chatPlainText(t, systemChat("§aGreen §lBold §r白色 §#A1b2C3RGB §x§1§2§3§4§5§6Hex")); got != "Green Bold 白色 RGB Hex" {
		t.Fatal(got)
	}
	response := "TPS (1m/5m/15m): 20.00 / 17.00 / 10.00 (target: 20; observed: 1m0s)"
	packet := commandChat(response)
	if chatPlainText(t, packet) != response {
		t.Fatal("formatting changed command text")
	}
	root, _, err := nbt.ParseNetwork(packet)
	if err != nil {
		t.Fatal(err)
	}
	o, _ := root.Object()
	extra, _ := o["extra"].Elements()
	for i, color := range map[int]string{1: "green", 3: "yellow", 5: "red"} {
		part, _ := extra[i].Object()
		got, _ := part.Text("color")
		if got != color {
			t.Fatalf("TPS %d color %s", i, got)
		}
	}
	for _, reply := range []string{"Usage: /time", "You do not have permission to use this command.", "World save failed: disk full"} {
		root, _, _ := nbt.ParseNetwork(commandChat(reply))
		object, _ := root.Object()
		parts, _ := object["extra"].Elements()
		part, _ := parts[0].Object()
		color, _ := part.Text("color")
		if color != "red" {
			t.Fatal("command error was not red")
		}
	}
	for _, text := range []string{"§", "§zunknown", "§#bad", "你好🚀"} {
		if got := chatPlainText(t, systemChat(text)); !strings.Contains(got, strings.TrimPrefix(text, "§")) {
			t.Fatal("text lost")
		}
	}
}
