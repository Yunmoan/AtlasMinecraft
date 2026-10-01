package atlasprobe;

import java.lang.reflect.Method;

// eoh is NormalNoise in Mojang's official 1.21.1 server mappings.
public final class NormalNoiseProbe {
    public static void main(String[] args) throws Exception {
        Class<?> random = Class.forName("ayw");
        Class<?> type = Class.forName("eoh");
        Method create = type.getMethod("a", random, int.class, double[].class);
        Method sample = type.getMethod("a", double.class, double.class, double.class);
        long[] seeds = {0L, 1L, -1L, 7122209284156038298L};
        int[] firstOctaves = {-4, -2, -10};
        double[][] amplitudes = {
            {1, 1, 0, 1, 1}, {0.5, 0, 1}, {1, 1, 1, 1, 1}
        };
        double[][] points = {
            {0.0, 0.0, 0.0}, {1.25, -2.75, 4.5},
            {-17.125, 68.875, 33.5}, {30000000.5, -63.25, -29999999.75},
            {0.001, 0.999, -0.001}
        };
        for (long seed : seeds) {
            for (int c = 0; c < firstOctaves.length; c++) {
                Object rng = Class.forName("dzz").getConstructor(long.class).newInstance(seed);
                Object noise = create.invoke(null, rng, firstOctaves[c], amplitudes[c]);
                for (int i = 0; i < points.length; i++) {
                    double[] p = points[i];
                    double value = (Double) sample.invoke(noise, p[0], p[1], p[2]);
                    System.out.printf("normal %d %d %d %s%n", seed, c, i,
                        Long.toUnsignedString(Double.doubleToRawLongBits(value)));
                }
            }
        }
    }
}
