package player

import (
	"atlas/internal/proto"
	"testing"
)

func TestWorldSinkDisconnectCancelsReservedChunks(t *testing.T) {
	p := New("test", [16]byte{}, 1, nil)
	sink := p.WorldPacketSink()
	if !sink.ReserveChunk(0, 0) {
		t.Fatal("could not reserve chunk")
	}
	p.Disconnect()
	if sink.WantsChunk(0, 0) {
		t.Fatal("disconnected view still wants queued generation")
	}
	if sink.ReserveChunk(1, 0) || sink.SendRequestedChunk(0, 0, []byte{1}) {
		t.Fatal("disconnected view accepted reservation or delivery")
	}
}

func TestWorldSinkReleasePreservesDeliveredTerrainAndNewView(t *testing.T) {
	p := New("test", [16]byte{}, 1, nil)
	old := p.WorldPacketSink()
	if !old.ReserveChunk(0, 0) || !old.SendRequestedChunk(0, 0, []byte{1}) {
		t.Fatal("could not deliver terrain")
	}
	old.ReleaseChunk(0, 0)
	if !old.IsChunkSent(0, 0) || old.ReserveChunk(0, 0) {
		t.Fatal("failure release forgot delivered terrain")
	}
	old.ReserveChunk(1, 0)
	p.ResetWorld("minecraft:overworld", [3]float64{0, 64, 0}, 0, 0, nil)
	current := p.WorldPacketSink()
	if !current.ReserveChunk(1, 0) {
		t.Fatal("new view could not reserve same coordinates")
	}
	old.ReleaseChunk(1, 0)
	if !current.WantsChunk(1, 0) || old.SendRequestedChunk(1, 0, []byte{2}) || old.ReserveChunk(2, 0) {
		t.Fatal("old view affected new view's reservation")
	}
	if _, wanted := old.ChunkPriority(1, 0); wanted {
		t.Fatal("old view kept generation alive")
	}
	current.ReleaseChunk(1, 0)
	if !current.ReserveChunk(1, 0) {
		t.Fatal("failure release prevented retry")
	}
}

func TestWorldSinkPriorityUsesLivePositionAndWideDistance(t *testing.T) {
	p := New("test", [16]byte{}, 1, nil)
	sink := p.WorldPacketSink()
	p.SetPos(.5, 64, .5, 0, 0, true)
	sink.ReserveChunk(625045, 624993)
	want := int64(625045)*625045 + int64(624993)*624993
	if distance, wanted := sink.ChunkPriority(625045, 624993); !wanted || distance != want {
		t.Fatal("distance overflow or missing reservation", distance, wanted)
	}
	p.SetPos(625045*16+.5, 64, 624993*16+.5, 0, 0, true)
	if distance, wanted := sink.ChunkPriority(625045, 624993); !wanted || distance != 0 {
		t.Fatal("priority did not follow movement", distance, wanted)
	}
}

func TestWorldSinkFullQueueDoesNotMarkPartialBatchDelivered(t *testing.T) {
	p := New("test", [16]byte{}, 1, nil)
	sink := p.WorldPacketSink()
	sink.ReserveChunk(0, 0)
	for i := 0; i < cap(p.Out)-1; i++ {
		p.SendPacket(proto.SKeepAlive, nil)
	}
	if sink.SendRequestedChunk(0, 0, []byte{1}) || sink.IsChunkSent(0, 0) {
		t.Fatal("partial batch counted as delivered")
	}
	select {
	case <-p.Done:
	default:
		t.Fatal("full queue did not disconnect slow client")
	}
}

func TestSoundSinkUsesDimensionThreeDimensionalRadiusAndEpoch(t *testing.T) {
	p := New("sound", [16]byte{}, 1, nil)
	p.SetPos(.5, 64.5, .5, 0, 0, true)
	if p.SoundPacketSink("minecraft:the_nether", .5, 64.5, .5, 16) != nil {
		t.Fatal("sound crossed dimensions")
	}
	if p.SoundPacketSink("minecraft:overworld", .5, 80.5, .5, 16) != nil {
		t.Fatal("sound included exact radius boundary")
	}
	sink := p.SoundPacketSink("minecraft:overworld", .5, 80.49, .5, 16)
	if sink == nil || !sink.Current() {
		t.Fatal("sound omitted nearby listener without a delivered source chunk")
	}
	p.ResetWorld("minecraft:the_nether", [3]float64{.5, 64.5, .5}, 0, 0, nil)
	if sink.Current() {
		t.Fatal("delayed sound survived a dimension change")
	}
}
