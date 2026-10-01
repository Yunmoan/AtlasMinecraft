package server

import (
	"atlas/internal/player"
	"os"
	"path/filepath"
	"testing"
)

// Opt-in fixtures for independent decoding with Java 1.21.1 packet codecs.
func TestExportVanillaCompatibilityFixtures(t *testing.T) {
	dir := os.Getenv("ATLAS_COMPAT_FIXTURES")
	if dir == "" {
		t.Skip("set ATLAS_COMPAT_FIXTURES to export reference packets")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	p := player.New("CodecProbe", offlineUUID("CodecProbe"), 7, nil)
	p.SetPos(.5, -60, 4.5, 180, 0, true)
	for name, data := range map[string][]byte{
		"spawn.bin": spawnEntityPayload(p), "commands.bin": declareCommandsPayloadForLevel(4),
		"chat.bin": systemChat("§a绿色 §b§lAtlas§r §#12ABCD彩色🚀"),
		"tps.bin":  commandChat("TPS (1m/5m/15m): 20.00 / 17.00 / 10.00 (target: 20; observed: 1m0s)"),
		"join.bin": playerPresenceChat("CodecProbe", true), "left.bin": playerPresenceChat("CodecProbe", false),
	} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
}
