package proto

// Packet IDs for Minecraft Java Edition protocol 767 (1.21.1).
// Extracted from the protocol definition in docs/protocol/.

// Handshaking, serverbound.
const (
	CHandshake = 0x00
)

// Status.
const (
	SStatusResponse = 0x00
	SStatusPong     = 0x01
	CStatusRequest  = 0x00
	CStatusPing     = 0x01
)

// Login.
const (
	SLoginDisconnect    = 0x00
	SLoginEncryption    = 0x01
	SLoginSuccess       = 0x02
	SLoginSetCompress   = 0x03
	SLoginPluginRequest = 0x04
	CLoginStart         = 0x00
	CLoginEncryption    = 0x01
	CLoginPluginResp    = 0x02
	CLoginAcknowledged  = 0x03
)

// Configuration.
const (
	SCfgCookieRequest    = 0x00
	SCfgCustomPayload    = 0x01
	SCfgDisconnect       = 0x02
	SCfgFinish           = 0x03
	SCfgKeepAlive        = 0x04
	SCfgPing             = 0x05
	SCfgRegistryData     = 0x07
	SCfgTags             = 0x0D
	SCfgFeatureFlags     = 0x0C
	SCfgSelectKnownPacks = 0x0E
	CCfgClientSettings   = 0x00
	CCfgCookieResponse   = 0x01
	CCfgCustomPayload    = 0x02
	CCfgFinishAck        = 0x03
	CCfgKeepAlive        = 0x04
	CCfgPong             = 0x05
	CCfgSelectKnownPacks = 0x07
)

// Play, clientbound.
const (
	SBundleDelimiter    = 0x00
	SSpawnEntity        = 0x01
	SAnimation          = 0x03
	SAckBlockChange     = 0x05 // acknowledge_player_digging
	SBlockChange        = 0x09
	SDifficulty         = 0x0B
	SChunkBatchFinished = 0x0C
	SChunkBatchStart    = 0x0D
	SGameStateChange    = 0x22
	SKickDisconnect     = 0x1D
	SUnloadChunk        = 0x21
	SKeepAlive          = 0x26
	SChunkData          = 0x27
	SUpdateLight        = 0x2A
	SLogin              = 0x2B
	SRelEntityMove      = 0x2E
	SEntityMoveLook     = 0x2F
	SEntityLook         = 0x30
	SSetEquipment       = 0x5B
	SSetPassengers      = 0x5F
	SAbilities          = 0x38
	SPlayerRemove       = 0x3D
	SPlayerInfo         = 0x3E
	SPositionSync       = 0x40
	SEntityDestroy      = 0x42
	SRespawn            = 0x47
	SEntityHeadLook     = 0x48
	SHeldItemSlot       = 0x53
	SSetViewPosition    = 0x54
	SUpdateViewDistance = 0x55
	SSpawnPosition      = 0x56
	SEntityMetadata     = 0x58
	SUpdateTime         = 0x64
	SSystemChat         = 0x6C
	SEntityTeleport     = 0x70
	SDeclareCommands    = 0x11
	SCommandSuggestions = 0x10
	SContainerContent   = 0x13
	SCloseContainer     = 0x12
	SOpenScreen         = 0x33
	SContainerSlot      = 0x15
	SCustomPayload      = 0x19
)

// Game Event reasons (SGameStateChange), protocol 767.
const GameEventLevelChunksLoadStart byte = 13

// Play, serverbound.
const (
	CTeleportConfirm    = 0x00
	CChatCommand        = 0x04
	CChatCommandSigned  = 0x05
	CChatMessage        = 0x06
	CChunkBatchReceived = 0x08
	CClientCommand      = 0x09
	CClientSettings     = 0x0A
	CCommandSuggestions = 0x0B
	CCfgAcknowledged    = 0x0C
	CCloseContainer     = 0x0F
	CContainerClick     = 0x0E
	CCustomPayload      = 0x12
	CInteractEntity     = 0x16
	CKeepAlive          = 0x18
	CPlayerPosition     = 0x1A
	CPlayerPositionLook = 0x1B
	CPlayerLook         = 0x1C
	CPlayerFlying       = 0x1D
	CPlayerAbilities    = 0x23
	CBlockDig           = 0x24
	CEntityAction       = 0x25
	CHeldItemSlot       = 0x2F
	CSetCreativeSlot    = 0x32
	CArmAnimation       = 0x36
	CBlockPlace         = 0x38
	CUseItem            = 0x39
)
