### play/toClient/packet_login
["container", [{"name": "entityId", "type": "i32"}, {"name": "isHardcore", "type": "bool"}, {"name": "worldNames", "type": ["array", {"countType": "varint", "type": "string"}]}, {"name": "maxPlayers", "type": "varint"}, {"name": "viewDistance", "type": "varint"}, {"name": "simulationDistance", "type": "varint"}, {"name": "reducedDebugInfo", "type": "bool"}, {"name": "enableRespawnScreen", "type": "bool"}, {"name": "doLimitedCrafting", "type": "bool"}, {"name": "worldState", "type": "SpawnInfo"}, {"name": "enforcesSecureChat", "type": "bool"}]]
### play/toClient/packet_position
["container", [{"name": "x", "type": "f64"}, {"name": "y", "type": "f64"}, {"name": "z", "type": "f64"}, {"name": "yaw", "type": "f32"}, {"name": "pitch", "type": "f32"}, {"name": "flags", "type": "PositionUpdateRelatives"}, {"name": "teleportId", "type": "varint"}]]
### play/toClient/packet_map_chunk
["container", [{"name": "x", "type": "i32"}, {"name": "z", "type": "i32"}, {"name": "heightmaps", "type": "anonymousNbt"}, {"name": "chunkData", "type": ["buffer", {"countType": "varint"}]}, {"name": "blockEntities", "type": ["array", {"countType": "varint", "type": "chunkBlockEntity"}]}, {"name": "skyLightMask", "type": ["array", {"countType": "varint", "type": "i64"}]}, {"name": "blockLightMask", "type": ["array", {"countType": "varint", "type": "i64"}]}, {"name": "emptySkyLightMask", "type": ["array", {"countType": "varint", "type": "i64"}]}, {"name": "emptyBlockLightMask", "type": ["array", {"countType": "varint", "type": "i64"}]}, {"name": "skyLight", "type": ["array", {"countType": "varint", "type": ["array", {"countType": "varint", "type": "u8"}]}]}, {"name": "blockLight", "type": ["array", {"countType": "varint", "type": ["array", {"countType": "varint", "type": "u8"}]}]}]]
### play/toClient/packet_chunk_batch_finished
["container", [{"name": "batchSize", "type": "varint"}]]
### play/toClient/packet_block_change
["container", [{"name": "location", "type": "position"}, {"name": "type", "type": "varint"}]]
### play/toClient/packet_keep_alive
["container", [{"name": "keepAliveId", "type": "i64"}]]
### play/toClient/packet_system_chat
["container", [{"name": "content", "type": "anonymousNbt"}, {"name": "isActionBar", "type": "bool"}]]
### play/toClient/packet_spawn_entity
["container", [{"name": "entityId", "type": "varint"}, {"name": "objectUUID", "type": "UUID"}, {"name": "type", "type": "varint"}, {"name": "x", "type": "f64"}, {"name": "y", "type": "f64"}, {"name": "z", "type": "f64"}, {"name": "pitch", "type": "i8"}, {"name": "yaw", "type": "i8"}, {"name": "headPitch", "type": "i8"}, {"name": "objectData", "type": "varint"}, {"name": "velocity", "type": "vec3i16"}]]
### play/toClient/packet_spawn_position
["container", [{"name": "location", "type": "position"}, {"name": "angle", "type": "f32"}]]
### play/toClient/packet_abilities
["container", [{"name": "flags", "type": "i8"}, {"name": "flyingSpeed", "type": "f32"}, {"name": "walkingSpeed", "type": "f32"}]]
### play/toClient/packet_update_time
["container", [{"name": "age", "type": "i64"}, {"name": "time", "type": "i64"}]]
### play/toClient/packet_held_item_slot
["container", [{"name": "slot", "type": "i8"}]]
### play/toClient/packet_player_info
["container", [{"name": "action", "type": ["bitflags", {"type": "u8", "flags": ["add_player", "initialize_chat", "update_game_mode", "update_listed", "update_latency", "update_display_name"]}]}, {"name": "data", "type": ["array", {"countType": "varint", "type": ["container", [{"name": "uuid", "type": "UUID"}, {"name": "player", "type": ["switch", {"compareTo": "../action/add_player", "fields": {"true": "game_profile"}, "default": "void"}]}, {"name": "chatSession", "type": ["switch", {"compareTo": "../action/initialize_chat", "fields": {"true": "chat_session"}, "default": "void"}]}, {"name": "gamemode", "type": ["switch", {"compareTo": "../action/update_game_mode", "fields": {"true": "varint"}, "default": "void"}]}, {"name": "listed", "type": ["switch", {"compareTo": "../action/update_listed", "fields": {"true": "varint"}, "default": "void"}]}, {"name": "latency", "type": ["switch", {"compareTo": "../action/update_latency", "fields": {"true": "varint"}, "default": "void"}]}, {"name": "displayName", "type": ["switch", {"compareTo": "../action/update_display_name", "fields": {"true": ["option", "anonymousNbt"]}, "default": "void"}]}]]}]}]]
### play/toClient/packet_player_remove
["container", [{"name": "players", "type": ["array", {"countType": "varint", "type": "UUID"}]}]]
### play/toClient/packet_update_view_position
["container", [{"name": "chunkX", "type": "varint"}, {"name": "chunkZ", "type": "varint"}]]
### play/toClient/packet_entity_teleport
["container", [{"name": "entityId", "type": "varint"}, {"name": "x", "type": "f64"}, {"name": "y", "type": "f64"}, {"name": "z", "type": "f64"}, {"name": "yaw", "type": "i8"}, {"name": "pitch", "type": "i8"}, {"name": "onGround", "type": "bool"}]]
### play/toClient/packet_rel_entity_move
["container", [{"name": "entityId", "type": "varint"}, {"name": "dX", "type": "i16"}, {"name": "dY", "type": "i16"}, {"name": "dZ", "type": "i16"}, {"name": "onGround", "type": "bool"}]]
### play/toClient/packet_entity_destroy
["container", [{"name": "entityIds", "type": ["array", {"countType": "varint", "type": "varint"}]}]]
### play/toClient/packet_unload_chunk
["container", [{"name": "chunkZ", "type": "i32"}, {"name": "chunkX", "type": "i32"}]]
### play/toClient/packet_declare_commands
["container", [{"name": "nodes", "type": ["array", {"countType": "varint", "type": "command_node"}]}, {"name": "rootIndex", "type": "varint"}]]
### play/toClient/packet_kick_disconnect
["container", [{"name": "reason", "type": "anonymousNbt"}]]
### play/toServer/packet_teleport_confirm
["container", [{"name": "teleportId", "type": "varint"}]]
### play/toServer/packet_position
["container", [{"name": "x", "type": "f64"}, {"name": "y", "type": "f64"}, {"name": "z", "type": "f64"}, {"name": "onGround", "type": "bool"}]]
### play/toServer/packet_position_look
["container", [{"name": "x", "type": "f64"}, {"name": "y", "type": "f64"}, {"name": "z", "type": "f64"}, {"name": "yaw", "type": "f32"}, {"name": "pitch", "type": "f32"}, {"name": "onGround", "type": "bool"}]]
### play/toServer/packet_look
["container", [{"name": "yaw", "type": "f32"}, {"name": "pitch", "type": "f32"}, {"name": "onGround", "type": "bool"}]]
### play/toServer/packet_flying
["container", [{"name": "onGround", "type": "bool"}]]
### play/toServer/packet_block_dig
["container", [{"name": "status", "type": "varint"}, {"name": "location", "type": "position"}, {"name": "face", "type": "i8"}, {"name": "sequence", "type": "varint"}]]
### play/toServer/packet_block_place
["container", [{"name": "hand", "type": "varint"}, {"name": "location", "type": "position"}, {"name": "direction", "type": "varint"}, {"name": "cursorX", "type": "f32"}, {"name": "cursorY", "type": "f32"}, {"name": "cursorZ", "type": "f32"}, {"name": "insideBlock", "type": "bool"}, {"name": "sequence", "type": "varint"}]]
### play/toServer/packet_keep_alive
["container", [{"name": "keepAliveId", "type": "i64"}]]
### play/toServer/packet_settings
null
### play/toServer/packet_custom_payload
["container", [{"name": "channel", "type": "string"}, {"name": "data", "type": "restBuffer"}]]
### play/toServer/packet_held_item_slot
["container", [{"name": "slotId", "type": "i16"}]]
### play/toServer/packet_set_creative_slot
["container", [{"name": "slot", "type": "i16"}, {"name": "item", "type": "Slot"}]]
### play/toServer/packet_arm_animation
["container", [{"name": "hand", "type": "varint"}]]
### play/toServer/packet_chat_message
["container", [{"name": "message", "type": "string"}, {"name": "timestamp", "type": "i64"}, {"name": "salt", "type": "i64"}, {"name": "signature", "type": ["option", ["buffer", {"count": 256}]]}, {"name": "offset", "type": "varint"}, {"name": "acknowledged", "type": ["buffer", {"count": 3}]}]]
### play/toServer/packet_chunk_batch_received
["container", [{"name": "chunksPerTick", "type": "f32"}]]
### configuration/toClient/packet_custom_payload
["container", [{"name": "channel", "type": "string"}, {"name": "data", "type": "restBuffer"}]]
### configuration/toClient/packet_registry_data
["container", [{"name": "id", "type": "string"}, {"name": "entries", "type": ["array", {"countType": "varint", "type": ["container", [{"name": "key", "type": "string"}, {"name": "value", "type": ["option", "anonymousNbt"]}]]}]}]]
### configuration/toClient/packet_feature_flags
["container", [{"name": "features", "type": ["array", {"countType": "varint", "type": "string"}]}]]
### configuration/toClient/packet_finish_configuration
["container", []]
### configuration/toClient/packet_keep_alive
["container", [{"name": "keepAliveId", "type": "i64"}]]
### configuration/toClient/packet_ping
["container", [{"name": "id", "type": "i32"}]]
### configuration/toClient/packet_select_known_packs
null
### configuration/toClient/packet_kick_disconnect
null
### configuration/toServer/packet_settings
null
### configuration/toServer/packet_custom_payload
["container", [{"name": "channel", "type": "string"}, {"name": "data", "type": "restBuffer"}]]
### configuration/toServer/packet_finish_configuration
["container", []]
### configuration/toServer/packet_keep_alive
["container", [{"name": "keepAliveId", "type": "i64"}]]
### configuration/toServer/packet_pong
["container", [{"name": "id", "type": "i32"}]]
### configuration/toServer/packet_select_known_packs
null
### login/toClient/packet_success
["container", [{"name": "uuid", "type": "UUID"}, {"name": "username", "type": "string"}, {"name": "properties", "type": ["array", {"countType": "varint", "type": ["container", [{"name": "name", "type": "string"}, {"name": "value", "type": "string"}, {"name": "signature", "type": ["option", "string"]}]]}]}, {"name": "strictErrorHandling", "type": "bool"}]]
### login/toClient/packet_compress
["container", [{"name": "threshold", "type": "varint"}]]
### login/toClient/packet_disconnect
["container", [{"name": "reason", "type": "string"}]]
### status/toClient/packet_server_info
["container", [{"name": "response", "type": "string"}]]
### status/toClient/packet_ping
["container", [{"name": "time", "type": "i64"}]]
### status/toServer/packet_ping_start
["container", []]
### status/toServer/packet_ping
["container", [{"name": "time", "type": "i64"}]]
### handshaking/toServer/packet_set_protocol
["container", [{"name": "protocolVersion", "type": "varint"}, {"name": "serverHost", "type": "string"}, {"name": "serverPort", "type": "u16"}, {"name": "nextState", "type": "varint"}]]