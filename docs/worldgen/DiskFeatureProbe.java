package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.lang.reflect.Method;
import java.lang.reflect.Proxy;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.HashMap;
import java.util.HexFormat;
import java.util.Map;
import java.util.Optional;
import java.util.stream.Stream;

public final class DiskFeatureProbe {
    public static void main(String[] args) throws Exception {
        Object version = Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a", Class.forName("ae")).invoke(null, version);
        Class.forName("akt").getMethod("a").invoke(null);
        Object registries = Class.forName("pd").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out), true));

        Object lookup = Class.forName("jo$a").getMethod("b").invoke(registries);
        Object featureLookup = Class.forName("jn$a").getMethod("b", Class.forName("akq"))
            .invoke(lookup, Class.forName("lu").getField("aQ").get(null));
        Map<String, Object> configurations = new HashMap<>();
        for (Object holder : ((Stream<?>) Class.forName("jo").getMethod("b")
            .invoke(featureLookup)).toList()) {
            Object key = ((Optional<?>) Class.forName("jm").getMethod("e").invoke(holder)).orElseThrow();
            String name = Class.forName("akq").getMethod("a").invoke(key).toString();
            Object placed = Class.forName("jm").getMethod("a").invoke(holder);
            Object configuredHolder = Class.forName("eiv").getMethod("b").invoke(placed);
            Object configured = Class.forName("jm").getMethod("a").invoke(configuredHolder);
            configurations.put(name, configured);
        }

        Class<?> registryClass = Class.forName("jz");
        Class<?> blockClass = Class.forName("dfy");
        Class<?> positionClass = Class.forName("jd");
        Object blockRegistry = Class.forName("lt").getField("e").get(null);
        Method blockAt = registryClass.getMethod("a", Class.forName("akr"));
        Method resourceLocation = Class.forName("akr").getMethod("a", String.class);
        Object dirt = blockClass.getMethod("o").invoke(blockAt.invoke(blockRegistry,
            resourceLocation.invoke(null, "minecraft:dirt")));
        Object air = blockClass.getMethod("o").invoke(Class.forName("dga").getField("a").get(null));
        Method coordinateX = positionClass.getMethod("u");
        Method coordinateY = positionClass.getMethod("v");
        Method coordinateZ = positionClass.getMethod("w");
        Method nextLong = Class.forName("dzx").getMethod("g");

        for (String name : new String[] {"minecraft:disk_clay", "minecraft:disk_gravel",
            "minecraft:ice_patch"}) {
            Object configured = configurations.get(name);
            if (configured == null) throw new IllegalStateException(name + " missing");
            Object feature = Class.forName("ebq").getMethod("b").invoke(configured);
            Object config = Class.forName("ebq").getMethod("c").invoke(configured);
            for (long seed : new long[] {0L, 1L, -1L, 123456789L}) {
                for (int[] origin : new int[][] {{8, 40, 8}, {15, 42, 15}, {-8, 40, -8},
                    {8, 100, 8}}) {
                    Map<String, Object> changed = new HashMap<>();
                    Object world = Proxy.newProxyInstance(Class.forName("dds").getClassLoader(),
                        new Class<?>[] {Class.forName("dds")}, (proxy, method, values) -> {
                            String methodName = method.getName();
                            if (methodName.equals("I_")) return -64;
                            if (methodName.equals("J_")) return 384;
                            if (methodName.equals("f_")) return true;
                            if (methodName.equals("toString")) return "atlas-disk-probe-world";
                            if (values != null && values.length > 0 && positionClass.isInstance(values[0])) {
                                int x = (int) coordinateX.invoke(values[0]);
                                int y = (int) coordinateY.invoke(values[0]);
                                int z = (int) coordinateZ.invoke(values[0]);
                                String key = x + " " + y + " " + z;
                                if (methodName.equals("a_")) return changed.getOrDefault(key,
                                    y == 40 ? dirt : air);
                                if (methodName.equals("a") && values.length == 3) {
                                    changed.put(key, values[1]);
                                    return true;
                                }
                            }
                            throw new UnsupportedOperationException(method.toString());
                        });
                    Object random = Class.forName("dzx").getConstructor(Class.forName("ayw"))
                        .newInstance(Class.forName("dzz").getConstructor(long.class).newInstance(seed));
                    Object start = positionClass.getConstructor(int.class, int.class, int.class)
                        .newInstance(origin[0], origin[1], origin[2]);
                    Object context = Class.forName("ecg").getConstructor(Optional.class,
                        Class.forName("dds"), Class.forName("duz"), Class.forName("ayw"),
                        positionClass, Class.forName("eei"))
                        .newInstance(Optional.empty(), world, null, random, start, config);
                    boolean placed = (boolean) Class.forName("ece").getMethod("a", Class.forName("ecg"))
                        .invoke(feature, context);
                    String positions = String.join("\n", changed.keySet().stream().sorted().toList());
                    String hash = HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256")
                        .digest(positions.getBytes(StandardCharsets.UTF_8)));
                    System.out.printf("disk_case %s %d %d %d %d %d %d %d %s%n", name, seed,
                        origin[0], origin[1], origin[2], placed ? 1 : 0, changed.size(),
                        nextLong.invoke(random), hash);
                }
            }
        }
    }
}
