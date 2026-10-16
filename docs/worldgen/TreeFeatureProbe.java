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
import java.util.function.Predicate;
import java.util.stream.Stream;

/** Runs the actual configured tree feature, including leaf-distance updates. */
public final class TreeFeatureProbe {
    public static java.util.Set<String> tagBlocks(String name) throws Exception {
        var input = TreeFeatureProbe.class.getClassLoader().getResourceAsStream("data/minecraft/tags/block/" + name + ".json");
        if (input == null) throw new IllegalStateException("missing tag " + name);
        String text;
        try (input) { text = new String(input.readAllBytes(), StandardCharsets.UTF_8); }
        var values = java.util.regex.Pattern.compile("\"(#?minecraft:[^\"]+)\"").matcher(text);
        var result = new java.util.TreeSet<String>();
        while (values.find()) {
            String value = values.group(1);
            if (value.startsWith("#")) result.addAll(tagBlocks(value.substring(11)));
            else result.add(value);
        }
        return result;
    }

    public static void bindTreeTags() throws Exception {
        Object registry = Class.forName("lt").getField("e").get(null);
        Class<?> registryClass = Class.forName("jz");
        var tags = new HashMap<Object, java.util.List<Object>>();
        for (Object tag : ((Stream<?>)registryClass.getMethod("k").invoke(registry)).toList()) tags.put(tag, java.util.List.of());
        for (String[] entry : new String[][]{{"replaceable_by_trees", "co"}, {"logs", "u"}, {"leaves", "P"}, {"dirt", "ag"}, {"dead_bush_may_place_on", "cm"}, {"sand", "I"}, {"mangrove_logs_can_grow_through", "ck"}, {"mangrove_roots_can_grow_through", "cl"}}) {
            var holders = new java.util.ArrayList<Object>();
            for (String name : tagBlocks(entry[0])) {
                Object key = Class.forName("akr").getMethod("a", String.class).invoke(null, name);
                Object block = registryClass.getMethod("a", Class.forName("akr")).invoke(registry, key);
                holders.add(registryClass.getMethod("e", Object.class).invoke(registry, block));
            }
            tags.put(Class.forName("awe").getField(entry[1]).get(null), holders);
        }
        registryClass.getMethod("a", Map.class).invoke(registry, tags);
        Object fluids=Class.forName("lt").getField("c").get(null);
        var fluidTags=new HashMap<Object,java.util.List<Object>>();
        for(String[] entry:new String[][]{{"water","a"},{"lava","b"}}) {
            var holders=new java.util.ArrayList<Object>();
            for(String name:new String[]{entry[0],"flowing_"+entry[0]}) {
                Object value=registryClass.getMethod("a",Class.forName("akr")).invoke(fluids,Class.forName("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name));
                holders.add(registryClass.getMethod("e",Object.class).invoke(fluids,value));
            }
            fluidTags.put(Class.forName("awk").getField(entry[1]).get(null),holders);
        }
        registryClass.getMethod("a",Map.class).invoke(fluids,fluidTags);
    }

    public static void main(String[] args) throws Exception {
        Object version = Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a", Class.forName("ae")).invoke(null, version);
        Class.forName("akt").getMethod("a").invoke(null);
        bindTreeTags();
        Object registries = Class.forName("pd").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out), true));
        Object lookup = Class.forName("jo$a").getMethod("b").invoke(registries);
        Object featureLookup = Class.forName("jn$a").getMethod("b", Class.forName("akq"))
            .invoke(lookup, Class.forName("lu").getField("aI").get(null));
        Map<String, Object> configurations = new HashMap<>();
        for (Object holder : ((Stream<?>) Class.forName("jo").getMethod("b").invoke(featureLookup)).toList()) {
            Object key = ((Optional<?>) Class.forName("jm").getMethod("e").invoke(holder)).orElseThrow();
            configurations.put(Class.forName("akq").getMethod("a").invoke(key).toString(),
                Class.forName("jm").getMethod("a").invoke(holder));
        }
        Class<?> pos = Class.forName("jd"), block = Class.forName("dfy");
        Object registry = Class.forName("lt").getField("e").get(null);
        Method lookupBlock = Class.forName("jz").getMethod("a", Class.forName("akr"));
        Method location = Class.forName("akr").getMethod("a", String.class);
        Map<String, Object> states = new HashMap<>();
        for (String name : new String[]{"air", "grass_block", "dirt", "podzol", "stone", "water", "vine", "oak_log", "oak_leaves", "mud"}) {
            states.put(name, block.getMethod("o").invoke(lookupBlock.invoke(registry, location.invoke(null, "minecraft:" + name))));
        }
        states.put("flowing_water", Class.forName("dte").getMethod("a",Class.forName("duf"),Comparable.class)
            .invoke(states.get("water"),Class.forName("dts").getField("aP").get(null),Integer.valueOf(1)));
        Method getX = pos.getMethod("u"), getY = pos.getMethod("v"), getZ = pos.getMethod("w");
        Method stateID = block.getMethod("i", Class.forName("dtc"));
        Method fluidState = Class.forName("dtb$a").getMethod("u");
        boolean decorated = args.length > 0 && (args[0].equals("decorated") || args[0].equals("cherry") || args[0].equals("mangrove"));
        String[] names = decorated ? new String[]{"jungle_tree", "jungle_tree_no_vine", "swamp_oak", "fancy_oak_bees", "super_birch_bees", "oak_bees_005", "birch_bees_005"}
            : new String[]{"oak", "birch", "spruce", "pine", "fancy_oak"};
        if (args.length > 0 && args[0].equals("giant")) names = new String[]{"mega_pine", "mega_spruce", "mega_jungle_tree", "jungle_bush", "azalea_tree"};
        if (args.length > 0 && args[0].equals("branching")) names = new String[]{"acacia", "dark_oak"};
        if (args.length > 0 && args[0].equals("cherry")) names = new String[]{"cherry", "cherry_bees_005"};
        if (args.length > 0 && args[0].equals("mangrove")) names = new String[]{"mangrove", "tall_mangrove"};
        for (String name : names) {
            // Dedicated-server WorldLoader loads configured features from
            // RegistryDataLoader JSON resources. In 1.21.1 cherry's serialized
            // corner chance is .25 although the bootstrap constructor is .5.
            // Decode the resource using the actual ConfiguredFeature codec.
            Object configured;
            try (var input = TreeFeatureProbe.class.getClassLoader().getResourceAsStream("data/minecraft/worldgen/configured_feature/" + name + ".json")) {
                if (input == null) throw new IllegalStateException("missing tree resource " + name);
                Object json = Class.forName("com.google.gson.JsonParser").getMethod("parseString", String.class).invoke(null, new String(input.readAllBytes(), StandardCharsets.UTF_8));
                Object codec = Class.forName("ebq").getField("a").get(null);
                Object result = Class.forName("com.mojang.serialization.Decoder").getMethod("parse",Class.forName("com.mojang.serialization.DynamicOps"),Object.class)
                    .invoke(codec,Class.forName("akp").getMethod("a",Class.forName("com.mojang.serialization.DynamicOps"),Class.forName("jo$a")).invoke(null,
                        Class.forName("com.mojang.serialization.JsonOps").getField("INSTANCE").get(null),registries),json);
                configured = Class.forName("com.mojang.serialization.DataResult").getMethod("getOrThrow").invoke(result);
            }
            if (configured == null) throw new IllegalStateException("missing " + name);
            Object feature = Class.forName("ebq").getMethod("b").invoke(configured);
            Object config = Class.forName("ebq").getMethod("c").invoke(configured);
            for (long seed : new long[]{0, 1, -1, 123456789, 2, 3, 42, 987654321}) {
                for (int scenario = 0; scenario < (name.contains("mangrove") ? 19 : 15); scenario++) {
                    int[] origin = scenario == 1 ? new int[]{-1, 41, -16} : scenario == 2 ? new int[]{15, 41, 15}
                        : scenario == 8 ? new int[]{8, 317, 8} : scenario == 9 ? new int[]{8, -64, 8}
                        : scenario == 11 ? new int[]{29999968,41,29999968} : scenario == 12 ? new int[]{-29999968,41,-29999968}
                        : scenario == 13 ? new int[]{8,313,8} : scenario == 14 ? new int[]{8,315,8} : new int[]{8, 41, 8};
                    int mode = scenario;
                    Map<String, Object> changed = new HashMap<>();
                    Object world = Proxy.newProxyInstance(Class.forName("dds").getClassLoader(), new Class<?>[]{Class.forName("dds")},
                        (proxy, method, values) -> {
                            String m = method.getName();
                            if (m.equals("I_")) return -64;
                            if (m.equals("J_")) return 384;
                            if (m.equals("am")) return 320;
                            if (m.equals("f_")) return true;
                            if (m.equals("toString")) return "atlas-tree-probe-world";
                            if (values != null && values.length > 0 && pos.isInstance(values[0])) {
                                int x = (int)getX.invoke(values[0]), y = (int)getY.invoke(values[0]), z = (int)getZ.invoke(values[0]);
                                String key = x + " " + y + " " + z;
                                Object base = y == 40 ? states.get(mode == 3 ? "podzol" : mode == 4 ? "stone" : "grass_block") : states.get("air");
                                if (mode == 5 && x == origin[0] && z == origin[2] && y == 44) base = states.get("stone");
                                if (mode == 6 && y >= 41 && y <= 43) base = states.get("water");
                                if (mode == 10 && y >= 41 && y <= 44) base = states.get("flowing_water");
                                if (mode >= 15 && y == 40) base = states.get("mud");
                                if (mode >= 15 && y <= 39) base = states.get("dirt");
                                if (mode == 16 && y >= 41 && y <= 44) base = states.get("water");
                                if (mode == 17 && y >= 41 && y <= 44) base = states.get("flowing_water");
                                if (mode == 18 && x == origin[0] && z == origin[2] && y == 42) base = states.get("stone");
                                if (mode == 7 && x == origin[0] + 1 && z == origin[2] && y >= 43 && y <= 45) base = states.get(y == 43 ? "oak_log" : "oak_leaves");
                                Object state = changed.getOrDefault(key, base);
                                if (y < -64 || y >= 320) state = states.get("air");
                                if (m.equals("a_")) return state;
                                if (m.equals("a") && values.length == 2 && Class.forName("dqj").isInstance(values[1])) return Optional.empty();
                                if (m.equals("a") && values.length == 2 && values[1] instanceof Predicate<?>) return ((Predicate<Object>)values[1]).test(state);
                                if (m.equals("b") && values.length == 2 && values[1] instanceof Predicate<?>) return ((Predicate<Object>)values[1]).test(fluidState.invoke(state));
                                if (m.equals("a") && values.length == 3 && Class.forName("dtc").isInstance(values[1])) {
                                    if (y < -64 || y >= 320) return false;
                                    changed.put(key, values[1]); return true;
                                }
                                if (method.getReturnType() == void.class) return null; // Scheduled leaf/fluid ticks do not run in worldgen.
                            }
                            if (method.isDefault()) return java.lang.reflect.InvocationHandler.invokeDefault(proxy,method,values);
                            throw new UnsupportedOperationException(method.toString());
                        });
                    Object random = Class.forName("dzx").getConstructor(Class.forName("ayw"))
                        .newInstance(Class.forName("dzz").getConstructor(long.class).newInstance(seed));
                    Object start = pos.getConstructor(int.class, int.class, int.class).newInstance(origin[0], origin[1], origin[2]);
                    if (decorated) {
                        var shuffle = java.util.Collections.class.getDeclaredField("r");
                        shuffle.setAccessible(true);
                        shuffle.set(null, new java.util.Random(seed ^ 0x123456789abcdef0L));
                    }
                    Object context = Class.forName("ecg").getConstructor(Optional.class, Class.forName("dds"), Class.forName("duz"),
                        Class.forName("ayw"), pos, Class.forName("eei")).newInstance(Optional.empty(), world, null, random, start, config);
                    boolean placed = (boolean)Class.forName("ece").getMethod("a", Class.forName("ecg")).invoke(feature, context);
                    java.util.List<String> output = new java.util.ArrayList<>();
                    for (var entry : changed.entrySet()) output.add(entry.getKey() + " " + stateID.invoke(null, entry.getValue()));
                    output.sort(String::compareTo);
                    if (args.length > 0 && name.equals("oak") && seed == 0 && scenario == 0) output.forEach(System.out::println);
                    String hash = HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(String.join("\n", output).getBytes(StandardCharsets.UTF_8)));
                    System.out.printf("tree_case minecraft:%s %d %d %d %d %d %d %d %d %s%n", name, seed, scenario, origin[0], origin[1], origin[2],
                        placed ? 1 : 0, changed.size(), Class.forName("dzx").getMethod("g").invoke(random), hash);
                }
            }
        }
    }
}
