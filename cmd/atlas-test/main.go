// Command atlas-test is an end-to-end smoke test: it drives the full
// protocol flow (status ping -> login -> configuration -> play) against a
// running Atlas server and reports PASS/FAIL per stage.
package main

import (
	"atlas/internal/item"
	"atlas/internal/proto"
	"crypto/md5"
	"flag"
	"fmt"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var addr = flag.String("addr", "127.0.0.1:25565", "server address to test")
var operator = flag.Bool("operator", false, "also test operator commands (start server with -ops TestBot)")
var expectWeather = flag.String("expect-weather", "", "expected initial weather: clear, rain, or thunder")
var expectDay = flag.Int64("expect-day", -1, "expected saved daytime; allows up to 1200 ticks of runtime")
var expectPosition = flag.String("expect-position", "", "expected first position as x,y,z")
var clientView = flag.Int("client-view", 8, "client view distance sent during configuration")
var expectView = flag.Int("expect-view", 0, "expected play login view distance (0 skips check)")
var expectSimulation = flag.Int("expect-simulation", 0, "expected play login simulation distance (0 skips check)")
var changeView = flag.Int("change-view", 0, "change client view distance in play and check the update packet")
var digY = flag.Int("dig-y", -61, "ground Y for the block-dig check (default flat-world surface)")
var expectSavedItems = flag.Bool("expect-saved-items", false, "verify the inventory sample written by a previous operator smoke run")
var dimensionSmoke = flag.Bool("dimension-smoke", false, "test real three-dimension login, teleport, chunk isolation, and reconnect")
var dimensionResume = flag.Bool("dimension-resume", false, "verify dimension state and edits saved by a previous dimension smoke run")
var dimensionEntities = flag.Bool("dimension-entities", false, "also verify generated End crystal packets and reconnect identity")
var structureEntities = flag.Bool("structure-entities", false, "also verify seed 1234 swamp-hut and igloo entity packets")
var initialAnimals = flag.Bool("initial-animals", false, "also verify seed 1234 original chunk-generation animal packets")
var armorStandSmoke = flag.Bool("armor-stand-smoke", false, "test two TCP clients exchanging fixture armor-stand equipment, saving and reconnecting")
var armorStandResume = flag.Bool("armor-stand-resume", false, "verify armor-stand equipment and player inventory after server restart")
var chestSmoke = flag.Bool("chest-smoke", false, "test native single chest original loot and authoritative transfer (requires -ops TestBot)")
var chestResume = flag.Bool("chest-resume", false, "verify the chest and player inventory persisted by a previous chest smoke")
var chestPosition = flag.String("chest-position", "", "native chest block position as x,y,z")
var chestLootTable = flag.String("chest-loot-table", "", "original generated chest LootTable resource name")
var chestLootSeed = flag.Int64("chest-loot-seed", 0, "original generated chest nonzero LootTableSeed")
var chestState = flag.String("chest-state", "chest-smoke-state.json", "path to persist or verify the chest smoke expectation")

var (
	failures int
	checks   int
)

func check(name string, ok bool, detail string) {
	checks++
	if ok {
		fmt.Printf("  PASS  %s\n", name)
	} else {
		failures++
		fmt.Printf("  FAIL  %s  (%s)\n", name, detail)
	}
}

func dial() *proto.FramedConn {
	c, err := net.Dial("tcp", *addr)
	if err != nil {
		fmt.Println("cannot connect:", err)
		os.Exit(2)
	}
	return proto.NewFramedConn(c)
}

func handshake(fc *proto.FramedConn, nextState int32) {
	w := proto.NewWriter().VarInt(767).String("127.0.0.1").U16(25565).VarInt(nextState)
	if err := fc.WritePacket(proto.CHandshake, w.Payload()); err != nil {
		panic(err)
	}
}

func readWithTimeout(fc *proto.FramedConn, d time.Duration) (int32, []byte, error) {
	// A timed-out goroutine must not remain attached to the socket and
	// consume a later command's reply concurrently with the next reader.
	if err := fc.Conn.SetReadDeadline(time.Now().Add(d)); err != nil {
		return -1, nil, err
	}
	defer fc.Conn.SetReadDeadline(time.Time{})
	return fc.ReadPacket()
}

func main() {
	flag.Parse()
	testStatus()
	if *chestSmoke || *chestResume {
		testChestTCP()
	} else if *armorStandSmoke || *armorStandResume {
		testArmorStandTCP()
	} else if *dimensionSmoke {
		testDimensions()
	} else {
		testLoginAndPlay()
	}
	fmt.Printf("\n%d/%d checks passed\n", checks-failures, checks)
	if failures > 0 {
		os.Exit(1)
	}
}

func testStatus() {
	fmt.Println("== status ping ==")
	fc := dial()
	handshake(fc, 1)
	if err := fc.WritePacket(proto.CStatusRequest, nil); err != nil {
		panic(err)
	}
	id, payload, err := readWithTimeout(fc, 5*time.Second)
	if err != nil {
		check("status response", false, err.Error())
		return
	}
	body, _ := proto.NewReader(payload).String(0)
	check("status response", id == proto.SStatusResponse && strings.Contains(body, `"protocol":767`), body)

	w := proto.NewWriter().I64(123456789)
	fc.WritePacket(proto.CStatusPing, w.Payload())
	id, payload, err = readWithTimeout(fc, 5*time.Second)
	if err != nil {
		check("status pong", false, err.Error())
		return
	}
	v, _ := proto.NewReader(payload).I64()
	check("status pong", id == proto.SStatusPong && v == 123456789, fmt.Sprint(v))
}

func testLoginAndPlay() {
	fmt.Println("== login / config / play ==")
	joinStarted := time.Now()
	fc := dial()
	handshake(fc, 2)

	name := "TestBot"
	offl := md5.Sum([]byte("OfflinePlayer:" + name))
	offl[6] = (offl[6] & 0x0F) | 0x30
	offl[8] = (offl[8] & 0x3F) | 0x80

	if err := fc.WritePacket(proto.CLoginStart,
		proto.NewWriter().String(name).UUID(offl).Payload()); err != nil {
		panic(err)
	}

	// login: compress (optional) then success
	var sawSuccess bool
	for i := 0; i < 5 && !sawSuccess; i++ {
		id, payload, err := readWithTimeout(fc, 10*time.Second)
		if err != nil {
			check("login success", false, err.Error())
			return
		}
		switch id {
		case proto.SLoginSetCompress:
			th, _ := proto.NewReader(payload).VarInt()
			fc.EnableCompression(th)
		case proto.SLoginSuccess:
			sawSuccess = true
		case proto.SLoginDisconnect:
			check("login success", false, "server disconnected login")
			return
		}
	}
	check("login success", sawSuccess, "")

	if err := fc.WritePacket(proto.CLoginAcknowledged, nil); err != nil {
		panic(err)
	}

	// configuration: expect registry data etc., reply to keepalives, ack finish
	sawFinish := false
	registryPackets := 0
	for i := 0; i < 20 && !sawFinish; i++ {
		id, payload, err := readWithTimeout(fc, 15*time.Second)
		if err != nil {
			check("configuration finish", false, err.Error())
			return
		}
		switch id {
		case proto.SCfgKeepAlive:
			v, _ := proto.NewReader(payload).I64()
			fc.WritePacket(proto.CCfgKeepAlive, proto.NewWriter().I64(v).Payload())
		case proto.SCfgRegistryData:
			r := proto.NewReader(payload)
			reg, err := r.String(0)
			if err != nil {
				check("registry data", false, err.Error())
				return
			}
			n, _ := r.VarInt()
			if registryPackets == 0 {
				// vanilla order: first registry is worldgen/biome
				ok := reg == "minecraft:worldgen/biome" && n == 64
				if ok {
					e, _ := r.String(0)
					present, _ := r.Bool()
					nbtTag, _ := r.U8()
					ok = e == "minecraft:badlands" && present && nbtTag == 0x0A
				}
				check("registry data (biome first, 64 entries, per-registry format)", ok, reg)
			}
			registryPackets++
		case proto.SCfgFinish:
			sawFinish = true
		}
	}
	check("configuration finish", sawFinish, "")
	check("registry data packets (11 registries)", registryPackets == 11, fmt.Sprint(registryPackets))
	if *clientView < 0 || *clientView > 255 {
		fmt.Fprintln(os.Stderr, "-client-view must be between 0 and 255")
		os.Exit(2)
	}
	if err := fc.WritePacket(proto.CCfgClientSettings,
		clientSettingsPayload(byte(*clientView))); err != nil {
		panic(err)
	}
	if err := fc.WritePacket(proto.CCfgFinishAck, nil); err != nil {
		panic(err)
	}

	// play: login packet, position sync -> teleport confirm, chunks stream
	var eid, chunks int32
	var completedChunks, batchChunks int32
	batchOpen := false
	batchError := ""
	var loginView, loginSimulation int32
	expectedChunks := int32(80)
	var firstPosition [3]float64
	var firstChunk [2]int32
	firstChunkSeen := false
	var initialInventory item.Inventory
	sawInventory := false
	var selectedSlot int8
	var firstChunkAt, loadStartAt time.Duration
	loadStarts := 0
	loadStartError := ""
	var initialDay int64 = -1
	initialWeather := "clear"
	var rain, thunder float32
	sawPlayLogin, sawPosition, sawChat, sawCommands := false, false, false, false
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		id, payload, err := readWithTimeout(fc, 10*time.Second)
		if err != nil {
			break
		}
		switch id {
		case proto.SContainerContent:
			initialInventory, sawInventory = readInventoryContent(payload)
		case proto.SHeldItemSlot:
			selectedSlot, _ = proto.NewReader(payload).I8()
		case proto.SLogin:
			r := proto.NewReader(payload)
			eid, _ = r.I32()
			r.Bool() // hardcore
			dimensionCount, _ := r.VarInt()
			for i := int32(0); i < dimensionCount; i++ {
				r.String(0)
			}
			r.VarInt() // max players
			loginView, _ = r.VarInt()
			expectedChunks = min(int32(80), (loginView*2+1)*(loginView*2+1))
			loginSimulation, _ = r.VarInt()
			sawPlayLogin = true
		case proto.SPositionSync:
			if !sawPosition {
				if loadStarts != 1 {
					loadStartError = "initial position arrived before level chunks load start"
				}
				// teleport id is the last varint
				r := proto.NewReader(payload)
				firstPosition[0], _ = r.F64()
				firstPosition[1], _ = r.F64()
				firstPosition[2], _ = r.F64()
				r.F32()
				r.F32()
				flags, _ := r.U8()
				_ = flags
				tid, _ := r.VarInt()
				fc.WritePacket(proto.CTeleportConfirm, proto.NewWriter().VarInt(tid).Payload())
				sawPosition = true
			}
		case proto.SChunkBatchStart:
			if batchOpen {
				batchError = "nested chunk batch"
			}
			batchOpen, batchChunks = true, 0
		case proto.SChunkData:
			if !batchOpen {
				batchError = "chunk outside a batch"
			}
			batchChunks++
			chunks++
			if !firstChunkSeen {
				firstChunkAt = time.Since(joinStarted)
				r := proto.NewReader(payload)
				firstChunk[0], _ = r.I32()
				firstChunk[1], _ = r.I32()
				firstChunkSeen = true
			}
		case proto.SChunkBatchFinished:
			count, countErr := proto.NewReader(payload).VarInt()
			if !batchOpen || countErr != nil || count != batchChunks {
				batchError = "chunk batch size mismatch"
			}
			completedChunks += batchChunks
			batchOpen = false
			fc.WritePacket(proto.CChunkBatchReceived, proto.NewWriter().F32(20).Payload())
		case proto.SKeepAlive:
			v, _ := proto.NewReader(payload).I64()
			fc.WritePacket(proto.CKeepAlive, proto.NewWriter().I64(v).Payload())
		case proto.SSystemChat:
			sawChat = true
		case proto.SUpdateTime:
			if initialDay < 0 {
				r := proto.NewReader(payload)
				r.I64()
				initialDay, _ = r.I64()
			}
		case proto.SGameStateChange:
			r := proto.NewReader(payload)
			reason, _ := r.U8()
			value, valueErr := r.F32()
			if reason == proto.GameEventLevelChunksLoadStart {
				loadStarts++
				loadStartAt = time.Since(joinStarted)
				if valueErr != nil || value != 0 || len(payload) != 5 || !sawPlayLogin || sawPosition || firstChunkSeen || loadStarts != 1 {
					loadStartError = "malformed, repeated or out-of-order level chunks load start"
				}
			}
			if reason == 7 {
				rain = value
			} else if reason == 8 {
				thunder = value
			}
			if rain > 0 {
				initialWeather = "rain"
			}
			if thunder > 0 {
				initialWeather = "thunder"
			}
		case proto.SDeclareCommands:
			n, _ := proto.NewReader(payload).VarInt()
			sawCommands = n > 1
		}
		if sawPlayLogin && sawPosition && sawChat && chunks >= expectedChunks && !batchOpen {
			break
		}
	}
	check("play login packet", sawPlayLogin && eid != 0, fmt.Sprintf("eid=%d", eid))
	check("player inventory content (46 slots)", sawInventory, "")
	if *expectSavedItems {
		checkSavedInventory(initialInventory, selectedSlot)
	}
	check("level chunks load start before position/chunks", loadStarts == 1 && loadStartError == "", fmt.Sprintf("events=%d %s", loadStarts, loadStartError))
	if *expectView > 0 {
		check("configured view distance", loginView == int32(*expectView), fmt.Sprintf("view=%d", loginView))
	}
	if *expectSimulation > 0 {
		check("configured simulation distance", loginSimulation == int32(*expectSimulation), fmt.Sprintf("simulation=%d", loginSimulation))
	}
	check("position sync", sawPosition, "")
	check("chunk stream", chunks >= expectedChunks, fmt.Sprintf("received %d of %d expected chunks", chunks, expectedChunks))
	check("complete chunk batches", batchError == "" && !batchOpen && completedChunks == chunks, batchError)
	check("system chat", sawChat, "")
	check("command tree", sawCommands, "root-only tree")
	fmt.Printf("  TCP join timing: load-start=%s first-chunk=%s received-%d-chunks=%s\n", loadStartAt, firstChunkAt, chunks, time.Since(joinStarted))
	if *expectPosition != "" {
		parts := strings.Split(*expectPosition, ",")
		match := len(parts) == 3
		var expected [3]float64
		if match {
			for i, part := range parts {
				var err error
				expected[i], err = strconv.ParseFloat(part, 64)
				match = match && err == nil && firstPosition[i] == expected[i]
			}
		}
		check("saved player position", match, fmt.Sprintf("position=%v", firstPosition))
		if match {
			cx, cz := int32(math.Floor(expected[0]/16)), int32(math.Floor(expected[2]/16))
			check("first chunk at player position", firstChunk == [2]int32{cx, cz}, fmt.Sprintf("first chunk=%v", firstChunk))
		}
	}
	if *expectWeather != "" {
		check("saved weather", initialWeather == *expectWeather, initialWeather)
	}
	if *expectDay >= 0 {
		delta := initialDay - *expectDay
		if delta < 0 {
			delta = -delta
		}
		check("saved time", initialDay >= *expectDay && delta <= 1200, fmt.Sprintf("daytime=%d", initialDay))
	}

	// chat echo: send a chat message, expect a system chat containing it
	fc.WritePacket(proto.CChatMessage, proto.NewWriter().String("hello from atlas-test").Payload())
	echoed := false
	deadline = time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) && !echoed {
		id, payload, err := readWithTimeout(fc, 5*time.Second)
		if err != nil {
			break
		}
		if id == proto.SSystemChat {
			// content is an NBT string component: 0x08 00 00 <len><text>
			if strings.Contains(string(payload), "hello from atlas-test") {
				echoed = true
			}
		} else if id == proto.SKeepAlive {
			v, _ := proto.NewReader(payload).I64()
			fc.WritePacket(proto.CKeepAlive, proto.NewWriter().I64(v).Payload())
		}
	}
	check("chat echo", echoed, "")

	// Operator checks prepare a nearby target. A remote bedrock dig is now
	// rejected by the server's interaction range validation.
	digX, digZ, digHeight := int32(0), int32(1), int32(*digY)
	if *operator {
		checkCommand(fc, "setblock 0 200 0 stone", "Set block at 0 200 0")
		checkCommand(fc, "setblock 0 200 1 stone", "Set block at 0 200 1")
		checkCommand(fc, "tp 0.5 201 0.5", "Teleported TestBot to 0.50 201.00 0.50")
		digHeight = 200
	}
	// block dig round-trip: expect block-change correction plus prediction ack
	fc.WritePacket(proto.CBlockDig, proto.NewWriter().
		VarInt(0). // start dig
		I64(proto.PackPosition(digX, digHeight, digZ)).
		I8(1).     // face top
		VarInt(7). // sequence
		Payload())
	dug, acked := false, false
	deadline = time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) && (!dug || !acked) {
		id, payload, err := readWithTimeout(fc, 5*time.Second)
		if err != nil {
			break
		}
		switch id {
		case proto.SBlockChange:
			r := proto.NewReader(payload)
			bx, by, bz, _ := r.Position()
			state, _ := r.VarInt()
			if bx == digX && bz == digZ && by == digHeight && state == 0 {
				dug = true
			}
		case proto.SAckBlockChange:
			sequence, _ := proto.NewReader(payload).VarInt()
			acked = sequence == 7
		case proto.SKeepAlive:
			v, _ := proto.NewReader(payload).I64()
			fc.WritePacket(proto.CKeepAlive, proto.NewWriter().I64(v).Payload())
		}
	}
	check("block dig -> block change", dug, "")
	check("block dig prediction ack", acked, "")
	checkCommand(fc, "help", "Commands:")
	checkCommand(fc, "where", "Position:")
	if *operator {
		checkCommand(fc, "tps", "TPS (1m/5m/15m):")
		checkPlayerSuggestions(fc)
		checkCommand(fc, "mspt", "MSPT work (last 1m")
		checkCommand(fc, "status", "Atlas status:")
		checkCommand(fc, "sysinfo", "World: cached chunks")
		checkCommand(fc, "minecraft:time query daytime", "Daytime:")
		checkCommand(fc, "time add 0.5s", "Added 10 ticks")
		checkCommand(fc, "help tp", "/tp [player]")
		checkCommand(fc, "minecraft:setblock 3 200 0 minecraft:oak_log[axis=x]", "Set block at 3 200 0")
		checkCommand(fc, "setblock 3 200 0 stone keep", "Block edit failed:")
		checkCommand(fc, "teleport ~ ~ ~", "Teleported TestBot to")
		checkCommand(fc, "save-all flush", "World saved.")
		checkMechanics(fc)
		// The clock can advance while the command saves generated regions.
		checkCommand(fc, "time set night", "Time set to ")
		checkCommand(fc, "weather rain", "Weather set to rain")
		checkCommand(fc, "setblock 0 74 0 stone", "Set block at 0 74 0")
		checkCommand(fc, "tp 34 85 -17", "Teleported TestBot to 34.50 85.00 -16.50")
	}
	if *changeView > 0 {
		fc.WritePacket(proto.CClientSettings,
			clientSettingsPayload(byte(*changeView)))
		gotView := int32(-1)
		unloaded := false
		deadline := time.Now().Add(10 * time.Second)
		for time.Now().Before(deadline) && (gotView < 0 || !unloaded) {
			id, payload, err := readWithTimeout(fc, 5*time.Second)
			if err != nil {
				break
			}
			switch id {
			case proto.SUpdateViewDistance:
				gotView, _ = proto.NewReader(payload).VarInt()
			case proto.SUnloadChunk:
				unloaded = true
			case proto.SKeepAlive:
				v, _ := proto.NewReader(payload).I64()
				fc.WritePacket(proto.CKeepAlive, proto.NewWriter().I64(v).Payload())
			}
		}
		check("live view distance update", gotView == int32(*changeView), fmt.Sprintf("view=%d", gotView))
		check("out-of-range chunks unloaded", unloaded, "no unload packet")
	}
	if *operator {
		seedInventorySmoke(fc)
	}

	fc.Conn.Close()
}

func clientSettingsPayload(view byte) []byte {
	return proto.NewWriter().String("en_us").U8(view).
		VarInt(0).Bool(true).U8(0x7f). // chat mode, chat colors, skin parts
		VarInt(1).Bool(false).Bool(true).VarInt(0).Payload()
}

func checkCommand(fc *proto.FramedConn, command, want string) {
	if err := fc.WritePacket(proto.CChatCommand, proto.NewWriter().String(command).Payload()); err != nil {
		check("/"+command, false, err.Error())
		return
	}
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		id, payload, err := readWithTimeout(fc, 5*time.Second)
		if err != nil {
			check("/"+command, false, err.Error())
			return
		}
		if id == proto.SKeepAlive {
			v, _ := proto.NewReader(payload).I64()
			fc.WritePacket(proto.CKeepAlive, proto.NewWriter().I64(v).Payload())
		} else if id == proto.SPositionSync {
			r := proto.NewReader(payload)
			for range 3 {
				r.F64()
			}
			r.F32()
			r.F32()
			r.U8()
			tid, err := r.VarInt()
			if err != nil {
				check("/"+command, false, "invalid teleport confirmation")
				return
			}
			fc.WritePacket(proto.CTeleportConfirm, proto.NewWriter().VarInt(tid).Payload())
		} else if id == proto.SSystemChat && strings.Contains(string(payload), want) {
			check("/"+command, true, "")
			return
		}
	}
	check("/"+command, false, "no matching response")
}

func checkPlayerSuggestions(fc *proto.FramedConn) {
	if err := fc.WritePacket(proto.CCommandSuggestions, proto.NewWriter().VarInt(42).String("/tp Tes").Payload()); err != nil {
		check("player name Tab completion", false, err.Error())
		return
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		id, payload, err := readWithTimeout(fc, 5*time.Second)
		if err != nil {
			check("player name Tab completion", false, err.Error())
			return
		}
		if id == proto.SKeepAlive {
			fc.WritePacket(proto.CKeepAlive, payload)
			continue
		}
		if id != proto.SCommandSuggestions {
			continue
		}
		r := proto.NewReader(payload)
		transaction, _ := r.VarInt()
		start, _ := r.VarInt()
		length, _ := r.VarInt()
		count, _ := r.VarInt()
		found := false
		for range count {
			value, _ := r.String(256)
			tooltip, _ := r.Bool()
			if tooltip {
				check("player name Tab completion", false, "unexpected tooltip")
				return
			}
			found = found || value == "TestBot"
		}
		check("player name Tab completion", transaction == 42 && start == 4 && length == 3 && found && r.Remaining() == 0, string(payload))
		return
	}
	check("player name Tab completion", false, "no suggestion response")
}
