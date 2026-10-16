package worldgen

import "math"

// The kernels and reduction constants below are derived from fdlibm.
// Copyright (C) 1993 by Sun Microsystems, Inc. All rights reserved.
// Developed at SunSoft, a Sun Microsystems, Inc. business.
// Permission to use, copy, modify, and distribute this software is freely
// granted, provided that this notice is preserved.
//
// Worldgen feature angles are bounded by 5*pi/2. Go's different trigonometric
// approximation can change ellipse boundary tests by one ulp. Explicit
// conversions preserve the Java evaluation order without fused operations.
func featureSinCos(angle float64) (float64, float64) {
	if math.Abs(angle) > 8 || math.IsNaN(angle) {
		panic("worldgen: feature angle outside fdlibm reduction range")
	}
	x, y, n := featureReduceAngle(angle)
	s, c := featureKernelSin(x, y, n != 0), featureKernelCos(x, y)
	switch n & 3 {
	case 1:
		return c, -s
	case 2:
		return -s, -c
	case 3:
		return -c, s
	}
	return s, c
}

func featureReduceAngle(angle float64) (float64, float64, int) {
	const (
		inv = 6.36619772367581382433e-01
		p1  = 1.57079632673412561417e+00
		t1  = 6.07710050650619224932e-11
		p2  = 6.07710050630396597660e-11
		t2  = 2.02226624879595063154e-21
		p3  = 2.02226624871116645580e-21
		t3  = 8.47842766036889956997e-32
	)
	ix := uint32(math.Float64bits(angle)>>32) & 0x7fffffff
	if ix <= 0x3fe921fb {
		return angle, 0, 0
	}
	t := math.Abs(angle)
	if ix < 0x4002d97c {
		z := t - p1
		tail := t1
		if ix == 0x3ff921fb {
			z -= p2
			tail = t2
		}
		x := z - tail
		y := (z - x) - tail
		if angle < 0 {
			return -x, -y, -1
		}
		return x, y, 1
	}
	n := int(float64(t*inv) + 0.5)
	fn := float64(n)
	r := t - float64(fn*p1)
	w := float64(fn * t1)
	x := r - w
	exponent := int(ix >> 20)
	if exponent-int((math.Float64bits(x)>>52)&0x7ff) > 16 {
		t = r
		w = float64(fn * p2)
		r = t - w
		w = float64(fn*t2) - ((t - r) - w)
		x = r - w
		if exponent-int((math.Float64bits(x)>>52)&0x7ff) > 49 {
			t = r
			w = float64(fn * p3)
			r = t - w
			w = float64(fn*t3) - ((t - r) - w)
			x = r - w
		}
	}
	y := (r - x) - w
	if angle < 0 {
		return -x, -y, -n
	}
	return x, y, n
}

func featureKernelSin(x, y float64, tail bool) float64 {
	const (
		s1 = -1.66666666666666324348e-01
		s2 = 8.33333333332248946124e-03
		s3 = -1.98412698298579493134e-04
		s4 = 2.75573137070700676789e-06
		s5 = -2.50507602534068634195e-08
		s6 = 1.58969099521155010221e-10
	)
	if math.Abs(x) < 0x1p-27 {
		return x
	}
	z := float64(x * x)
	v := float64(z * x)
	r := s2 + float64(z*(s3+float64(z*(s4+float64(z*(s5+float64(z*s6)))))))
	if !tail {
		return x + float64(v*(s1+float64(z*r)))
	}
	return x - ((float64(z*(float64(0.5*y)-float64(v*r))) - y) - float64(v*s1))
}

func featureKernelCos(x, y float64) float64 {
	const (
		c1 = 4.16666666666666019037e-02
		c2 = -1.38888888888741095749e-03
		c3 = 2.48015872894767294178e-05
		c4 = -2.75573143513906633035e-07
		c5 = 2.08757232129817482790e-09
		c6 = -1.13596475577881948265e-11
	)
	ix := uint32(math.Float64bits(x)>>32) & 0x7fffffff
	if ix < 0x3e400000 {
		return 1
	}
	z := float64(x * x)
	r := float64(z * (c1 + float64(z*(c2+float64(z*(c3+float64(z*(c4+float64(z*(c5+float64(z*c6)))))))))))
	correction := float64(z*r) - float64(x*y)
	if ix < 0x3fd33333 {
		return 1 - (float64(0.5*z) - correction)
	}
	qx := 0.28125
	if ix <= 0x3fe90000 {
		qx = math.Float64frombits(uint64(ix-0x00200000) << 32)
	}
	hz := float64(0.5*z) - qx
	return (1 - qx) - (hz - correction)
}
