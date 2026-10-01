package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.lang.reflect.Method;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;
import java.util.Optional;
import java.util.stream.Stream;

public final class BiomeTemperatureProbe {
    public static void main(String[] args) throws Exception {
        Object version = Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a", Class.forName("ae")).invoke(null, version);
        Class.forName("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out), true));
        Object registries = Class.forName("pd").getMethod("a").invoke(null);
        Object lookup = Class.forName("jo$a").getMethod("b").invoke(registries);
        Object biomeRegistry = Class.forName("jn$a").getMethod("b", Class.forName("akq"))
            .invoke(lookup, Class.forName("lu").getField("aF").get(null));
        List<Object> biomes = new ArrayList<>(((Stream<?>) Class.forName("jo").getMethod("b").invoke(biomeRegistry)).toList());
        biomes.sort(Comparator.comparing(biome -> {
            try {
                Object key = ((Optional<?>) Class.forName("jm").getMethod("e").invoke(biome)).orElseThrow();
                return Class.forName("akq").getMethod("a").invoke(key).toString();
            } catch (ReflectiveOperationException error) {
                throw new RuntimeException(error);
            }
        }));
        if (biomes.size() != 64) throw new IllegalStateException("unexpected biome count");
        Class<?> positionClass = Class.forName("jd");
        Method temperature = Class.forName("ddw").getDeclaredMethod("f", positionClass);
        temperature.setAccessible(true);
        Method cold = Class.forName("ddw").getMethod("b", positionClass);
        int[][] points = {{0, 64, 0}, {0, 80, 0}, {0, 81, 0}, {173, -32, -421},
            {-10000, 120, 24567}, {31, 200, -47}, {-2, 256, -1}, {65536, 300, 1048576},
            {127, 318, 611}, {-4670, 67, 14270}};
        for (Object holder : biomes) {
            Object key = ((Optional<?>) Class.forName("jm").getMethod("e").invoke(holder)).orElseThrow();
            String name = Class.forName("akq").getMethod("a").invoke(key).toString();
            Object biome = Class.forName("jm").getMethod("a").invoke(holder);
            for (int[] point : points) {
                Object position = positionClass.getConstructor(int.class, int.class, int.class)
                    .newInstance(point[0], point[1], point[2]);
                float value = (Float) temperature.invoke(biome, position);
                System.out.printf("temperature %s %d %d %d %s %b%n", name,
                    point[0], point[1], point[2], Integer.toUnsignedString(Float.floatToRawIntBits(value)),
                    cold.invoke(biome, position));
            }
        }
    }
}
