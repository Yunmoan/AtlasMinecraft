package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.lang.reflect.Field;
import java.lang.reflect.Method;

public final class SurfaceNoiseProbe {
    public static void main(String[] args) throws Exception {
        Object version = Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a", Class.forName("ae")).invoke(null, version);
        Class.forName("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out), true));
        Object registries = Class.forName("pd").getMethod("a").invoke(null);
        Object lookup = Class.forName("jo$a").getMethod("b").invoke(registries);
        Object settingsKey = Class.forName("dzd").getField("c").get(null);
        Class<?> registryKey = Class.forName("akq");
        Object settingsRegistry = Class.forName("jn$a").getMethod("b", registryKey)
            .invoke(lookup, Class.forName("lu").getField("aO").get(null));
        Object settingsHolder = Class.forName("jn").getMethod("b", registryKey).invoke(settingsRegistry, settingsKey);
        Class<?> surfaceClass = Class.forName("dzq");
        Method sample = Class.forName("eoh").getMethod("a", double.class, double.class, double.class);
        String[] fields = {"n", "o", "p", "q", "r", "s"};
        String[] names = {"badlands_pillar", "badlands_pillar_roof", "badlands_surface",
            "iceberg_pillar", "iceberg_pillar_roof", "iceberg_surface"};
        int[][] positions = {{2, 0}, {0, 1}, {13, 13}, {5, 14}, {173, -432}, {-10000, 24567}};
        long[] seeds = {0L, 1L, -1L, 7122209284156038298L};
        for (long seed : seeds) {
            Object random = Class.forName("dzm").getMethod("a", Class.forName("jn$a"), registryKey, long.class)
                .invoke(null, lookup, settingsKey, seed);
            Object surface = Class.forName("dzm").getMethod("c").invoke(random);
            for (int index = 0; index < fields.length; index++) {
                Field field = surfaceClass.getDeclaredField(fields[index]);
                field.setAccessible(true);
                Object noise = field.get(surface);
                for (int[] position : positions) {
                    int x = position[0], z = position[1];
                    double scale = 1;
                    if (index == 0) scale = 0.2;
                    if (index == 1) scale = 0.75;
                    if (index == 3) scale = 1.28;
                    if (index == 4) scale = 1.17;
                    double value = (Double) sample.invoke(noise, x*scale, 0.0, z*scale);
                    System.out.printf("surface_noise %d %s %d %d %s%n", seed, names[index], x, z,
                        Long.toUnsignedString(Double.doubleToRawLongBits(value)));
                }
            }
        }
    }
}
