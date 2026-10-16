package world

import "testing"

func BenchmarkGeneratedAnimalRealization(b *testing.B) {
	pending, _ := pendingEntityTag("minecraft:pig", -16, 70, 16, "InitialAnimal:0").Object()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, ok, err := realizeGeneratedEntity(1234, "minecraft:overworld", pending); err != nil || !ok {
			b.Fatal(ok, err)
		}
	}
}
