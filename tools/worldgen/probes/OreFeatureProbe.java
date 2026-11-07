package atlasprobe;

import com.mojang.serialization.Lifecycle;
import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.lang.reflect.Method;
import java.lang.reflect.Proxy;
import java.util.ArrayList;
import java.util.List;
import java.util.function.Function;
import java.util.stream.Stream;

public final class OreFeatureProbe {
    public static void main(String[] args) throws Exception {
        Object version = Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a", Class.forName("ae")).invoke(null, version);
        Class.forName("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out), true));
        Object biomeRegistry = Class.forName("ju").getConstructor(Class.forName("akq"), Lifecycle.class)
            .newInstance(Class.forName("lu").getField("aF").get(null), Lifecycle.stable());
        Object registries = Class.forName("pd").getMethod("a").invoke(null);
        Object lookup = Class.forName("jo$a").getMethod("b").invoke(registries);
        Object biomeLookup = Class.forName("jn$a").getMethod("b", Class.forName("akq"))
            .invoke(lookup, Class.forName("lu").getField("aF").get(null));
        Method register = Class.forName("jz").getMethod("a", Class.forName("jz"),
            Class.forName("akq"), Object.class);
        for (Object holder : ((Stream<?>) Class.forName("jo").getMethod("b")
            .invoke(biomeLookup)).toList()) {
            Object key = Class.forName("jm$c").getMethod("h").invoke(holder);
            register.invoke(null, biomeRegistry, key, Class.forName("jm").getMethod("a").invoke(holder));
        }
        Object stoneBlock = Class.forName("dga").getField("b").get(null);
        Object stoneState = stoneBlock.getClass().getMethod("o").invoke(stoneBlock);
        Object dirtBlock = Class.forName("lt").getField("e").get(null);
        Object dirt = Class.forName("jz").getMethod("a", Class.forName("akr"))
            .invoke(dirtBlock, Class.forName("akr").getMethod("a", String.class)
                .invoke(null, "minecraft:dirt"));
        Object dirtState = dirt.getClass().getMethod("o").invoke(dirt);
        Object predicate = Class.forName("emv").getConstructor(Class.forName("dfy"))
            .newInstance(stoneBlock);
        Object heights = Class.forName("dcy").getMethod("e", int.class, int.class)
            .invoke(null, -64, 384);
        Object feature = Class.forName("ece").getField("I").get(null);
        Method place = Class.forName("ece").getMethod("a", Class.forName("ecg"));
        Method setState = Class.forName("dvs").getMethod("a", Class.forName("jd"),
            Class.forName("dtc"), boolean.class);
        Method getState = Class.forName("dvs").getMethod("a_", Class.forName("jd"));
        Method nextLong = Class.forName("dzx").getMethod("g");
        long[] seeds = {0L, 1L, -1L, 123456789L};
        int[] sizes = {5, 9, 17};
        for (long seed : seeds) {
            for (int size : sizes) {
                Object chunkPosition = Class.forName("dcd").getConstructor(int.class, int.class)
                    .newInstance(0, 0);
                Object chunk = Class.forName("dvs").getConstructor(Class.forName("dcd"),
                    Class.forName("dvv"), Class.forName("dcy"), Class.forName("jz"),
                    Class.forName("eab")).newInstance(chunkPosition,
                        Class.forName("dvv").getField("a").get(null), heights, biomeRegistry, null);
                for (int y = 0; y < 40; y++) {
                    for (int z = 0; z < 16; z++) {
                        for (int x = 0; x < 16; x++) {
                            Object position = Class.forName("jd")
                                .getConstructor(int.class, int.class, int.class).newInstance(x, y, z);
                            setState.invoke(chunk, position, stoneState, false);
                        }
                    }
                }
                Object world = Proxy.newProxyInstance(Class.forName("dds").getClassLoader(),
                    new Class<?>[] {Class.forName("dds")}, (proxy, method, values) -> {
                        String name = method.getName();
                        if (name.equals("e") && values != null && values.length == 1)
                            return ((int) values[0] + 64) >> 4;
                        if (name.equals("an")) return 24;
                        if (name.equals("a") && values != null && values.length == 2
                            && values[0] instanceof Integer) return chunk;
                        if (name.equals("a") && values != null && values.length == 3)
                            return 320;
                        if (name.equals("d") && values != null && values.length == 1
                            && values[0] instanceof Integer)
                            return (int) values[0] < -64 || (int) values[0] >= 320;
                        if (name.equals("f_") && values != null && values.length == 1) return true;
                        if (name.equals("toString")) return "atlas-ore-probe-world";
                        throw new UnsupportedOperationException(method.toString());
                    });
                Object target = Class.forName("eeq").getMethod("a", Class.forName("enn"),
                    Class.forName("dtc")).invoke(null, predicate, dirtState);
                Object config = Class.forName("eeq").getConstructor(List.class, int.class, float.class)
                    .newInstance(List.of(target), size, 0.0f);
                Object random = Class.forName("dzx").getConstructor(Class.forName("ayw"))
                    .newInstance(Class.forName("dzz").getConstructor(long.class).newInstance(seed));
                Object origin = Class.forName("jd").getConstructor(int.class, int.class, int.class)
                    .newInstance(8, 20, 8);
                Object context = Class.forName("ecg").getConstructor(java.util.Optional.class,
                    Class.forName("dds"), Class.forName("duz"), Class.forName("ayw"),
                    Class.forName("jd"), Class.forName("eei"))
                    .newInstance(java.util.Optional.empty(), world, null, random, origin, config);
                boolean placed = (boolean) place.invoke(feature, context);
                List<String> changed = new ArrayList<>();
                for (int y = 0; y < 40; y++) {
                    for (int z = 0; z < 16; z++) {
                        for (int x = 0; x < 16; x++) {
                            Object position = Class.forName("jd")
                                .getConstructor(int.class, int.class, int.class).newInstance(x, y, z);
                            if (getState.invoke(chunk, position) == dirtState)
                                changed.add(x + " " + y + " " + z);
                        }
                    }
                }
                System.out.printf("ore_count %d %d %d %d %d%n", seed, size,
                    placed ? 1 : 0, changed.size(), nextLong.invoke(random));
                for (String position : changed) System.out.printf("ore_block %d %d %s%n", seed, size, position);
            }
        }
        Object airBlock = Class.forName("dga").getField("a").get(null);
        Object airState = airBlock.getClass().getMethod("o").invoke(airBlock);
        Method mayReplace = Class.forName("ecz").getMethod("a", Class.forName("dtc"),
            Function.class, Class.forName("ayw"), Class.forName("eeq"),
            Class.forName("eeq$a"), Class.forName("jd$a"));
        for (long seed : seeds) {
            for (float chance : new float[] {0.0f, 0.25f, 0.5f, 1.0f}) {
                for (boolean exposed : new boolean[] {false, true}) {
                    Object target = Class.forName("eeq").getMethod("a", Class.forName("enn"),
                        Class.forName("dtc")).invoke(null, predicate, dirtState);
                    Object config = Class.forName("eeq").getConstructor(List.class, int.class, float.class)
                        .newInstance(List.of(target), 9, chance);
                    Object random = Class.forName("dzx").getConstructor(Class.forName("ayw"))
                        .newInstance(Class.forName("dzz").getConstructor(long.class).newInstance(seed));
                    Object position = Class.forName("jd$a").getConstructor(int.class, int.class, int.class)
                        .newInstance(8, exposed ? 39 : 20, 8);
                    Function<Object, Object> neighbor = ignored -> exposed ? airState : stoneState;
                    boolean allowed = (boolean) mayReplace.invoke(null, stoneState, neighbor,
                        random, config, target, position);
                    System.out.printf("ore_check %d %d %d %d %d%n", seed,
                        Float.floatToRawIntBits(chance), exposed ? 1 : 0,
                        allowed ? 1 : 0, nextLong.invoke(random));
                }
            }
        }
    }
}
