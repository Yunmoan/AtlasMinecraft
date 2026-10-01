package atlasprobe;

import com.google.gson.JsonArray;
import com.google.gson.JsonObject;
import com.google.gson.JsonParser;
import java.lang.reflect.Method;
import java.nio.charset.StandardCharsets;
import java.util.List;
import java.util.zip.ZipFile;

// Reads noise parameters directly from the extracted 1.21.1 server JAR.
public final class NamedNoiseProbe {
    public static void main(String[] args) throws Exception {
        Class<?> random = Class.forName("ayw");
        Class<?> rngType = Class.forName("dzz");
        Method positional = rngType.getMethod("e");
        Class<?> noiseType = Class.forName("eoh");
        Method create = noiseType.getMethod("a", random, int.class, double[].class);
        Method sample = noiseType.getMethod("a", double.class, double.class, double.class);
        long[] seeds = {0L, 1L, -1L, 7122209284156038298L};
        double[][] points = {{0, 0, 0}, {-17.125, 68.875, 33.5},
            {30000000.5, -63.25, -29999999.75}};
        try (ZipFile jar = new ZipFile(args[0])) {
			String prefix = "data/minecraft/worldgen/noise/";
			List<String> names = jar.stream().map(entry -> entry.getName())
				.filter(path -> path.startsWith(prefix) && path.endsWith(".json"))
				.map(path -> path.substring(prefix.length(), path.length() - 5))
				.sorted().toList();
            for (long seed : seeds) {
                for (String name : names) {
                    String path = "data/minecraft/worldgen/noise/" + name + ".json";
                    String json = new String(jar.getInputStream(jar.getEntry(path)).readAllBytes(),
                        StandardCharsets.UTF_8);
                    JsonObject parameters = JsonParser.parseString(json).getAsJsonObject();
                    int octave = parameters.get("firstOctave").getAsInt();
                    JsonArray source = parameters.getAsJsonArray("amplitudes");
                    double[] amplitudes = new double[source.size()];
                    for (int i = 0; i < source.size(); i++) {
                        amplitudes[i] = source.get(i).getAsDouble();
                    }
                    Object parent = rngType.getConstructor(long.class).newInstance(seed);
                    Object factory = positional.invoke(parent);
                    Method fromHash = factory.getClass().getMethod("a", String.class);
                    Object child = fromHash.invoke(factory, "minecraft:" + name);
                    Object noise = create.invoke(null, child, octave, amplitudes);
                    for (int i = 0; i < points.length; i++) {
                        double[] p = points[i];
                        double value = (Double) sample.invoke(noise, p[0], p[1], p[2]);
                        System.out.printf("named %d %s %d %s%n", seed, name, i,
                            Long.toUnsignedString(Double.doubleToRawLongBits(value)));
                    }
                }
            }
        }
    }
}
