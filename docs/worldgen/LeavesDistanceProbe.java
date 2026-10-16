package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.lang.reflect.Method;
import java.lang.reflect.Proxy;
import java.util.LinkedHashMap;
import java.util.Map;
import java.util.ArrayList;
import java.util.List;
import java.util.HashMap;
import java.util.HashSet;
import java.util.Set;
import java.util.stream.Stream;
import java.util.regex.Pattern;
import java.nio.charset.StandardCharsets;

// Execute the actual private LeavesBlock scheduled-tick distance rule on
// overlapping trees and authored leaf states, then settle its repeated ticks.
public final class LeavesDistanceProbe {
    record Position(int x, int y, int z) {}
    static Class<?> stateClass, positionClass;
    static Method stateID, blockAt, update, propertySet;
    static Object blockRegistry, distance, persistent, waterlogged;

    static Set<String> tagBlocks(String name) throws Exception {
        try (var input = LeavesDistanceProbe.class.getClassLoader().getResourceAsStream("data/minecraft/tags/block/" + name + ".json")) {
            if (input == null) throw new IllegalStateException("missing tag " + name);
            var values = Pattern.compile("\"([^\"]+)\"").matcher(new String(input.readAllBytes(), StandardCharsets.UTF_8));
            Set<String> result = new HashSet<>();
            while (values.find()) {
                String value = values.group(1);
                if (value.startsWith("#minecraft:")) result.addAll(tagBlocks(value.substring(11)));
                else if (value.startsWith("minecraft:")) result.add(value);
            }
            return result;
        }
    }

    static void bindLogs() throws Exception {
        Class<?> registry = Class.forName("jz");
        Map<Object, List<Object>> tags = new HashMap<>();
        for (Object tag : ((Stream<?>) registry.getMethod("k").invoke(blockRegistry)).toList()) tags.put(tag, List.of());
        List<Object> holders = new ArrayList<>();
        for (String name : tagBlocks("logs")) {
            Object location = Class.forName("akr").getMethod("a", String.class).invoke(null, name);
            Object block = registry.getMethod("a", Class.forName("akr")).invoke(blockRegistry, location);
            holders.add(registry.getMethod("e", Object.class).invoke(blockRegistry, block));
        }
        tags.put(Class.forName("awe").getField("u").get(null), holders);
        registry.getMethod("a", Map.class).invoke(blockRegistry, tags);
        var optional = (java.util.OptionalInt) Class.forName("dki").getMethod("n", stateClass).invoke(null, state("stripped_oak_log", 0, false, false));
        if (optional.orElse(7) != 0) throw new IllegalStateException("native logs tag was not bound");
    }

    static Object state(String name, int d, boolean p, boolean w) throws Exception {
        Object location = Class.forName("akr").getMethod("a", String.class).invoke(null, "minecraft:" + name);
        Object block = Class.forName("jz").getMethod("a", Class.forName("akr")).invoke(blockRegistry, location);
        Object result = Class.forName("dfy").getMethod("o").invoke(block);
        if (name.endsWith("leaves")) {
            result = propertySet.invoke(result, distance, d);
            result = propertySet.invoke(result, persistent, p);
            result = propertySet.invoke(result, waterlogged, w);
        }
        return result;
    }

    public static void main(String[] args) throws Exception {
        Object version = Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a", Class.forName("ae")).invoke(null, version);
        Class.forName("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out), true));
        stateClass = Class.forName("dtc");
        positionClass = Class.forName("jd");
        stateID = Class.forName("dfy").getMethod("i", stateClass);
        propertySet = Class.forName("dte").getMethod("a", Class.forName("duf"), Comparable.class);
        blockRegistry = Class.forName("lt").getField("e").get(null);
        distance = Class.forName("dki").getField("d").get(null);
        persistent = Class.forName("dki").getField("e").get(null);
        waterlogged = Class.forName("dki").getField("f").get(null);
        bindLogs();
        update = Class.forName("dki").getDeclaredMethod("a", stateClass, Class.forName("dcx"), positionClass);
        update.setAccessible(true);
        Object air = state("air", 7, false, false);
        String[] species = {"oak_leaves", "birch_leaves", "spruce_leaves", "mangrove_leaves", "azalea_leaves", "flowering_azalea_leaves"};
        for (int scenario = 0; scenario < 4; scenario++) {
            Map<Position, Object> states = new LinkedHashMap<>();
            int baseX = scenario == 2 ? -2 : 14;
            if (scenario != 3) states.put(new Position(baseX-1, 64, 8), state("stripped_oak_log", 0, false, false));
            if (scenario == 0 || scenario == 2) states.put(new Position(baseX+6, 64, 9), state("birch_log", 0, false, false));
            if (scenario == 3) states.put(new Position(baseX-1, 64, 8), state("scaffolding", 0, false, false));
            for (int i = 0; i < 8; i++) {
                states.put(new Position(baseX+i, 64, 8), state(species[i % species.length], 1+i%7, i%2==0, i%3==0));
            }
            // Add a vertical bend, ensuring the native six-neighbor rule is
            // checked beyond a straight horizontal chain.
            for (int y = 65; y <= 68; y++) {
                states.put(new Position(baseX+2, y, 8), state("oak_leaves", 2, true, true));
            }
            Map<Position, Object> initial = new LinkedHashMap<>(states);
            Object world = Proxy.newProxyInstance(Class.forName("dcx").getClassLoader(), new Class<?>[]{Class.forName("dcx")},
                (proxy, method, values) -> {
                    if (method.getName().equals("a_") && values != null && values.length == 1) {
                        Object p = values[0];
                        return states.getOrDefault(new Position(
                            (Integer) positionClass.getMethod("u").invoke(p),
                            (Integer) positionClass.getMethod("v").invoke(p),
                            (Integer) positionClass.getMethod("w").invoke(p)), air);
                    }
                    throw new UnsupportedOperationException(method.toString());
                });
            int passes = 0;
            for (; passes < 64; passes++) {
                boolean changed = false;
                List<Map.Entry<Position, Object>> entries = new ArrayList<>(states.entrySet());
                for (var entry : entries) {
                    Object s = entry.getValue();
                    if (!(Boolean) stateClass.getMethod("b", Class.forName("duf")).invoke(s, distance)) continue;
                    Position p = entry.getKey();
                    Object position = positionClass.getConstructor(int.class, int.class, int.class).newInstance(p.x, p.y, p.z);
                    Object updated = update.invoke(null, s, world, position);
                    if (updated != s) {
                        states.put(p, updated);
                        changed = true;
                    }
                }
                if (!changed) break;
            }
            if (passes == 64) throw new IllegalStateException("leaf updates did not settle");
            System.out.printf("leaf_case %d %d %d%n", scenario, passes, states.size());
            for (var entry : initial.entrySet()) {
                Position p = entry.getKey();
                System.out.printf("leaf_state %d %d %d %d %d %d%n", scenario, p.x, p.y, p.z,
                    stateID.invoke(null, entry.getValue()), stateID.invoke(null, states.get(p)));
            }
        }
    }
}
