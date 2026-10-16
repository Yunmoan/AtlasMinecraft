package main

import (
	"atlas/internal/nbt"
	"atlas/internal/proto"
	"atlas/internal/world"
	"crypto/md5"
	"fmt"
	"math"
	"strings"
	"time"
)

type dimensionClient struct {
	fc               *proto.FramedConn
	dimension        string
	seed             int64
	crystals         map[[16]byte]dimensionCrystal
	mobs             map[[16]byte]dimensionMob
	crystalMetadata  map[int32]bool
	receivedChunks   map[[2]int32]bool
	entityOrderError string
}

type dimensionCrystal struct {
	ID       int32
	Position [3]float64
}
type dimensionMob struct {
	ID, Type int32
	Position [3]float64
}

func testDimensions() {
	fmt.Println("== dimensions / chunk isolation / reconnect ==")
	initial := "minecraft:overworld"
	if *dimensionResume {
		initial = "minecraft:the_end"
	}
	c := loginDimensionClient(initial)
	if c == nil {
		return
	}
	defer c.fc.Conn.Close()
	if !*dimensionResume && !c.teleport("minecraft:the_nether", [3]float64{.5, 201, .5}, "dirt", false) {
		return
	}
	markers := []struct{ dimension, block string }{
		{"minecraft:overworld", "stone"},
		{"minecraft:the_nether", "dirt"},
		{"minecraft:the_end", "gold_block"},
	}
	for _, marker := range markers {
		if !c.teleport(marker.dimension, [3]float64{.5, 201, .5}, marker.block, *dimensionResume) {
			return
		}
		if !*dimensionResume && !c.setMarker(marker.block) {
			return
		}
	}
	if *structureEntities {
		if !c.teleport("minecraft:overworld", [3]float64{-495.5, 90, 1904.5}, "stone", false) {
			return
		}
		if !c.waitForMobs(map[int32]string{118: "witch", 15: "cat"}) {
			return
		}
		if !c.teleport("minecraft:the_end", [3]float64{.5, 201, .5}, "gold_block", false) {
			return
		}
		if !c.teleport("minecraft:overworld", [3]float64{-1535.5, 90, 1280.5}, "stone", false) {
			return
		}
		if !c.waitForMobs(map[int32]string{113: "villager", 126: "zombie_villager"}) {
			return
		}
		if !c.teleport("minecraft:the_end", [3]float64{.5, 201, .5}, "gold_block", false) {
			return
		}
	}
	if *initialAnimals {
		for _, target := range []struct {
			kind     int32
			name     string
			position [3]float64
		}{
			{77, "pig", [3]float64{-2039, 73, -939}},
			{122, "wolf", [3]float64{-3071, 95, -1142}},
			{42, "fox", [3]float64{-3027, 96, 1058}},
			{43, "frog", [3]float64{-2922, 66, 2962}},
		} {
			if !c.teleport("minecraft:overworld", [3]float64{target.position[0] + .5, 110, target.position[2] + .5}, "stone", false) {
				return
			}
			if !c.waitForInitialAnimal(target.kind, target.name, target.position) {
				return
			}
			if !c.teleport("minecraft:the_end", [3]float64{.5, 201, .5}, "gold_block", false) {
				return
			}
		}
	}
	// Every visit uses the same chunk and block coordinates, so a stale
	// Overworld chunk or a shared dimension region fails the marker check.
	for _, marker := range markers {
		if !c.teleport(marker.dimension, [3]float64{.5, 201, .5}, marker.block, true) {
			return
		}
	}
	var crystalUUID [16]byte
	if *dimensionEntities {
		var ok bool
		crystalUUID, ok = c.waitForCrystal()
		if !ok {
			return
		}
	}
	c.command("tp 8.5 90 8.5", "Teleported TestBot to 8.50 90.00 8.50")
	checkCommand(c.fc, "save-all flush", "World saved.")
	c.fc.Conn.Close()
	time.Sleep(150 * time.Millisecond)
	rejoined := loginDimensionClient("minecraft:the_end")
	if rejoined == nil {
		return
	}
	defer rejoined.fc.Conn.Close()
	check("reconnected hashed seed", rejoined.seed == c.seed, fmt.Sprint(rejoined.seed))
	if *dimensionEntities {
		saved, ok := rejoined.waitForCrystal()
		check("reconnected End crystal UUID", ok && saved == crystalUUID, fmt.Sprintf("got%x want%x", saved, crystalUUID))
	}
}

func loginDimensionClient(expected string) *dimensionClient {
	fc := dial()
	failed := true
	defer func() {
		if failed {
			fc.Conn.Close()
		}
	}()
	handshake(fc, 2)
	uuid := md5.Sum([]byte("OfflinePlayer:TestBot"))
	uuid[6] = uuid[6]&15 | 0x30
	uuid[8] = uuid[8]&63 | 0x80
	fc.WritePacket(proto.CLoginStart, proto.NewWriter().String("TestBot").UUID(uuid).Payload())
	success := false
	for i := 0; i < 5 && !success; i++ {
		id, payload, err := readWithTimeout(fc, 10*time.Second)
		if err != nil {
			check("dimension login", false, err.Error())
			return nil
		}
		switch id {
		case proto.SLoginSetCompress:
			threshold, _ := proto.NewReader(payload).VarInt()
			fc.EnableCompression(threshold)
		case proto.SLoginSuccess:
			success = true
		case proto.SLoginDisconnect:
			check("dimension login", false, string(payload))
			return nil
		}
	}
	if !success {
		check("dimension login", false, "missing login success")
		return nil
	}
	fc.WritePacket(proto.CLoginAcknowledged, nil)
	finished, registryOK := false, false
	for i := 0; i < 30 && !finished; i++ {
		id, payload, err := readWithTimeout(fc, 15*time.Second)
		if err != nil {
			check("dimension configuration", false, err.Error())
			return nil
		}
		switch id {
		case proto.SCfgKeepAlive:
			fc.WritePacket(proto.CCfgKeepAlive, payload)
		case proto.SCfgRegistryData:
			r := proto.NewReader(payload)
			name, _ := r.String(0)
			if name != "minecraft:dimension_type" {
				continue
			}
			count, _ := r.VarInt()
			want := []string{"minecraft:overworld", "minecraft:overworld_caves", "minecraft:the_end", "minecraft:the_nether"}
			registryOK = count == 4
			for i := int32(0); i < count; i++ {
				entry, err := r.String(0)
				present, _ := r.Bool()
				registryOK = registryOK && err == nil && present && int(i) < len(want) && entry == want[i]
				if present {
					if err := skipDimensionNBT(r); err != nil {
						registryOK = false
						break
					}
				}
			}
			registryOK = registryOK && r.Remaining() == 0
		case proto.SCfgFinish:
			finished = true
		}
	}
	check("dimension registry order 0/2/3", registryOK, "")
	if !finished || !registryOK {
		return nil
	}
	fc.WritePacket(proto.CCfgClientSettings, clientSettingsPayload(2))
	fc.WritePacket(proto.CCfgFinishAck, nil)
	c := &dimensionClient{fc: fc}
	c.crystals = make(map[[16]byte]dimensionCrystal)
	c.mobs = make(map[[16]byte]dimensionMob)
	c.crystalMetadata = make(map[int32]bool)
	c.receivedChunks = make(map[[2]int32]bool)
	sawLogin, sawPosition, sawChunk, sawBatch := false, false, false, false
	var position [3]float64
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) && !(sawLogin && sawPosition && sawChunk && sawBatch) {
		id, payload, err := readWithTimeout(fc, 15*time.Second)
		if err != nil {
			check("dimension initial play", false, err.Error())
			return nil
		}
		switch id {
		case proto.SLogin:
			r := proto.NewReader(payload)
			r.I32()
			r.Bool()
			count, _ := r.VarInt()
			worlds := map[string]bool{}
			for i := int32(0); i < count; i++ {
				name, _ := r.String(0)
				worlds[name] = true
			}
			r.VarInt()
			r.VarInt()
			r.VarInt()
			r.Bool()
			r.Bool()
			r.Bool()
			name, seed, err := readDimensionSpawnInfo(r, false)
			sawLogin = err == nil && name == expected && count == 3 && worlds["minecraft:overworld"] && worlds["minecraft:the_nether"] && worlds["minecraft:the_end"]
			check("play login dimension "+expected, sawLogin, fmt.Sprintf("dimension=%s error=%v", name, err))
			if !sawLogin {
				return nil
			}
			c.dimension, c.seed = name, seed
		case proto.SPositionSync:
			position, err = c.confirmPosition(payload)
			sawPosition = err == nil
		case proto.SChunkData:
			if !sawLogin || !sawPosition {
				check("login chunk order", false, "chunk preceded login/position")
				return nil
			}
			x, z, _, err := readDimensionChunk(payload, c.dimension)
			if err != nil {
				check("initial dimension chunk", false, err.Error())
				return nil
			}
			if x == int32(math.Floor(position[0]/16)) && z == int32(math.Floor(position[2]/16)) {
				sawChunk = true
			}
			c.receivedChunks[[2]int32{x, z}] = true
		case proto.SChunkBatchFinished:
			sawBatch = true
			c.control(id, payload)
		default:
			c.control(id, payload)
		}
	}
	check("initial dimension chunk and batch", sawChunk && sawBatch, "")
	if expected == "minecraft:the_end" {
		check("saved End player position", position == [3]float64{8.5, 90, 8.5}, fmt.Sprint(position))
	}
	if !sawChunk || !sawPosition || !sawBatch {
		return nil
	}
	failed = false
	return c
}

func readDimensionSpawnInfo(r *proto.Reader, respawn bool) (string, int64, error) {
	typeID, err := r.VarInt()
	if err != nil {
		return "", 0, err
	}
	name, err := r.String(256)
	if err != nil {
		return "", 0, err
	}
	seed, err := r.I64()
	if err != nil {
		return "", 0, err
	}
	game, _ := r.I8()
	previous, _ := r.I8()
	debug, _ := r.Bool()
	r.Bool()
	death, _ := r.Bool()
	cooldown, err := r.VarInt()
	want := map[string]int32{"minecraft:overworld": 0, "minecraft:the_end": 2, "minecraft:the_nether": 3}
	id, ok := want[name]
	if err != nil || !ok || typeID != id || game != 1 || debug || death || cooldown != 0 {
		return name, seed, fmt.Errorf("invalid CommonPlayerSpawnInfo")
	}
	if respawn {
		flags, err := r.U8()
		if err != nil || flags != 3 || previous != 1 || r.Remaining() != 0 {
			return name, seed, fmt.Errorf("invalid Respawn flags or length")
		}
	}
	return name, seed, nil
}

func (c *dimensionClient) control(id int32, payload []byte) {
	switch id {
	case proto.SKeepAlive:
		c.fc.WritePacket(proto.CKeepAlive, payload)
	case proto.SChunkBatchFinished:
		c.fc.WritePacket(proto.CChunkBatchReceived, proto.NewWriter().F32(20).Payload())
	case proto.SPositionSync:
		c.confirmPosition(payload)
	case proto.SSpawnEntity:
		r := proto.NewReader(payload)
		eid, _ := r.VarInt()
		uuid, _ := r.UUID()
		kind, _ := r.VarInt()
		var position [3]float64
		for i := range position {
			position[i], _ = r.F64()
		}
		if kind == 30 {
			cx, cz := int32(math.Floor(position[0]/16)), int32(math.Floor(position[2]/16))
			if c.dimension != "minecraft:the_end" || !c.receivedChunks[[2]int32{cx, cz}] {
				c.entityOrderError = fmt.Sprintf("crystal preceded visible End chunk %d,%d", cx, cz)
			}
			c.crystals[uuid] = dimensionCrystal{eid, position}
		} else if kind == 118 || kind == 15 || kind == 113 || kind == 126 || kind == 87 || kind == 77 || kind == 19 || kind == 22 || kind == 53 || kind == 25 || kind == 2 || kind == 42 || kind == 43 || kind == 49 || kind == 65 || kind == 70 || kind == 74 || kind == 75 || kind == 81 || kind == 84 || kind == 111 || kind == 122 {
			cx, cz := int32(math.Floor(position[0]/16)), int32(math.Floor(position[2]/16))
			if c.dimension != "minecraft:overworld" || !c.receivedChunks[[2]int32{cx, cz}] {
				c.entityOrderError = fmt.Sprintf("mob preceded visible Overworld chunk %d,%d", cx, cz)
			}
			c.mobs[uuid] = dimensionMob{eid, kind, position}
		}
	case proto.SEntityMetadata:
		eid, err := proto.NewReader(payload).VarInt()
		if err == nil {
			c.crystalMetadata[eid] = true
		}
	case proto.SEntityDestroy:
		r := proto.NewReader(payload)
		count, _ := r.VarInt()
		for i := int32(0); i < count; i++ {
			eid, _ := r.VarInt()
			for uuid, e := range c.crystals {
				if e.ID == eid {
					delete(c.crystals, uuid)
					delete(c.crystalMetadata, eid)
				}
			}
		}
	case proto.SChunkData:
		r := proto.NewReader(payload)
		x, _ := r.I32()
		z, err := r.I32()
		if err == nil {
			c.receivedChunks[[2]int32{x, z}] = true
		}
	}
}
func (c *dimensionClient) confirmPosition(payload []byte) ([3]float64, error) {
	r := proto.NewReader(payload)
	var p [3]float64
	for i := range p {
		p[i], _ = r.F64()
	}
	r.F32()
	r.F32()
	flags, _ := r.U8()
	tid, err := r.VarInt()
	if err != nil || flags != 0 || r.Remaining() != 0 {
		return p, fmt.Errorf("invalid absolute position packet")
	}
	err = c.fc.WritePacket(proto.CTeleportConfirm, proto.NewWriter().VarInt(tid).Payload())
	return p, err
}

func (c *dimensionClient) teleport(dimension string, position [3]float64, block string, verify bool) bool {
	command := fmt.Sprintf("execute in %s run tp @s %g %g %g", dimension, position[0], position[1], position[2])
	c.fc.WritePacket(proto.CChatCommand, proto.NewWriter().String(command).Payload())
	respawn, positionOK, chunkOK, batchOK, chatOK := false, false, false, false, false
	wantBlock, _ := world.BlockStateID("minecraft:"+block, nil)
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) && !(respawn && positionOK && chunkOK && batchOK && chatOK) {
		id, payload, err := readWithTimeout(c.fc, 15*time.Second)
		if err != nil {
			check("teleport "+dimension, false, err.Error())
			return false
		}
		switch id {
		case proto.SRespawn:
			name, seed, err := readDimensionSpawnInfo(proto.NewReader(payload), true)
			if err != nil || name != dimension || seed != c.seed {
				check("respawn "+dimension, false, fmt.Sprintf("name=%s seed=%d error=%v", name, seed, err))
				return false
			}
			respawn = true
			c.dimension = name
			c.crystals = make(map[[16]byte]dimensionCrystal)
			c.mobs = make(map[[16]byte]dimensionMob)
			c.crystalMetadata = make(map[int32]bool)
			c.receivedChunks = make(map[[2]int32]bool)
		case proto.SPositionSync:
			p, err := c.confirmPosition(payload)
			if respawn {
				positionOK = err == nil && p == position
			}
		case proto.SChunkData:
			// Old queued packets may precede Respawn; after it, the epoch
			// barrier must prevent chunks of the old section height arriving.
			if !respawn {
				c.control(id, payload)
				continue
			}
			x, z, state, err := readDimensionChunk(payload, dimension)
			if err != nil {
				check("chunk after respawn "+dimension, false, err.Error())
				return false
			}
			c.receivedChunks[[2]int32{x, z}] = true
			if x == int32(math.Floor(position[0]/16)) && z == int32(math.Floor(position[2]/16)) {
				chunkOK = positionOK && (!verify || state == wantBlock)
				if verify && state != wantBlock {
					check("dimension block isolation "+dimension, false, fmt.Sprintf("state=%d want=%d", state, wantBlock))
					return false
				}
			}
		case proto.SChunkBatchFinished:
			if respawn {
				batchOK = true
			}
			c.control(id, payload)
		case proto.SSystemChat:
			chatOK = chatOK || strings.Contains(string(payload), "Teleported TestBot to")
		default:
			c.control(id, payload)
		}
	}
	ok := respawn && positionOK && chunkOK && batchOK && chatOK
	check("respawn, position, chunk, batch "+dimension, ok, fmt.Sprintf("respawn=%v position=%v chunk=%v batch=%v chat=%v", respawn, positionOK, chunkOK, batchOK, chatOK))
	return ok
}

func (c *dimensionClient) waitForMobs(want map[int32]string) bool {
	seen := map[int32]bool{}
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		for uuid, mob := range c.mobs {
			name, ok := want[mob.Type]
			if !ok || seen[mob.Type] || !c.crystalMetadata[mob.ID] {
				continue
			}
			check("generated "+name+" after visible chunk", c.entityOrderError == "", c.entityOrderError)
			fmt.Printf("  ENTITY %s uuid=%x position=%g,%g,%g\n", name, uuid, mob.Position[0], mob.Position[1], mob.Position[2])
			seen[mob.Type] = true
		}
		if len(seen) == len(want) {
			return true
		}
		id, payload, err := readWithTimeout(c.fc, 5*time.Second)
		if err != nil {
			check("generated structure entities", false, err.Error())
			return false
		}
		c.control(id, payload)
	}
	check("generated structure entities", false, fmt.Sprintf("seen=%v want=%v", seen, want))
	return false
}

func (c *dimensionClient) waitForInitialAnimal(kind int32, name string, position [3]float64) bool {
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		for uuid, mob := range c.mobs {
			if mob.Type == kind && mob.Position == position && c.crystalMetadata[mob.ID] {
				check("generated initial "+name+" after visible chunk", c.entityOrderError == "", c.entityOrderError)
				fmt.Printf("  ENTITY %s uuid=%x position=%g,%g,%g\n", name, uuid, mob.Position[0], mob.Position[1], mob.Position[2])
				return c.entityOrderError == ""
			}
		}
		id, payload, err := readWithTimeout(c.fc, 5*time.Second)
		if err != nil {
			check("initial animal packet", false, err.Error())
			return false
		}
		c.control(id, payload)
	}
	check("initial animal packet", false, "no generated animal at expected column")
	return false
}

func (c *dimensionClient) command(command, want string) bool {
	c.fc.WritePacket(proto.CChatCommand, proto.NewWriter().String(command).Payload())
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		id, payload, err := readWithTimeout(c.fc, 5*time.Second)
		if err != nil {
			check("/"+command, false, err.Error())
			return false
		}
		if id == proto.SSystemChat && strings.Contains(string(payload), want) {
			check("/"+command, true, "")
			return true
		}
		c.control(id, payload)
	}
	check("/"+command, false, "no command response")
	return false
}

func (c *dimensionClient) waitForCrystal() ([16]byte, bool) {
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		for uuid, crystal := range c.crystals {
			if crystal.Position[0] == 42.5 && crystal.Position[2] == .5 && c.crystalMetadata[crystal.ID] {
				check("generated End crystal after visible chunk", c.dimension == "minecraft:the_end" && crystal.Position[1] >= 76 && crystal.Position[1] <= 105 && c.entityOrderError == "", fmt.Sprintf("position=%v order=%s", crystal.Position, c.entityOrderError))
				fmt.Printf("  ENTITY crystal uuid=%x position=%g,%g,%g\n", uuid, crystal.Position[0], crystal.Position[1], crystal.Position[2])
				return uuid, true
			}
		}
		id, payload, err := readWithTimeout(c.fc, 5*time.Second)
		if err != nil {
			check("generated End crystal packet", false, err.Error())
			return [16]byte{}, false
		}
		c.control(id, payload)
	}
	check("generated End crystal packet", false, "no crystal at primary spike")
	return [16]byte{}, false
}

func (c *dimensionClient) setMarker(block string) bool {
	c.fc.WritePacket(proto.CChatCommand, proto.NewWriter().String("setblock 0 200 0 "+block).Payload())
	want, _ := world.BlockStateID("minecraft:"+block, nil)
	updated, chat := false, false
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) && !(updated && chat) {
		id, payload, err := readWithTimeout(c.fc, 5*time.Second)
		if err != nil {
			break
		}
		switch id {
		case proto.SBlockChange:
			r := proto.NewReader(payload)
			x, y, z, _ := r.Position()
			state, _ := r.VarInt()
			updated = updated || (x == 0 && y == 200 && z == 0 && uint32(state) == want)
		case proto.SSystemChat:
			chat = chat || strings.Contains(string(payload), "Set block at 0 200 0")
		default:
			c.control(id, payload)
		}
	}
	check("dimension marker "+c.dimension, updated && chat, block)
	return updated && chat
}

func skipDimensionNBT(r *proto.Reader) error {
	_, used, err := nbt.ParseNetwork(r.Rest())
	if err != nil {
		return err
	}
	_, err = r.Bytes(used)
	return err
}

// Decode the complete chunk packet, including all sections and light masks.
// Return the marker block at 0,200,0 for the identical-coordinate check.
func readDimensionChunk(payload []byte, dimension string) (int32, int32, uint32, error) {
	r := proto.NewReader(payload)
	x, _ := r.I32()
	z, err := r.I32()
	if err != nil {
		return x, z, 0, err
	}
	if err = skipDimensionNBT(r); err != nil {
		return x, z, 0, err
	}
	length, err := r.VarInt()
	if err != nil || length < 0 {
		return x, z, 0, fmt.Errorf("bad section byte length")
	}
	data, err := r.Bytes(int(length))
	if err != nil {
		return x, z, 0, err
	}
	s := proto.NewReader(data)
	sections, minY := 16, 0
	if dimension == "minecraft:overworld" {
		sections, minY = 24, -64
	}
	var marker uint32
	for section := 0; section < sections; section++ {
		if _, err = s.I16(); err != nil {
			return x, z, 0, err
		}
		value, err := readDimensionContainer(s, 9, 4096, int((200&15)<<8))
		if err != nil {
			return x, z, 0, err
		}
		if section == (200-minY)/16 {
			marker = value
		}
		if _, err = readDimensionContainer(s, 4, 64, 0); err != nil {
			return x, z, 0, err
		}
	}
	if s.Remaining() != 0 {
		return x, z, 0, fmt.Errorf("wrong section count for %s", dimension)
	}
	count, err := r.VarInt()
	if err != nil || count < 0 {
		return x, z, 0, fmt.Errorf("bad block entity count")
	}
	for i := int32(0); i < count; i++ {
		r.U8()
		r.I16()
		r.VarInt()
		if err = skipDimensionNBT(r); err != nil {
			return x, z, 0, err
		}
	}
	var skyMask uint64
	for mask := 0; mask < 4; mask++ {
		count, err := r.VarInt()
		if err != nil || count < 0 || count > 1 {
			return x, z, 0, fmt.Errorf("invalid light bitset")
		}
		for i := int32(0); i < count; i++ {
			value, err := r.I64()
			if err != nil {
				return x, z, 0, err
			}
			if uint64(value)>>uint(sections+2) != 0 {
				return x, z, 0, fmt.Errorf("light mask outside dimension")
			}
			if mask == 0 {
				skyMask = uint64(value)
			}
		}
	}
	if dimension != "minecraft:overworld" && skyMask != 0 {
		return x, z, 0, fmt.Errorf("sky light in %s", dimension)
	}
	for kind := 0; kind < 2; kind++ {
		count, err := r.VarInt()
		if err != nil || count < 0 || count > int32(sections+2) {
			return x, z, 0, fmt.Errorf("invalid light array count")
		}
		if kind == 0 && dimension != "minecraft:overworld" && count != 0 {
			return x, z, 0, fmt.Errorf("sky light array in %s", dimension)
		}
		for i := int32(0); i < count; i++ {
			length, err := r.VarInt()
			if err != nil || length != 2048 {
				return x, z, 0, fmt.Errorf("invalid light array length")
			}
			if _, err = r.Bytes(int(length)); err != nil {
				return x, z, 0, err
			}
		}
	}
	if r.Remaining() != 0 {
		return x, z, 0, fmt.Errorf("extra chunk packet bytes")
	}
	return x, z, marker, nil
}

func readDimensionContainer(r *proto.Reader, limit uint8, size, index int) (uint32, error) {
	width, err := r.U8()
	if err != nil || width > 32 {
		return 0, fmt.Errorf("invalid palette width")
	}
	var palette []uint32
	if width == 0 {
		value, err := r.VarInt()
		if err != nil {
			return 0, err
		}
		palette = []uint32{uint32(value)}
	} else if width < limit {
		count, err := r.VarInt()
		if err != nil || count <= 0 || count > int32(size) {
			return 0, fmt.Errorf("invalid palette count")
		}
		palette = make([]uint32, count)
		for i := range palette {
			value, err := r.VarInt()
			if err != nil {
				return 0, err
			}
			palette[i] = uint32(value)
		}
	}
	count, err := r.VarInt()
	if err != nil || count < 0 {
		return 0, fmt.Errorf("invalid packed array")
	}
	if width == 0 {
		if count != 0 {
			return 0, fmt.Errorf("nonempty single-value array")
		}
		return palette[0], nil
	}
	perLong := 64 / int(width)
	expected := (size + perLong - 1) / perLong
	if int(count) != expected {
		return 0, fmt.Errorf("packed array length %d want %d", count, expected)
	}
	var value uint32
	for i := int32(0); i < count; i++ {
		packed, err := r.I64()
		if err != nil {
			return 0, err
		}
		if int(i) == index/perLong {
			value = uint32(uint64(packed) >> uint((index%perLong)*int(width)) & ((1 << width) - 1))
		}
	}
	if len(palette) > 0 {
		if int(value) >= len(palette) {
			return 0, fmt.Errorf("palette index out of bounds")
		}
		value = palette[value]
	}
	return value, nil
}
