package atlasprobe;

import java.lang.reflect.Method;

// Run against the extracted 1.21.1 server JAR and its bundled libraries.
// Mojang's official 1.21.1 server mappings identify dzz as
// XoroshiroRandomSource. This probe is used only to verify the Go port.
public final class ReferenceProbe {
	private static final Class<?> RNG;
	private static final Method NEXT_LONG, NEXT_INT, NEXT_FLOAT, NEXT_DOUBLE, POSITIONAL;
	static {
		try {
			RNG = Class.forName("dzz");
			NEXT_LONG = RNG.getMethod("g");
			NEXT_INT = RNG.getMethod("a", int.class);
			NEXT_FLOAT = RNG.getMethod("i");
			NEXT_DOUBLE = RNG.getMethod("j");
			POSITIONAL = RNG.getMethod("e");
		} catch (ReflectiveOperationException error) {
			throw new ExceptionInInitializerError(error);
		}
	}
	private static Object rng(long seed) throws ReflectiveOperationException {
		return RNG.getConstructor(long.class).newInstance(seed);
	}
	public static void main(String[] args) throws Exception {
        long[] seeds = {0L, 1L, -1L, 7122209284156038298L};
        int[] bounds = {1, 3, 10, 1000, Integer.MAX_VALUE};
        for (long seed : seeds) {
            Object longs = rng(seed);
            for (int i = 0; i < 5; i++) {
                System.out.printf("long %d %d %s%n", seed, i,
                    Long.toUnsignedString((Long) NEXT_LONG.invoke(longs)));
            }
            Object bounded = rng(seed);
            for (int bound : bounds) {
                System.out.printf("bound %d %d %d%n", seed, bound,
                    NEXT_INT.invoke(bounded, bound));
            }
            Object fractions = rng(seed);
            System.out.printf("float %d %s%n", seed,
                Integer.toUnsignedString(Float.floatToRawIntBits((Float) NEXT_FLOAT.invoke(fractions))));
            System.out.printf("double %d %s%n", seed,
                Long.toUnsignedString(Double.doubleToRawLongBits((Double) NEXT_DOUBLE.invoke(fractions))));
            Object positions = POSITIONAL.invoke(rng(seed));
            Method at = positions.getClass().getMethod("a", int.class, int.class, int.class);
            Method fromHash = positions.getClass().getMethod("a", String.class);
            System.out.printf("position %d %s%n", seed,
                Long.toUnsignedString((Long) NEXT_LONG.invoke(at.invoke(positions, -17, 68, 33))));
            System.out.printf("hash %d %s%n", seed,
                Long.toUnsignedString((Long) NEXT_LONG.invoke(fromHash.invoke(positions, "minecraft:terrain"))));
        }
    }
}
