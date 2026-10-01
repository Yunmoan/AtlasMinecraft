package nbt

import (
	"bytes"
	"testing"
)

func TestJavaModifiedUTF8KnownBytes(t *testing.T) {
	// DataOutput.writeUTF("A\u0000🚀") writes A, encoded NUL and D83D DE80.
	want := []byte{String, 0, 9, 'A', 0xc0, 0x80, 0xed, 0xa0, 0xbd, 0xed, 0xba, 0x80}
	encoded, err := MarshalNetwork(Tag{Type: String, Value: "A\x00🚀"})
	if err != nil || !bytes.Equal(encoded, want) {
		t.Fatalf("Java UTF bytes mismatch: %x %v", encoded, err)
	}
	decoded, used, err := ParseNetwork(want)
	text, _ := decoded.Text()
	if err != nil || used != len(want) || text != "A\x00🚀" {
		t.Fatalf("Java UTF decode mismatch: %q %v", text, err)
	}
	for _, malformed := range [][]byte{{String, 0, 1, 0xc0}, {String, 0, 3, 0xed, 0xa0, 0xbd}} {
		if _, _, err := ParseNetwork(malformed); err == nil {
			t.Fatal("malformed Java UTF accepted")
		}
	}
}
