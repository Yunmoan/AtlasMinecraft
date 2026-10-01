package atlasprobe;

import java.lang.reflect.Method;

// dyz is LegacyRandomSource in Mojang's official 1.21.1 server mappings.
public final class LegacyRandomProbe {
    public static void main(String[] args) throws Exception {
        Class<?> type = Class.forName("dyz");
        Method nextLong = type.getMethod("g");
        Method nextBound = type.getMethod("a", int.class);
        Method nextFloat = type.getMethod("i");
        Method nextDouble = type.getMethod("j");
        Method positional = type.getMethod("e");
        long[] seeds = {0L, 1L, -1L, 7122209284156038298L};
        int[] bounds = {1, 3, 10, 1000, Integer.MAX_VALUE};
        for (long seed : seeds) {
            Object longs = type.getConstructor(long.class).newInstance(seed);
            for (int i = 0; i < 5; i++) {
                System.out.printf("long %d %d %s%n", seed, i,
                    Long.toUnsignedString((Long) nextLong.invoke(longs)));
            }
            Object bounded = type.getConstructor(long.class).newInstance(seed);
            for (int bound : bounds) {
                System.out.printf("bound %d %d %d%n", seed, bound,
                    nextBound.invoke(bounded, bound));
            }
            Object fractions = type.getConstructor(long.class).newInstance(seed);
            System.out.printf("float %d %s%n", seed,
                Integer.toUnsignedString(Float.floatToRawIntBits((Float) nextFloat.invoke(fractions))));
            System.out.printf("double %d %s%n", seed,
                Long.toUnsignedString(Double.doubleToRawLongBits((Double) nextDouble.invoke(fractions))));
            Object positions = positional.invoke(type.getConstructor(long.class).newInstance(seed));
            Method at = positions.getClass().getMethod("a", int.class, int.class, int.class);
            Method hash = positions.getClass().getMethod("a", String.class);
            System.out.printf("position %d %s%n", seed,
                Long.toUnsignedString((Long) nextLong.invoke(at.invoke(positions, -17, 68, 33))));
            System.out.printf("hash %d %s%n", seed,
                Long.toUnsignedString((Long) nextLong.invoke(hash.invoke(positions, "minecraft:terrain"))));
        }
    }
}
