package atlasprobe;

import java.lang.reflect.Constructor;
import java.lang.reflect.Method;

// eof is ImprovedNoise in Mojang's official 1.21.1 server mappings.
public final class ImprovedNoiseProbe {
    public static void main(String[] args) throws Exception {
        Class<?> random = Class.forName("ayw");
        Constructor<?> rng = Class.forName("dzz").getConstructor(long.class);
        Class<?> type = Class.forName("eof");
        Constructor<?> constructor = type.getConstructor(random);
        Method sample = type.getMethod("a", double.class, double.class, double.class);
        Method shifted = type.getMethod("a", double.class, double.class, double.class,
            double.class, double.class);
        long[] seeds = {0L, 1L, -1L, 7122209284156038298L};
        double[][] points = {
            {0.0, 0.0, 0.0}, {1.25, -2.75, 4.5},
            {-17.125, 68.875, 33.5}, {30000000.5, -63.25, -29999999.75},
            {0.001, 0.999, -0.001}
        };
        for (long seed : seeds) {
            Object noise = constructor.newInstance(rng.newInstance(seed));
            for (String field : new String[] {"a", "b", "c"}) {
                double offset = type.getField(field).getDouble(noise);
                System.out.printf("offset %d %s %s%n", seed, field,
                    Long.toUnsignedString(Double.doubleToRawLongBits(offset)));
            }
            for (int i = 0; i < points.length; i++) {
                double[] p = points[i];
                double value = (Double) sample.invoke(noise, p[0], p[1], p[2]);
                System.out.printf("sample %d %d %s%n", seed, i,
                    Long.toUnsignedString(Double.doubleToRawLongBits(value)));
                value = (Double) shifted.invoke(noise, p[0], p[1], p[2], 0.25, 0.5);
                System.out.printf("shifted %d %d %s%n", seed, i,
                    Long.toUnsignedString(Double.doubleToRawLongBits(value)));
            }
        }
    }
}
