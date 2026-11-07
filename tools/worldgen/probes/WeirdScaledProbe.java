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

public final class WeirdScaledProbe {
    public static void main(String[] args) throws Exception {
        Object version = Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a", Class.forName("ae")).invoke(null, version);
        Class.forName("akt").getMethod("a").invoke(null);
        Class<?> density = Class.forName("dyr");
        Class<?> context = Class.forName("dyr$b");
        Class<?> holder = Class.forName("jm");
        Class<?> noiseHolder = Class.forName("dyr$c");
        Class<?> normalNoise = Class.forName("eoh");
        Class<?> random = Class.forName("ayw");
        Class<?> rng = Class.forName("dzz");
        Class<?> mapper = Class.forName("dys$z$a");
        Method createNoise = normalNoise.getMethod("a", random, int.class, double[].class);
        Method positional = rng.getMethod("e");
        Method directHolder = holder.getMethod("a", Object.class);
        Method constant = Class.forName("dys").getMethod("a", double.class);
        Method sample = density.getMethod("a", context);
        Constructor<?> ctor = Class.forName("dys$z").getDeclaredConstructor(density, noiseHolder, mapper);
        ctor.setAccessible(true);
        Constructor<?> noiseHolderCtor = noiseHolder.getConstructor(holder, normalNoise);
        long[] seeds = {0L, 1L, -1L, 7122209284156038298L};
        int[][] points = {{0, 64, 0}, {173, -32, -421}, {-10000, 120, 24567}};
        double[] inputs = {-0.9, -0.6, -0.2, 0.2, 0.6, 0.9};
        String[] types = {"type_1", "type_2"};
        String[] names = {"spaghetti_3d_1", "spaghetti_2d"};
        try (ZipFile jar = new ZipFile(args[0])) {
            for (int m = 0; m < types.length; m++) {
                String path = "data/minecraft/worldgen/noise/" + names[m] + ".json";
                JsonObject p = JsonParser.parseString(new String(
                    jar.getInputStream(jar.getEntry(path)).readAllBytes(), StandardCharsets.UTF_8)).getAsJsonObject();
                JsonArray values = p.getAsJsonArray("amplitudes");
                double[] amplitudes = new double[values.size()];
                List<Double> list = new ArrayList<>();
                for (int i = 0; i < values.size(); i++) {
                    amplitudes[i] = values.get(i).getAsDouble();
                    list.add(amplitudes[i]);
                }
                Object parameters = Class.forName("eoh$a").getConstructor(int.class, List.class)
                    .newInstance(p.get("firstOctave").getAsInt(), list);
                Object direct = directHolder.invoke(null, parameters);
                Object enumValue = mapper.getField(m == 0 ? "a" : "b").get(null);
                for (long seed : seeds) {
                    Object parent = rng.getConstructor(long.class).newInstance(seed);
                    Object factory = positional.invoke(parent);
                    Object child = factory.getClass().getMethod("a", String.class)
                        .invoke(factory, "minecraft:" + names[m]);
                    Object normal = createNoise.invoke(null, child, p.get("firstOctave").getAsInt(), amplitudes);
                    Object wrappedNoise = noiseHolderCtor.newInstance(direct, normal);
                    for (int inputIndex = 0; inputIndex < inputs.length; inputIndex++) {
                        Object node = ctor.newInstance(constant.invoke(null, inputs[inputIndex]), wrappedNoise, enumValue);
                        for (int pointIndex = 0; pointIndex < points.length; pointIndex++) {
                            int[] q = points[pointIndex];
                            Object point = Proxy.newProxyInstance(context.getClassLoader(),
                                new Class<?>[] {context}, (proxy, method, unused) -> switch (method.getName()) {
                                    case "a" -> q[0];
                                    case "b" -> q[1];
                                    case "c" -> q[2];
                                    default -> throw new UnsupportedOperationException(method.getName());
                                });
                            double result = (Double) sample.invoke(node, point);
                            System.out.printf("weird %s %d %d %d %s%n", types[m], seed,
                                inputIndex, pointIndex,
                                Long.toUnsignedString(Double.doubleToRawLongBits(result)));
                        }
                    }
                }
            }
        }
    }
}
