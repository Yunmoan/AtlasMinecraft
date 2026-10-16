package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.util.*;
import java.util.function.*;
import java.util.stream.Stream;

/** FeatureSorter schedules from the original Nether and End biome sources. */
public final class DimensionFeatureOrderProbe {
    public static void main(String[] args) throws Exception {
        Object version = Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a", Class.forName("ae")).invoke(null, version);
        Class.forName("akt").getMethod("a").invoke(null);
        Object registries = Class.forName("pd").getMethod("a").invoke(null);
        Object lookup = Class.forName("jo$a").getMethod("b").invoke(registries);
        Method lookupRegistry = Class.forName("jn$a").getMethod("b", Class.forName("akq"));
        Object biomes = lookupRegistry.invoke(lookup, Class.forName("lu").getField("aF").get(null));
        Object placed = lookupRegistry.invoke(lookup, Class.forName("lu").getField("aQ").get(null));
        Object settings = lookupRegistry.invoke(lookup, Class.forName("lu").getField("aO").get(null));
        Map<Object, String> featureNames = new IdentityHashMap<>();
        for (Object holder : ((Stream<?>) Class.forName("jo").getMethod("b").invoke(placed)).toList()) {
            Object key = ((Optional<?>) Class.forName("jm").getMethod("e").invoke(holder)).orElseThrow();
            featureNames.put(Class.forName("jm").getMethod("a").invoke(holder), Class.forName("akq").getMethod("a").invoke(key).toString());
        }
        Object end = Class.forName("den").getMethod("a", Class.forName("jn")).invoke(null, biomes);
        Object preset = Class.forName("dek$a").getField("a").get(null);
        Object provider = Class.forName("dek$a").getMethod("c").invoke(preset);
        Method apply = Class.forName("dek$a$a").getMethod("apply", Function.class);
        apply.setAccessible(true);
        Object points = apply.invoke(provider, (Function<Object,Object>) key -> {
            try { return Class.forName("jn").getMethod("b", Class.forName("akq")).invoke(biomes, key); }
            catch (ReflectiveOperationException e) { throw new RuntimeException(e); }
        });
        Object nether = Class.forName("dej").getMethod("a", Class.forName("def$c")).invoke(null, points);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out), true));
        for (String dimension : new String[]{"nether", "end"}) {
            Object source = dimension.equals("nether") ? nether : end;
            Object key = Class.forName("dzd").getField(dimension.equals("nether") ? "f" : "g").get(null);
            Object holder = Class.forName("jn").getMethod("b", Class.forName("akq")).invoke(settings, key);
            Object generator = Class.forName("dzb").getConstructor(Class.forName("dea"), Class.forName("jm")).newInstance(source, holder);
            Field sorter = Class.forName("duz").getDeclaredField("c");
            sorter.setAccessible(true);
            List<?> steps = (List<?>) ((Supplier<?>) sorter.get(generator)).get();
            for (int step = 0; step < steps.size(); step++) {
                List<?> features = (List<?>) Class.forName("deg$b").getMethod("a").invoke(steps.get(step));
                for (int index = 0; index < features.size(); index++) {
                    String name = featureNames.get(features.get(index));
                    if (name == null) throw new IllegalStateException("unregistered feature");
                    System.out.printf("feature %s %d %d %s%n", dimension, step, index, name);
                }
            }
            @SuppressWarnings("unchecked")
            Set<Object> possible = (Set<Object>) Class.forName("dea").getMethod("c").invoke(source);
            Map<String, Object> biomeHolders = new TreeMap<>();
            for (Object biomeHolder : possible) {
                Object biomeKey = ((Optional<?>) Class.forName("jm").getMethod("e").invoke(biomeHolder)).orElseThrow();
                biomeHolders.put(Class.forName("akq").getMethod("a").invoke(biomeKey).toString(), biomeHolder);
            }
            for (var biomeEntry : biomeHolders.entrySet()) {
                System.out.printf("source %s %s%n", dimension, biomeEntry.getKey());
                Object biome = Class.forName("jm").getMethod("a").invoke(biomeEntry.getValue());
                Object generation = Class.forName("ddw").getMethod("d").invoke(biome);
                List<?> biomeSteps = (List<?>) Class.forName("ddx").getMethod("b").invoke(generation);
                for (int step = 0; step < biomeSteps.size(); step++) {
                    int index = 0;
                    for (Object featureHolder : ((Stream<?>) Class.forName("jq").getMethod("a").invoke(biomeSteps.get(step))).toList()) {
                        Object feature = Class.forName("jm").getMethod("a").invoke(featureHolder);
                        System.out.printf("biome %s %s %d %d %s%n", dimension, biomeEntry.getKey(), step, index++, featureNames.get(feature));
                    }
                }
            }
            List<List<String>> groups = new ArrayList<>();
            for (String name : biomeHolders.keySet()) groups.add(List.of(name));
            groups.add(new ArrayList<>(biomeHolders.keySet()));
            for (List<String> group : groups) {
                String label = String.join(",", group);
                for (int step = 0; step < steps.size(); step++) {
                    List<?> features = (List<?>) Class.forName("deg$b").getMethod("a").invoke(steps.get(step));
                    @SuppressWarnings("unchecked")
                    ToIntFunction<Object> indexOf = (ToIntFunction<Object>) Class.forName("deg$b").getMethod("b").invoke(steps.get(step));
                    Set<Integer> selected = new TreeSet<>();
                    for (String name : group) {
                        Object biome = Class.forName("jm").getMethod("a").invoke(biomeHolders.get(name));
                        Object generation = Class.forName("ddw").getMethod("d").invoke(biome);
                        List<?> biomeSteps = (List<?>) Class.forName("ddx").getMethod("b").invoke(generation);
                        if (step >= biomeSteps.size()) continue;
                        for (Object featureHolder : ((Stream<?>) Class.forName("jq").getMethod("a").invoke(biomeSteps.get(step))).toList())
                            selected.add(indexOf.applyAsInt(Class.forName("jm").getMethod("a").invoke(featureHolder)));
                    }
                    for (int index : selected)
                        System.out.printf("schedule %s %s %d %d %s%n", dimension, label, step, index, featureNames.get(features.get(index)));
                }
            }
        }
    }
}
