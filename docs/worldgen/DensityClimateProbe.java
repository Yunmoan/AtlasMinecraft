package atlasprobe;

import com.google.gson.JsonArray;
import com.google.gson.JsonObject;
import com.google.gson.JsonParser;
import java.lang.reflect.Constructor;
import java.lang.reflect.Method;
import java.lang.reflect.Proxy;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.List;
import java.util.zip.ZipFile;

// Assemble the same vanilla density nodes used by the climate graph, then
// sample them through DensityFunction.compute. The Go runtime never uses this.
public final class DensityClimateProbe {
    static Class<?> density;
    static Class<?> context;
    static Class<?> holder;
    static Class<?> noiseHolder;
    static Class<?> normalNoise;
    static Class<?> random;
    static Class<?> randomSource;
    static Method createNoise;
    static Method sampleDensity;
    static Method positional;
    static Method directHolder;
    static Constructor<?> shiftA;
    static Constructor<?> shiftB;
    static Constructor<?> shifted;
    static Constructor<?> noiseHolderConstructor;
    static ZipFile jar;

    static Object namedNoise(long seed, String name) throws Exception {
        String path = "data/minecraft/worldgen/noise/" + name + ".json";
        JsonObject p = JsonParser.parseString(new String(
            jar.getInputStream(jar.getEntry(path)).readAllBytes(), StandardCharsets.UTF_8)).getAsJsonObject();
        JsonArray source = p.getAsJsonArray("amplitudes");
        double[] amplitudes = new double[source.size()];
        for (int i = 0; i < source.size(); i++) amplitudes[i] = source.get(i).getAsDouble();
        Object parent = randomSource.getConstructor(long.class).newInstance(seed);
        Object factory = positional.invoke(parent);
        Object child = factory.getClass().getMethod("a", String.class).invoke(factory, "minecraft:" + name);
        return createNoise.invoke(null, child, p.get("firstOctave").getAsInt(), amplitudes);
    }

    static Object holder(long seed, String name) throws Exception {
        String path = "data/minecraft/worldgen/noise/" + name + ".json";
        JsonObject p = JsonParser.parseString(new String(
            jar.getInputStream(jar.getEntry(path)).readAllBytes(), StandardCharsets.UTF_8)).getAsJsonObject();
        JsonArray source = p.getAsJsonArray("amplitudes");
        List<Double> amplitudes = new ArrayList<>();
        for (int i = 0; i < source.size(); i++) amplitudes.add(source.get(i).getAsDouble());
        Object parameters = Class.forName("eoh$a").getConstructor(int.class, List.class)
            .newInstance(p.get("firstOctave").getAsInt(), amplitudes);
        Object direct = directHolder.invoke(null, parameters);
        return noiseHolderConstructor.newInstance(direct, namedNoise(seed, name));
    }

    static Object point(int x, int y, int z) {
        return Proxy.newProxyInstance(context.getClassLoader(), new Class<?>[] {context},
            (proxy, method, args) -> switch (method.getName()) {
                case "a" -> x;
                case "b" -> y;
                case "c" -> z;
                default -> throw new UnsupportedOperationException(method.getName());
            });
    }

    public static void main(String[] args) throws Exception {
        Object version = Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a", Class.forName("ae")).invoke(null, version);
        Class.forName("akt").getMethod("a").invoke(null); // Bootstrap registries.
        density = Class.forName("dyr");
        context = Class.forName("dyr$b");
        holder = Class.forName("jm");
        noiseHolder = Class.forName("dyr$c");
        normalNoise = Class.forName("eoh");
        random = Class.forName("ayw");
        randomSource = Class.forName("dzz");
        createNoise = normalNoise.getMethod("a", random, int.class, double[].class);
        sampleDensity = density.getMethod("a", context);
        positional = randomSource.getMethod("e");
        directHolder = holder.getMethod("a", Object.class);
        noiseHolderConstructor = noiseHolder.getConstructor(holder, normalNoise);
        shiftA = Class.forName("dys$s").getDeclaredConstructor(noiseHolder);
        shiftB = Class.forName("dys$t").getDeclaredConstructor(noiseHolder);
        shifted = Class.forName("dys$v").getDeclaredConstructor(
            density, density, density, double.class, double.class, noiseHolder);
        shiftA.setAccessible(true);
        shiftB.setAccessible(true);
        shifted.setAccessible(true);
        Method zero = Class.forName("dys").getMethod("a", double.class);
        long[] seeds = {0L, 1L, -1L, 7122209284156038298L};
        int[][] points = {{0, 64, 0}, {173, -32, -421}, {-10000, 120, 24567}};
        String[] climateNames = {"temperature", "vegetation", "continentalness", "erosion", "ridge"};
        String[] fields = {"temperature", "vegetation", "continents", "erosion", "ridges"};
        try (ZipFile opened = new ZipFile(args[0])) {
            jar = opened;
            for (long seed : seeds) {
                Object offset = holder(seed, "offset");
                Object sx = shiftA.newInstance(offset);
                Object sz = shiftB.newInstance(offset);
                Object[] functions = new Object[7];
                functions[0] = sx;
                functions[1] = sz;
                for (int i = 0; i < climateNames.length; i++) {
                    functions[i + 2] = shifted.newInstance(sx, zero.invoke(null, 0.0), sz,
                        0.25, 0.0, holder(seed, climateNames[i]));
                }
                for (int i = 0; i < functions.length; i++) {
                    String field = i == 0 ? "shift_x" : i == 1 ? "shift_z" : fields[i - 2];
                    for (int j = 0; j < points.length; j++) {
                        int[] p = points[j];
                        double value = (Double) sampleDensity.invoke(functions[i], point(p[0], p[1], p[2]));
                        System.out.printf("density %d %s %d %s%n", seed, field, j,
                            Long.toUnsignedString(Double.doubleToRawLongBits(value)));
                    }
                }
            }
        }
    }
}
