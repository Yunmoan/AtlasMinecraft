package atlasprobe;

import com.google.gson.JsonObject;
import com.google.gson.JsonParser;
import java.lang.reflect.Method;
import java.lang.reflect.Proxy;
import java.nio.charset.StandardCharsets;
import java.util.zip.ZipFile;

// Reference values for old_blended_noise, using the terrain positional seed.
public final class BlendedNoiseProbe {
    public static void main(String[] args) throws Exception {
        Object version = Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a", Class.forName("ae")).invoke(null, version);
        Class.forName("akt").getMethod("a").invoke(null);
        Class<?> source = Class.forName("ayw");
        Class<?> context = Class.forName("dyr$b");
        Class<?> blended = Class.forName("eoe");
        Class<?> rng = Class.forName("dzz");
        Method positional = rng.getMethod("e");
        Method sample = blended.getMethod("a", context);
        long[] seeds = {0L, 1L, -1L, 7122209284156038298L};
        int[][] points = {{0, 64, 0}, {173, -32, -421}, {-10000, 120, 24567}};
        String[] names = {"overworld", "nether", "end"};
        try (ZipFile jar = new ZipFile(args[0])) {
            for (String name : names) {
                String path = "data/minecraft/worldgen/density_function/" + name + "/base_3d_noise.json";
                JsonObject p = JsonParser.parseString(new String(
                    jar.getInputStream(jar.getEntry(path)).readAllBytes(), StandardCharsets.UTF_8)).getAsJsonObject();
                for (long seed : seeds) {
                    Object parent = rng.getConstructor(long.class).newInstance(seed);
                    Object factory = positional.invoke(parent);
                    Object terrain = factory.getClass().getMethod("a", String.class)
                        .invoke(factory, "minecraft:terrain");
                    Object noise = blended.getConstructor(source, double.class, double.class,
                        double.class, double.class, double.class).newInstance(terrain,
                        p.get("xz_scale").getAsDouble(), p.get("y_scale").getAsDouble(),
                        p.get("xz_factor").getAsDouble(), p.get("y_factor").getAsDouble(),
                        p.get("smear_scale_multiplier").getAsDouble());
                    for (int i = 0; i < points.length; i++) {
                        int[] q = points[i];
                        Object point = Proxy.newProxyInstance(context.getClassLoader(),
                            new Class<?>[] {context}, (proxy, method, unused) -> switch (method.getName()) {
                                case "a" -> q[0];
                                case "b" -> q[1];
                                case "c" -> q[2];
                                default -> throw new UnsupportedOperationException(method.getName());
                            });
                        double value = (Double) sample.invoke(noise, point);
                        System.out.printf("blended %s %d %d %s%n", name, seed, i,
                            Long.toUnsignedString(Double.doubleToRawLongBits(value)));
                    }
                }
            }
        }
    }
}
