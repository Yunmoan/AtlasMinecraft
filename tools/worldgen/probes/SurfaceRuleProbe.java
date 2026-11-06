package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.lang.reflect.Field;
import java.lang.reflect.Method;
import java.lang.reflect.Proxy;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;
import java.util.Optional;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.function.Function;
import java.util.stream.Stream;

public final class SurfaceRuleProbe {
    public static void main(String[] args) throws Exception {
        Object version = Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a", Class.forName("ae")).invoke(null, version);
        Class.forName("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out), true));
        Object registries = Class.forName("pd").getMethod("a").invoke(null);
        Object lookup = Class.forName("jo$a").getMethod("b").invoke(registries);
        Method lookupRegistry = Class.forName("jn$a").getMethod("b", Class.forName("akq"));
        Object settingsKey = Class.forName("dzd").getField("c").get(null);
        Object settingsRegistry = lookupRegistry.invoke(lookup, Class.forName("lu").getField("aO").get(null));
        Object settingsHolder = Class.forName("jn").getMethod("b", Class.forName("akq")).invoke(settingsRegistry, settingsKey);
        Object settings = Class.forName("jm").getMethod("a").invoke(settingsHolder);
        Object noiseSettings = Class.forName("dzd").getMethod("f").invoke(settings);
        Object ruleSource = Class.forName("dzd").getMethod("j").invoke(settings);
        Object biomeRegistry = lookupRegistry.invoke(lookup, Class.forName("lu").getField("aF").get(null));
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
        Object biomeSource = Class.forName("deh").getConstructor(Class.forName("jm")).newInstance(biomes.get(0));
        Object generator = Class.forName("dzb").getConstructor(Class.forName("dea"), Class.forName("jm"))
            .newInstance(biomeSource, settingsHolder);
        Object heights = Class.forName("dcy").getMethod("e", int.class, int.class).invoke(null, -64, 384);
        Object generationContext = Class.forName("dzv").getConstructor(Class.forName("duz"), Class.forName("dcy"))
            .newInstance(generator, heights);
        Method fluidFactory = Class.forName("dzb").getDeclaredMethod("a", Class.forName("dzd"));
        fluidFactory.setAccessible(true);
        Object fluidPicker = fluidFactory.invoke(null, settings);
        Object beardifier = Class.forName("dys$b").getField("a").get(null);
        Object blender = Class.forName("eaa").getMethod("a").invoke(null);
        Class<?> contextClass = Class.forName("dzp$g");
        var contextConstructor = contextClass.getDeclaredConstructor(Class.forName("dzq"), Class.forName("dzm"),
            Class.forName("duy"), Class.forName("dzc"), Function.class, Class.forName("jz"), Class.forName("dzv"));
        contextConstructor.setAccessible(true);
        Method updateXZ = contextClass.getDeclaredMethod("a", int.class, int.class);
        Method updateY = contextClass.getDeclaredMethod("a", int.class, int.class, int.class, int.class, int.class, int.class);
        Method secondaryDepth = contextClass.getDeclaredMethod("a");
        Method minSurface = contextClass.getDeclaredMethod("b");
        for (Method method : new Method[] {updateXZ, updateY, secondaryDepth, minSurface}) method.setAccessible(true);
        Field depth = contextClass.getDeclaredField("t");
        Field cold = contextClass.getDeclaredField("f");
        Field steep = contextClass.getDeclaredField("g");
        for (Field field : new Field[] {depth, cold, steep}) field.setAccessible(true);
        Method conditionValue = Class.forName("dzp$e").getDeclaredMethod("b");
        conditionValue.setAccessible(true);
        Method tryApply = Class.forName("dzp$u").getDeclaredMethod("tryApply", int.class, int.class, int.class);
        tryApply.setAccessible(true);
        Method bandAt = Class.forName("dzq").getDeclaredMethod("a", int.class, int.class, int.class);
        bandAt.setAccessible(true);
        Method blockID = Class.forName("dfy").getMethod("i", Class.forName("dtc"));
        int dry = Integer.MIN_VALUE;
        int[][] cases = {{-64, 1, 1, dry}, {-63, 1, 1, dry}, {-62, 1, 1, dry}, {-60, 1, 1, dry},
            {-59, 1, 1, dry}, {0, 1, 1, dry}, {1, 1, 1, dry}, {3, 1, 1, dry}, {7, 1, 1, dry},
            {8, 1, 1, dry}, {40, 1, 8, 63}, {62, 1, 1, 63}, {63, 1, 1, 63}, {64, 1, 1, dry},
            {64, 4, 8, dry}, {95, 1, 1, dry}, {97, 1, 1, dry}, {100, 1, 1, dry}, {200, 1, 1, dry}};
        long[] seeds = {0L, 1L, -1L, 7122209284156038298L};
        int[][] points = {{0, 0}, {173, -421}, {-10000, 24567}};
        for (long seed : seeds) {
            Object randomState = Class.forName("dzm").getMethod("a", Class.forName("jn$a"), Class.forName("akq"), long.class)
                .invoke(null, lookup, settingsKey, seed);
            Object surface = Class.forName("dzm").getMethod("c").invoke(randomState);
            for (int[] point : points) {
                int blockX = point[0], blockZ = point[1];
                Object noiseChunk = Class.forName("dzc").getConstructor(int.class, Class.forName("dzm"), int.class,
                    int.class, Class.forName("dzg"), Class.forName("dys$c"), Class.forName("dzd"),
                    Class.forName("dyk$a"), Class.forName("eaa"))
                    .newInstance(4, randomState, Math.floorDiv(blockX, 16) * 16, Math.floorDiv(blockZ, 16) * 16,
                        noiseSettings, beardifier, settings, fluidPicker, blender);
                for (int blockY = -64; blockY < 320; blockY++) {
                    Object state = bandAt.invoke(surface, blockX, blockY, blockZ);
                    System.out.printf("band %d %d %d %d %d%n", seed, blockX, blockY, blockZ, blockID.invoke(null, state));
                }
                for (Object biome : biomes) {
                    Object biomeKey = ((Optional<?>) Class.forName("jm").getMethod("e").invoke(biome)).orElseThrow();
                    String biomeName = Class.forName("akq").getMethod("a").invoke(biomeKey).toString();
                    Function<Object, Object> biomeAt = ignored -> biome;
                    Object context = contextConstructor.newInstance(surface, randomState, null, noiseChunk,
                        biomeAt, null, generationContext);
                    AtomicBoolean steepValue = new AtomicBoolean();
                    Object steepCondition = Proxy.newProxyInstance(Class.forName("dzp$e").getClassLoader(),
                        new Class<?>[] {Class.forName("dzp$e")}, (proxy, method, arguments) -> steepValue.get());
                    steep.set(context, steepCondition);
                    Object rule = Function.class.getMethod("apply", Object.class).invoke(ruleSource, context);
                    updateXZ.invoke(context, blockX, blockZ);
                    int surfaceDepth = (Integer) depth.get(context);
                    long secondaryBits = Double.doubleToRawLongBits((Double) secondaryDepth.invoke(context));
                    int minimum = (Integer) minSurface.invoke(context);
                    for (int caseIndex = 0; caseIndex < cases.length; caseIndex++) {
                        int[] sample = cases[caseIndex];
                        int blockY = sample[0];
                        steepValue.set(caseIndex % 2 != 0);
                        updateY.invoke(context, sample[1], sample[2], sample[3], blockX, blockY, blockZ);
                        Object state = tryApply.invoke(rule, blockX, blockY, blockZ);
                        int id = state == null ? -1 : (Integer) blockID.invoke(null, state);
                        boolean coldValue = (Boolean) conditionValue.invoke(cold.get(context));
                        System.out.printf("rule %d %d %d %d %s %d %s %d %d %d %d %b %b %d%n",
                            seed, blockX, blockY, blockZ, biomeName, surfaceDepth, Long.toUnsignedString(secondaryBits),
                            sample[1], sample[2], sample[3], minimum, coldValue, steepValue.get(), id);
                    }
                }
            }
        }
    }
}
