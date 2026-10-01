package worldgen

// ImprovedNoise is Minecraft 1.21.1's gradient noise octave. Its permutation
// table and offsets consume a RandomSource in the same order as the JAR.
type ImprovedNoise struct {
	permutation               [256]byte
	XOffset, YOffset, ZOffset float64
}

func NewImprovedNoise(random *Xoroshiro) *ImprovedNoise {
	return newImprovedNoise(random)
}

type octaveRandom interface {
	NextDouble() float64
	NextIntBound(int32) int32
}

func newImprovedNoise(random octaveRandom) *ImprovedNoise {
	n := &ImprovedNoise{
		XOffset: random.NextDouble() * 256,
		YOffset: random.NextDouble() * 256,
		ZOffset: random.NextDouble() * 256,
	}
	for i := range n.permutation {
		n.permutation[i] = byte(i)
	}
	for i := range n.permutation {
		j := i + int(random.NextIntBound(int32(256-i)))
		n.permutation[i], n.permutation[j] = n.permutation[j], n.permutation[i]
	}
	return n
}

func (n *ImprovedNoise) Noise(x, y, z float64) float64 {
	return n.NoiseShifted(x, y, z, 0, 0)
}

// NoiseShifted implements the optional Y quantization used by octave samplers.
func (n *ImprovedNoise) NoiseShifted(x, y, z, yScale, yMax float64) float64 {
	x += n.XOffset
	y += n.YOffset
	z += n.ZOffset
	ix, iy, iz := fastFloor(x), fastFloor(y), fastFloor(z)
	fx, fy, fz := x-float64(ix), y-float64(iy), z-float64(iz)
	yShift := 0.0
	if yScale != 0 {
		toQuantize := fy
		if yMax >= 0 && yMax < fy {
			toQuantize = yMax
		}
		yShift = float64(fastFloor(toQuantize/yScale+1.0000000116860974e-7)) * yScale
	}
	return n.sampleAndLerp(ix, iy, iz, fx, fy-yShift, fz, fy)
}

func fastFloor(x float64) int32 {
	i := int32(x)
	if x < float64(i) {
		return i - 1
	}
	return i
}

func (n *ImprovedNoise) perm(i int32) int32 {
	return int32(n.permutation[uint32(i)&255])
}

var gradients = [16][3]float64{
	{1, 1, 0}, {-1, 1, 0}, {1, -1, 0}, {-1, -1, 0},
	{1, 0, 1}, {-1, 0, 1}, {1, 0, -1}, {-1, 0, -1},
	{0, 1, 1}, {0, -1, 1}, {0, 1, -1}, {0, -1, -1},
	{1, 1, 0}, {0, -1, 1}, {-1, 1, 0}, {0, -1, -1},
}

func gradientDot(hash int32, x, y, z float64) float64 {
	g := gradients[hash&15]
	return g[0]*x + g[1]*y + g[2]*z
}

func smoothstep(x float64) float64 { return x * x * x * (x*(x*6-15) + 10) }
func lerp(t, a, b float64) float64 { return a + t*(b-a) }
func lerp2(x, y, a, b, c, d float64) float64 {
	return lerp(y, lerp(x, a, b), lerp(x, c, d))
}
func lerp3(x, y, z, a, b, c, d, e, f, g, h float64) float64 {
	return lerp(z, lerp2(x, y, a, b, c, d), lerp2(x, y, e, f, g, h))
}

func (n *ImprovedNoise) sampleAndLerp(ix, iy, iz int32, x, y, z, fadeY float64) float64 {
	a, b := n.perm(ix), n.perm(ix+1)
	aa, ab := n.perm(a+iy), n.perm(a+iy+1)
	ba, bb := n.perm(b+iy), n.perm(b+iy+1)
	g000 := gradientDot(n.perm(aa+iz), x, y, z)
	g100 := gradientDot(n.perm(ba+iz), x-1, y, z)
	g010 := gradientDot(n.perm(ab+iz), x, y-1, z)
	g110 := gradientDot(n.perm(bb+iz), x-1, y-1, z)
	g001 := gradientDot(n.perm(aa+iz+1), x, y, z-1)
	g101 := gradientDot(n.perm(ba+iz+1), x-1, y, z-1)
	g011 := gradientDot(n.perm(ab+iz+1), x, y-1, z-1)
	g111 := gradientDot(n.perm(bb+iz+1), x-1, y-1, z-1)
	return lerp3(smoothstep(x), smoothstep(fadeY), smoothstep(z),
		g000, g100, g010, g110, g001, g101, g011, g111)
}
