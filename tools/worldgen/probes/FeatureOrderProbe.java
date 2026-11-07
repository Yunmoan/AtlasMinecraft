package atlasprobe;

import com.mojang.datafixers.util.Pair;
import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.lang.reflect.Field;
import java.lang.reflect.Method;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.HashMap;
import java.util.IdentityHashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.Set;
import java.util.TreeSet;
import java.util.function.Consumer;
import java.util.function.Supplier;
import java.util.function.ToIntFunction;
import java.util.stream.Stream;

public final class FeatureOrderProbe {
    public static void main(String[] args) throws Exception {
        Object version = Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a", Class.forName("ae")).invoke(null, version);
        Class.forName("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out), true));
        Object registries = Class.forName("pd").getMethod("a").invoke(null);
        Object lookup = Class.forName("jo$a").getMethod("b").invoke(registries);
        Method lookupRegistry = Class.forName("jn$a").getMethod("b", Class.forName("akq"));
        Object biomeLookup = lookupRegistry.invoke(lookup, Class.forName("lu").getField("aF").get(null));
        Map<String, Object> biomeHolders = new HashMap<>();
        for (Object holder : ((Stream<?>) Class.forName("jo").getMethod("b").invoke(biomeLookup)).toList()) {
            Object key = ((Optional<?>) Class.forName("jm").getMethod("e").invoke(holder)).orElseThrow();
            biomeHolders.put(Class.forName("akq").getMethod("a").invoke(key).toString(), holder);
        }
        Object featureLookup = lookupRegistry.invoke(lookup, Class.forName("lu").getField("aQ").get(null));
        Map<Object, String> featureNames = new IdentityHashMap<>();
        for (Object holder : ((Stream<?>) Class.forName("jo").getMethod("b").invoke(featureLookup)).toList()) {
            Object key = ((Optional<?>) Class.forName("jm").getMethod("e").invoke(holder)).orElseThrow();
            featureNames.put(Class.forName("jm").getMethod("a").invoke(holder),
                Class.forName("akq").getMethod("a").invoke(key).toString());
        }
        List<Pair<Object, Object>> points = new ArrayList<>();
        Method addPoints = Class.forName("dem").getDeclaredMethod("a", Consumer.class);
        addPoints.setAccessible(true);
        addPoints.invoke(Class.forName("dem").getConstructor().newInstance(),
            (Consumer<Pair<Object, Object>>) pair -> {
                try {
                    String name = Class.forName("akq").getMethod("a").invoke(pair.getSecond()).toString();
                    points.add(Pair.of(pair.getFirst(), biomeHolders.get(name)));
                } catch (ReflectiveOperationException error) {
                    throw new RuntimeException(error);
                }
            });
        Object parameterList = Class.forName("def$c").getConstructor(List.class).newInstance(points);
        Object biomeSource = Class.forName("dej").getMethod("a", Class.forName("def$c"))
            .invoke(null, parameterList);
        Object settingsLookup = lookupRegistry.invoke(lookup, Class.forName("lu").getField("aO").get(null));
        Object settingsKey = Class.forName("dzd").getField("c").get(null);
        Object settingsHolder = Class.forName("jn").getMethod("b", Class.forName("akq"))
            .invoke(settingsLookup, settingsKey);
        Object generator = Class.forName("dzb").getConstructor(Class.forName("dea"), Class.forName("jm"))
            .newInstance(biomeSource, settingsHolder);
        Field sorter = Class.forName("duz").getDeclaredField("c");
        sorter.setAccessible(true);
        List<?> steps = (List<?>) ((Supplier<?>) sorter.get(generator)).get();
        for (int step = 0; step < steps.size(); step++) {
            List<?> features = (List<?>) Class.forName("deg$b").getMethod("a").invoke(steps.get(step));
            for (int index = 0; index < features.size(); index++) {
                String name = featureNames.get(features.get(index));
                if (name == null) throw new IllegalStateException("unregistered global feature");
                System.out.printf("feature %d %d %s%n", step, index, name);
            }
        }
        List<String> biomeNames = new ArrayList<>(biomeHolders.keySet());
        biomeNames.sort(Comparator.naturalOrder());
        for (String biomeName : biomeNames) {
            Object biome = Class.forName("jm").getMethod("a").invoke(biomeHolders.get(biomeName));
            Object generation = Class.forName("ddw").getMethod("d").invoke(biome);
            List<?> biomeSteps = (List<?>) Class.forName("ddx").getMethod("b").invoke(generation);
            for (int step = 0; step < biomeSteps.size(); step++) {
                int index = 0;
                for (Object holder : ((Stream<?>) Class.forName("jq").getMethod("a").invoke(biomeSteps.get(step))).toList()) {
                    Object feature = Class.forName("jm").getMethod("a").invoke(holder);
                    String name = featureNames.get(feature);
                    if (name == null) throw new IllegalStateException("unregistered biome feature");
                    System.out.printf("biome %s %d %d %s%n", biomeName, step, index++, name);
                }
            }
        }
        String[][] groups = {{"plains"}, {"desert"}, {"plains", "desert"},
            {"frozen_ocean", "deep_dark"}, {"savanna", "jungle", "meadow"}};
        for (String[] group : groups) {
            String label = String.join("_", group);
            for (int step = 0; step < steps.size(); step++) {
                List<?> features = (List<?>) Class.forName("deg$b").getMethod("a").invoke(steps.get(step));
                @SuppressWarnings("unchecked")
                ToIntFunction<Object> indexOf = (ToIntFunction<Object>) Class.forName("deg$b")
                    .getMethod("b").invoke(steps.get(step));
                Set<Integer> selected = new TreeSet<>();
                for (String biomeName : group) {
                    Object biome = Class.forName("jm").getMethod("a")
                        .invoke(biomeHolders.get("minecraft:" + biomeName));
                    Object generation = Class.forName("ddw").getMethod("d").invoke(biome);
                    List<?> biomeSteps = (List<?>) Class.forName("ddx").getMethod("b").invoke(generation);
                    if (step >= biomeSteps.size()) continue;
                    for (Object holder : ((Stream<?>) Class.forName("jq").getMethod("a")
                        .invoke(biomeSteps.get(step))).toList()) {
                        selected.add(indexOf.applyAsInt(Class.forName("jm").getMethod("a").invoke(holder)));
                    }
                }
                for (int index : selected) {
                    System.out.printf("schedule %s %d %d %s%n", label, step, index,
                        featureNames.get(features.get(index)));
                }
            }
        }
    }
}
