// Command atlas-capture connects to a Minecraft server, walks the full
// protocol (status -> login -> configuration -> play) and dumps every
// clientbound packet to a binary log for byte-level comparison against a
// reference implementation. Responds to keepalives and known-packs so the
// peer behaves normally.
package main

import (
	"atlas/internal/proto"
	"crypto/md5"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

var (
	addr   = flag.String("addr", "127.0.0.1:25599", "server address")
	outDir = flag.String("out", "capture", "output directory")
	name   = flag.String("name", "CaptureBot", "bot username")
)

func main() {
	flag.Parse()
	os.MkdirAll(*outDir, 0o755)

	c, err := net.Dial("tcp", *addr)
	if err != nil {
		fmt.Println("dial:", err)
		os.Exit(2)
	}
	fc := proto.NewFramedConn(c)
	seq := 0
	dump := func(tag string, id int32, payload []byte) {
		f := filepath.Join(*outDir, fmt.Sprintf("%04d-%s-pkt%02x.bin", seq, tag, id))
		os.WriteFile(f, payload, 0o644)
		seq++
	}

	// handshake + login
	handshake(fc, 2)
	offl := md5.Sum([]byte("OfflinePlayer:" + *name))
	offl[6] = (offl[6] & 0x0F) | 0x30
	offl[8] = (offl[8] & 0x3F) | 0x80
	must(fc.WritePacket(proto.CLoginStart, proto.NewWriter().String(*name).UUID(offl).Payload()))

	deadline := time.Now().Add(45 * time.Second)
	phase := "login"
	for time.Now().Before(deadline) {
		c.SetReadDeadline(time.Now().Add(20 * time.Second))
		id, payload, err := fc.ReadPacket()
		if err != nil {
			fmt.Println("read:", err)
			return
		}
		switch phase {
		case "login":
			switch id {
			case proto.SLoginSetCompress:
				th, _ := proto.NewReader(payload).VarInt()
				fc.EnableCompression(th)
			case proto.SLoginSuccess:
				must(fc.WritePacket(proto.CLoginAcknowledged, nil))
				phase = "config"
			case proto.SLoginDisconnect:
				fmt.Println("login disconnect:", string(payload))
				return
			}
			dump("login", id, payload)
		case "config":
			dump("config", id, payload)
			switch id {
			case proto.SCfgKeepAlive:
				v, _ := proto.NewReader(payload).I64()
				must(fc.WritePacket(proto.CCfgKeepAlive, proto.NewWriter().I64(v).Payload()))
			case proto.SCfgPing:
				v, _ := proto.NewReader(payload).I64()
				must(fc.WritePacket(proto.CCfgPong, proto.NewWriter().I64(v).Payload()))
			case proto.SCfgSelectKnownPacks:
				must(fc.WritePacket(proto.CCfgSelectKnownPacks, proto.NewWriter().VarInt(0).Payload()))
			case proto.SCfgFinish:
				must(fc.WritePacket(proto.CCfgFinishAck, nil))
				phase = "play"
			}
		case "play":
			dump("play", id, payload)
			switch id {
			case proto.SKeepAlive:
				v, _ := proto.NewReader(payload).I64()
				must(fc.WritePacket(proto.CKeepAlive, proto.NewWriter().I64(v).Payload()))
			case proto.SPositionSync:
				r := proto.NewReader(payload)
				r.F64(); r.F64(); r.F64(); r.F32(); r.F32()
				r.U8()
				tid, _ := r.VarInt()
				must(fc.WritePacket(proto.CTeleportConfirm, proto.NewWriter().VarInt(tid).Payload()))
			case proto.SChunkBatchStart:
				// vanilla client sends batch received after the batch finishes
			case proto.SChunkBatchFinished:
				must(fc.WritePacket(proto.CChunkBatchReceived, proto.NewWriter().F32(40).Payload()))
			}
		}
	}
	fmt.Println("captured", seq, "packets")
}

func handshake(fc *proto.FramedConn, nextState int32) {
	w := proto.NewWriter().VarInt(767).String("127.0.0.1").U16(25599).VarInt(nextState)
	must(fc.WritePacket(proto.CHandshake, w.Payload()))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
