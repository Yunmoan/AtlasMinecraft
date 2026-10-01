package server

import (
	"atlas/internal/nbt"
	"strconv"
	"strings"
	"unicode/utf8"
)

type textSpan struct {
	Text, Color, Command                                string
	Bold, Italic, Underlined, Strikethrough, Obfuscated bool
}

func (s textSpan) tag() nbt.Tag {
	o := nbt.Object{"text": {Type: nbt.String, Value: s.Text}}
	if s.Color != "" {
		o["color"] = nbt.Tag{Type: nbt.String, Value: s.Color}
	}
	for name, value := range map[string]bool{"bold": s.Bold, "italic": s.Italic, "underlined": s.Underlined, "strikethrough": s.Strikethrough, "obfuscated": s.Obfuscated} {
		b := int8(0)
		if value {
			b = 1
		}
		o[name] = nbt.Tag{Type: nbt.Byte, Value: b}
	}
	if s.Command != "" {
		o["clickEvent"] = nbt.Tag{Type: nbt.Compound, Value: nbt.Object{
			"action": {Type: nbt.String, Value: "suggest_command"},
			"value":  {Type: nbt.String, Value: s.Command},
		}}
	}
	return nbt.Tag{Type: nbt.Compound, Value: o}
}

func spansTag(spans []textSpan) nbt.Tag {
	extra := make([]nbt.Tag, 0, len(spans))
	for _, span := range spans {
		extra = append(extra, span.tag())
	}
	o := nbt.Object{"text": {Type: nbt.String, Value: ""}}
	if len(extra) > 0 {
		o["extra"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: extra}
	}
	return nbt.Tag{Type: nbt.Compound, Value: o}
}

func encodeText(tag nbt.Tag) []byte {
	payload, err := nbt.MarshalNetwork(tag)
	if err != nil {
		payload, _ = nbt.MarshalNetwork(nbt.Tag{Type: nbt.String, Value: "Invalid server message"})
	}
	return payload
}

func styledChat(spans ...textSpan) []byte { return append(encodeText(spansTag(spans)), 0) }

func playerPresenceChat(name string, joined bool) []byte {
	key := "multiplayer.player.left"
	if joined {
		key = "multiplayer.player.joined"
	}
	component := nbt.Tag{Type: nbt.Compound, Value: nbt.Object{
		"translate": {Type: nbt.String, Value: key},
		"color":     {Type: nbt.String, Value: "yellow"},
		"with":      {Type: nbt.List, ListType: nbt.Compound, Value: []nbt.Tag{textSpan{Text: name}.tag()}},
	}}
	return append(encodeText(component), 0)
}

func componentJSON(tag nbt.Tag) any {
	if o, ok := tag.Object(); ok {
		result := map[string]any{}
		for key, value := range o {
			result[key] = componentJSON(value)
			if value.Type == nbt.Byte {
				number, _ := value.Number()
				result[key] = number != 0
			}
		}
		return result
	}
	if list, ok := tag.Elements(); ok {
		result := make([]any, 0, len(list))
		for _, value := range list {
			result = append(result, componentJSON(value))
		}
		return result
	}
	return tag.Value
}

var legacyColors = []string{"black", "dark_blue", "dark_green", "dark_aqua", "dark_red", "dark_purple", "gold", "gray", "dark_gray", "blue", "green", "aqua", "red", "light_purple", "yellow", "white"}

// Vanilla literal NBT text does not interpret legacy section-sign formatting.
func legacySpans(text string) []textSpan {
	style := textSpan{Color: "white"}
	var spans []textSpan
	var buffer strings.Builder
	flush := func() {
		if buffer.Len() > 0 {
			part := style
			part.Text = buffer.String()
			spans = append(spans, part)
			buffer.Reset()
		}
	}
	for len(text) > 0 {
		if strings.HasPrefix(text, "§") && len(text) > len("§") {
			rest := text[len("§"):]
			code := strings.ToLower(rest[:1])
			if index := strings.Index("0123456789abcdef", code); index >= 0 {
				flush()
				style = textSpan{Color: legacyColors[index]}
				text = rest[1:]
				continue
			}
			if code == "#" && len(rest) >= 7 {
				if _, err := strconv.ParseUint(rest[1:7], 16, 24); err == nil {
					flush()
					style = textSpan{Color: rest[:7]}
					text = rest[7:]
					continue
				}
			}
			if code == "x" {
				remaining := rest[1:]
				var rgb strings.Builder
				for i := 0; i < 6 && strings.HasPrefix(remaining, "§") && len(remaining) >= len("§")+1; i++ {
					hex := remaining[len("§") : len("§")+1]
					if !strings.ContainsAny(strings.ToLower(hex), "0123456789abcdef") {
						break
					}
					rgb.WriteString(hex)
					remaining = remaining[len("§")+1:]
				}
				if rgb.Len() == 6 {
					flush()
					style = textSpan{Color: "#" + rgb.String()}
					text = remaining
					continue
				}
			}
			if strings.Contains("klmnor", code) {
				flush()
				switch code {
				case "k":
					style.Obfuscated = true
				case "l":
					style.Bold = true
				case "m":
					style.Strikethrough = true
				case "n":
					style.Underlined = true
				case "o":
					style.Italic = true
				case "r":
					style = textSpan{Color: "white"}
				}
				text = rest[1:]
				continue
			}
		}
		r, size := utf8.DecodeRuneInString(text)
		buffer.WriteRune(r)
		text = text[size:]
	}
	flush()
	return spans
}

func commandChat(response string) []byte {
	var spans []textSpan
	errorResponse := strings.HasPrefix(response, "Usage:") || strings.HasPrefix(response, "Unknown") ||
		strings.HasPrefix(response, "Invalid") || strings.Contains(response, "do not have permission") ||
		strings.Contains(response, "failed") || strings.Contains(response, "not found") ||
		strings.HasPrefix(response, "Teleport destination unavailable") || strings.HasPrefix(response, "In online mode,")
	for i, line := range strings.Split(response, "\n") {
		if i > 0 {
			spans = append(spans, textSpan{Text: "\n"})
		}
		color := "green"
		if errorResponse {
			color = "red"
		} else if strings.HasPrefix(response, "Commands:") {
			color = "aqua"
		} else if strings.Contains(line, ":") {
			color = "gray"
		}
		if strings.HasPrefix(line, "TPS (1m/5m/15m): ") && !strings.Contains(line, "collecting") {
			prefix := "TPS (1m/5m/15m): "
			values, suffix, _ := strings.Cut(strings.TrimPrefix(line, prefix), " (")
			spans = append(spans, textSpan{Text: prefix, Color: "aqua"})
			for j, value := range strings.Split(values, " / ") {
				if j > 0 {
					spans = append(spans, textSpan{Text: " / ", Color: "dark_gray"})
				}
				tps, _ := strconv.ParseFloat(value, 64)
				c := "green"
				if tps < 18 {
					c = "yellow"
				}
				if tps < 15 {
					c = "red"
				}
				spans = append(spans, textSpan{Text: value, Color: c})
			}
			spans = append(spans, textSpan{Text: " (" + suffix, Color: "dark_gray"})
			continue
		}
		span := textSpan{Text: line, Color: color}
		if strings.HasPrefix(line, "/") {
			span.Command = strings.Fields(line)[0] + " "
		}
		spans = append(spans, span)
	}
	return styledChat(spans...)
}
