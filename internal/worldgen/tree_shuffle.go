package worldgen

import (
	"crypto/sha256"
	"encoding/binary"
	"math"
)

// Vanilla beehives call Collections.shuffle using a process-global Java RNG,
// independent of the world seed. Atlas freezes that independent choice per
// canonical tree invocation because neighboring decoration is replayed. The
// salt is random per saved world, persists across restarts, and requires no
// unbounded tree cache. This source differs from vanilla's process RNG while
// preserving the decorator algorithm and its worldgen RNG consumption.
type treeShuffleContext struct {
	seed, salt int64
	sourceX, sourceZ int32
	step, index int
}

func (context treeShuffleContext) forTree(name string,p FeaturePosition,r *WorldgenRandom) func([]FeaturePosition) {
	var key [89]byte
	binary.LittleEndian.PutUint64(key[0:],uint64(context.seed))
	binary.LittleEndian.PutUint64(key[8:],uint64(context.salt))
	binary.LittleEndian.PutUint32(key[16:],uint32(context.sourceX))
	binary.LittleEndian.PutUint32(key[20:],uint32(context.sourceZ))
	binary.LittleEndian.PutUint64(key[24:],uint64(context.step))
	binary.LittleEndian.PutUint64(key[32:],uint64(context.index))
	binary.LittleEndian.PutUint32(key[40:],uint32(p.X))
	binary.LittleEndian.PutUint32(key[44:],uint32(p.Y))
	binary.LittleEndian.PutUint32(key[48:],uint32(p.Z))
	lo,hi:=r.XoroshiroState()
	binary.LittleEndian.PutUint64(key[52:],lo)
	binary.LittleEndian.PutUint64(key[60:],hi)
	binary.LittleEndian.PutUint64(key[68:],math.Float64bits(r.gaussian))
	if r.hasGaussian { key[76]=1 }
	binary.LittleEndian.PutUint64(key[77:],uint64(len(name)))
	// The final four bytes reserve a derivation version without changing the
	// world generation RNG or confusing one configured feature with another.
	binary.LittleEndian.PutUint32(key[85:],1)
	hash:=sha256.New()
	hash.Write(key[:])
	hash.Write([]byte(name))
	seed:=int64(binary.LittleEndian.Uint64(hash.Sum(nil)))
	return func(positions []FeaturePosition) {
		random:=NewLegacyRandom(seed)
		for i:=len(positions);i>1;i-- {
			j:=int(random.NextIntBound(int32(i)))
			positions[i-1],positions[j]=positions[j],positions[i-1]
		}
	}
}
