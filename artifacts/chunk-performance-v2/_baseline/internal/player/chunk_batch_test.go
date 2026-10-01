package player

import (
	"atlas/internal/proto"
	"sync"
	"testing"
)

func TestSingleGeneratedChunkFinishesBatch(t *testing.T) {
	p := New("test", [16]byte{}, 1, nil)
	p.SendChunkData([]byte{42})
	checkChunkBatches(t, p, 1)
}

func TestConcurrentRegionChunksHaveCompleteBatches(t *testing.T) {
	p := New("test", [16]byte{}, 1, nil)
	var workers sync.WaitGroup
	for i := 0; i < 64; i++ {
		workers.Add(1)
		go func(i int) { defer workers.Done(); p.SendChunkData([]byte{byte(i)}) }(i)
	}
	workers.Wait()
	checkChunkBatches(t, p, 64)
}

func checkChunkBatches(t *testing.T, p *Player, chunks int) {
	t.Helper()
	seen := make(map[byte]bool)
	for i := 0; i < chunks; i++ {
		if len(p.Out) < 3 {
			t.Fatalf("batch %d was not completed immediately", i)
		}
		start, data, finish := <-p.Out, <-p.Out, <-p.Out
		if start.ID != proto.SChunkBatchStart || data.ID != proto.SChunkData || finish.ID != proto.SChunkBatchFinished {
			t.Fatalf("incomplete or nested batch %d: %d,%d,%d", i, start.ID, data.ID, finish.ID)
		}
		count, err := proto.NewReader(finish.Payload).VarInt()
		if err != nil || count != 1 {
			t.Fatalf("finished batch size: %d, %v", count, err)
		}
		if len(data.Payload) != 1 || seen[data.Payload[0]] {
			t.Fatal("missing or duplicated chunk data")
		}
		seen[data.Payload[0]] = true
	}
	select {
	case <-p.Out:
		t.Fatal("extra chunk batch packets")
	default:
	}
}
