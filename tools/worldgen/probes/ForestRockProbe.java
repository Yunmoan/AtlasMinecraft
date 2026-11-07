package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.lang.reflect.Method;
import java.lang.reflect.Proxy;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.stream.Stream;

public final class ForestRockProbe {
    public static void main(String[] args) throws Exception {
        Object version = Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a", Class.forName("ae")).invoke(null, version);
        Class.forName("akt").getMethod("a").invoke(null);

        Class<?> registryClass = Class.forName("jz");
        Class<?> blockClass = Class.forName("dfy");
        Class<?> positionClass = Class.forName("jd");
        Object blockRegistry = Class.forName("lt").getField("e").get(null);
        Object stoneBlock = Class.forName("dga").getField("b").get(null);
        Method blockAt = registryClass.getMethod("a", Class.forName("akr"));
        Method resourceLocation = Class.forName("akr").getMethod("a", String.class);
        Object mossBlock = blockAt.invoke(blockRegistry, resourceLocation.invoke(null,
            "minecraft:mossy_cobblestone"));
        Object stone = blockClass.getMethod("o").invoke(stoneBlock);
        Object moss = blockClass.getMethod("o").invoke(mossBlock);
        Object airBlock = Class.forName("dga").getField("a").get(null);
        Object air = blockClass.getMethod("o").invoke(airBlock);
        Object holder = registryClass.getMethod("e", Object.class).invoke(blockRegistry, stoneBlock);
        Object dirtBlock = blockAt.invoke(blockRegistry, resourceLocation.invoke(null, "minecraft:dirt"));
        Object dirtHolder = registryClass.getMethod("e", Object.class).invoke(blockRegistry, dirtBlock);
        Map<Object, List<Object>> tags = new HashMap<>();
        for (Object tag : ((Stream<?>) registryClass.getMethod("k").invoke(blockRegistry)).toList())
            tags.put(tag, List.of());
        tags.put(Class.forName("awe").getField("bf").get(null), List.of(holder));
        tags.put(Class.forName("awe").getField("ag").get(null), List.of(dirtHolder));
        registryClass.getMethod("a", Map.class).invoke(blockRegistry, tags);
        Object registries = Class.forName("pd").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out), true));

        Method coordinateX = positionClass.getMethod("u");
        Method coordinateY = positionClass.getMethod("v");
        Method coordinateZ = positionClass.getMethod("w");
        Method nextLong = Class.forName("dzx").getMethod("g");
        Method place = Class.forName("ece").getMethod("a", Class.forName("ecg"));
        Object forestRock = Class.forName("ece").getField("F").get(null);
        Object config = Class.forName("eeb").getConstructor(Class.forName("dtc")).newInstance(moss);
        long[] seeds = {0L, 1L, -1L, 123456789L};
        int[][] origins = {{8, 45, 8}, {15, 42, 15}, {-8, 70, -8},
            {8, -61, 8}, {8, 100, 8}};
        for (long seed : seeds) {
            for (int[] origin : origins) {
                boolean ground = origin[1] != 100;
                Map<String, Object> changed = new HashMap<>();
                Object world = Proxy.newProxyInstance(Class.forName("dds").getClassLoader(),
                    new Class<?>[] {Class.forName("dds")}, (proxy, method, values) -> {
                        String name = method.getName();
                        if (name.equals("I_")) return -64;
                        if (name.equals("f_")) return true;
                        if (name.equals("toString")) return "atlas-forest-rock-probe-world";
                        if (values != null && values.length > 0 && positionClass.isInstance(values[0])) {
                            int x = (int) coordinateX.invoke(values[0]);
                            int y = (int) coordinateY.invoke(values[0]);
                            int z = (int) coordinateZ.invoke(values[0]);
                            String key = x + " " + y + " " + z;
                            if (name.equals("u")) return (!ground || y > 39) && !changed.containsKey(key);
                            if (name.equals("a_")) return changed.getOrDefault(key,
                                ground && y <= 39 ? stone : air);
                            if (name.equals("a") && values.length == 3) {
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
                boolean placed = (boolean) place.invoke(forestRock, context);
                System.out.printf("rock_count %d %d %d %d %d %d %d%n", seed,
                    origin[0], origin[1], origin[2], placed ? 1 : 0, changed.size(), nextLong.invoke(random));
                changed.keySet().stream().sorted(Comparator.comparingInt((String value) ->
                    Integer.parseInt(value.split(" ")[1])).thenComparingInt(value ->
                    Integer.parseInt(value.split(" ")[2])).thenComparingInt(value ->
                    Integer.parseInt(value.split(" ")[0]))).forEach(position ->
                    System.out.printf("rock_block %d %d %d %d %s%n", seed,
                        origin[0], origin[1], origin[2], position));
            }
        }

        Object lookup = Class.forName("jo$a").getMethod("b").invoke(registries);
        Object featureLookup = Class.forName("jn$a").getMethod("b", Class.forName("akq"))
            .invoke(lookup, Class.forName("lu").getField("aQ").get(null));
        Object placedRock = null;
        for (Object featureHolder : ((Stream<?>) Class.forName("jo").getMethod("b")
            .invoke(featureLookup)).toList()) {
            Object key = ((Optional<?>) Class.forName("jm").getMethod("e")
                .invoke(featureHolder)).orElseThrow();
            if (Class.forName("akq").getMethod("a").invoke(key).toString()
                .equals("minecraft:forest_rock")) {
                placedRock = Class.forName("jm").getMethod("a").invoke(featureHolder);
                break;
            }
        }
        if (placedRock == null) throw new IllegalStateException("forest rock missing");
        Object settingsLookup = Class.forName("jn$a").getMethod("b", Class.forName("akq"))
            .invoke(lookup, Class.forName("lu").getField("aO").get(null));
        Object settingsHolder = Class.forName("jn").getMethod("b", Class.forName("akq"))
            .invoke(settingsLookup, Class.forName("dzd").getField("c").get(null));
        Object biomeLookup = Class.forName("jn$a").getMethod("b", Class.forName("akq"))
            .invoke(lookup, Class.forName("lu").getField("aF").get(null));
        Object plains = null;
        for (Object biomeHolder : ((Stream<?>) Class.forName("jo").getMethod("b")
            .invoke(biomeLookup)).toList()) {
            Object key = ((Optional<?>) Class.forName("jm").getMethod("e")
                .invoke(biomeHolder)).orElseThrow();
            if (Class.forName("akq").getMethod("a").invoke(key).toString()
                .equals("minecraft:plains")) plains = biomeHolder;
        }
        Object biomeSource = Class.forName("deh").getConstructor(Class.forName("jm"))
            .newInstance(plains);
        Object generator = Class.forName("dzb").getConstructor(Class.forName("dea"),
            Class.forName("jm")).newInstance(biomeSource, settingsHolder);
        Object heightWorld = Proxy.newProxyInstance(Class.forName("dds").getClassLoader(),
            new Class<?>[] {Class.forName("dds")}, (proxy, method, values) -> {
                if (method.getName().equals("I_")) return -64;
                if (method.getName().equals("J_")) return 384;
                if (method.getName().equals("a") && values != null && values.length == 3)
                    return 40 + (((int) values[1] ^ (int) values[2]) & 1);
                throw new UnsupportedOperationException(method.toString());
            });
        Object placementContext = Class.forName("eiw").getConstructor(Class.forName("dds"),
            Class.forName("duz"), Optional.class)
            .newInstance(heightWorld, generator, Optional.empty());
        Method modifiers = Class.forName("eiv").getMethod("c");
        Method positions = Class.forName("eiy").getMethod("a_", Class.forName("eiw"),
            Class.forName("ayw"), positionClass);
        Method decorationSeed = Class.forName("dzx").getMethod("a", long.class, int.class, int.class);
        Method featureSeed = Class.forName("dzx").getMethod("b", long.class, int.class, int.class);
        for (long seed : seeds) {
            for (int[] chunk : new int[][] {{0, 0}, {-17, 23}}) {
                Object random = Class.forName("dzx").getConstructor(Class.forName("ayw"))
                    .newInstance(Class.forName("dzz").getConstructor(long.class).newInstance(0L));
                long decoration = (long) decorationSeed.invoke(random, seed, chunk[0] * 16,
                    chunk[1] * 16);
                featureSeed.invoke(random, decoration, 4, 2);
                List<?> current = List.of(positionClass.getConstructor(int.class, int.class, int.class)
                    .newInstance(chunk[0] * 16, 0, chunk[1] * 16));
                for (Object modifier : (List<?>) modifiers.invoke(placedRock)) {
                    if (modifier.getClass().getName().equals("eii")) break;
                    ArrayList<Object> next = new ArrayList<>();
                    for (Object position : current)
                        next.addAll(((Stream<?>) positions.invoke(modifier, placementContext,
                            random, position)).toList());
                    current = next;
                }
                System.out.printf("rock_candidates %d %d %d %d %d%n", seed,
                    chunk[0], chunk[1], current.size(), nextLong.invoke(random));
                for (Object position : current)
                    System.out.printf("rock_candidate %d %d %d %d %d %d%n", seed,
                        chunk[0], chunk[1], coordinateX.invoke(position),
                        coordinateY.invoke(position), coordinateZ.invoke(position));
            }
        }
    }
}
