package atlasprobe;

import java.lang.reflect.Method;
import java.lang.reflect.Proxy;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.stream.Stream;

public final class DimensionPlacementProbe {
    public static void main(String[] args) throws Exception {
        boolean streaming = true;
        Object version = Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a", Class.forName("ae")).invoke(null, version);
        Class.forName("akt").getMethod("a").invoke(null);
        Object registries = Class.forName("pd").getMethod("a").invoke(null);
        Object lookup = Class.forName("jo$a").getMethod("b").invoke(registries);
        Object featureLookup = Class.forName("jn$a").getMethod("b", Class.forName("akq"))
            .invoke(lookup, Class.forName("lu").getField("aQ").get(null));
        Map<String, Object> features = new HashMap<>();
        for (Object holder : ((Stream<?>) Class.forName("jo").getMethod("b")
            .invoke(featureLookup)).toList()) {
            Object key = ((Optional<?>) Class.forName("jm").getMethod("e").invoke(holder)).orElseThrow();
            String name = Class.forName("akq").getMethod("a").invoke(key).toString();
            features.put(name, Class.forName("jm").getMethod("a").invoke(holder));
        }

        Object settingsLookup = Class.forName("jn$a").getMethod("b", Class.forName("akq"))
            .invoke(lookup, Class.forName("lu").getField("aO").get(null));
        Object settingsHolder = Class.forName("jn").getMethod("b", Class.forName("akq"))
            .invoke(settingsLookup, Class.forName("dzd").getField("f").get(null));
        Object biomeLookup = Class.forName("jn$a").getMethod("b", Class.forName("akq"))
            .invoke(lookup, Class.forName("lu").getField("aF").get(null));
        Object plains = null;
        for (Object holder : ((Stream<?>) Class.forName("jo").getMethod("b")
            .invoke(biomeLookup)).toList()) {
            Object key = ((Optional<?>) Class.forName("jm").getMethod("e").invoke(holder)).orElseThrow();
            if (Class.forName("akq").getMethod("a").invoke(key).toString()
                .equals("minecraft:plains")) plains = holder;
        }
        if (plains == null) throw new IllegalStateException("plains biome missing");
        Object biomeSource = Class.forName("deh").getConstructor(Class.forName("jm"))
            .newInstance(plains);
        Object generator = Class.forName("dzb").getConstructor(Class.forName("dea"),
            Class.forName("jm")).newInstance(biomeSource, settingsHolder);
        Map<String,Object> states = new HashMap<>();
        for (String name : new String[]{"air","bedrock","netherrack","soul_sand","lava"}) {
            Object key=Class.forName("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name);
            Object block=Class.forName("jz").getMethod("a",Class.forName("akr")).invoke(Class.forName("lt").getField("e").get(null),key);
            states.put(name,Class.forName("dfy").getMethod("o").invoke(block));
        }
        Object world = Proxy.newProxyInstance(Class.forName("dds").getClassLoader(),
            new Class<?>[]{Class.forName("dds")},(proxy,method,values)->{
                if (method.getName().equals("I_")) return 0;
                if (method.getName().equals("J_")) return 256;
                if (method.getName().equals("a") && values.length==3 && values[0].getClass().getName().equals("dyy$a")) return 121;
                if (method.getName().equals("a_") && values.length==1) {
                    int y=(int)Class.forName("jd").getMethod("v").invoke(values[0]);
                    String name=y==0||y>=111&&y<=120 ? "bedrock" : y>0&&y<=20||y>=61&&y<=70 ? "netherrack" : y>=21&&y<=30 ? "lava" : y>=51&&y<=55 ? "soul_sand" : "air";
                    return states.get(name);
                }
                if (method.isDefault()) return java.lang.reflect.InvocationHandler.invokeDefault(proxy,method,values);
                throw new UnsupportedOperationException(method.toString());
            });
        Object variableCount=Class.forName("bqc").getMethod("a",int.class,int.class).invoke(null,0,5);
        Object variableLayer=Class.forName("eim").getMethod("a",Class.forName("bpw")).invoke(null,variableCount);
        Object context = Class.forName("eiw").getConstructor(Class.forName("dds"),
            Class.forName("duz"), Optional.class).newInstance(world, generator, Optional.empty());

        System.setOut(new java.io.PrintStream(new java.io.FileOutputStream(java.io.FileDescriptor.out),true));
        Class<?> randomClass = Class.forName("dzx");
        Object random = randomClass.getConstructor(Class.forName("ayw"))
            .newInstance(Class.forName("dzz").getConstructor(long.class).newInstance(0L));
        Method decorationSeed = randomClass.getMethod("a", long.class, int.class, int.class);
        Method featureSeed = randomClass.getMethod("b", long.class, int.class, int.class);
        Method positions = Class.forName("eiy").getMethod("a_", Class.forName("eiw"),
            Class.forName("ayw"), Class.forName("jd"));
        Method modifiers = Class.forName("eiv").getMethod("c");
        Class<?> blockPos = Class.forName("jd");
        Method x = blockPos.getMethod("u");
        Method y = blockPos.getMethod("v");
        Method z = blockPos.getMethod("w");
        Object[][] selected = {{"minecraft:patch_fire",7,4}, {"minecraft:spring_open",7,3},
            {"minecraft:glowstone",7,7}, {"minecraft:glowstone_extra",7,6},
            {"minecraft:ore_debris_small",7,22}, {"minecraft:warped_fungi",9,3},
            {"minecraft:small_basalt_columns",4,1}, {"minecraft:delta",4,0},
            {"minecraft:end_island_decorated",0,0}, {"atlas:variable_layer",4,0}, {"minecraft:end_platform",10,0}};
        long[] seeds = {0L, 123456789L, -1L};
        int[][] chunks = {{0, 0}, {-17, 23}, {6,0}};
        for (long seed : seeds) {
            for (int[] chunk : chunks) {
                int blockX = chunk[0] * 16;
                int blockZ = chunk[1] * 16;
                long decoration = (long) decorationSeed.invoke(random, seed, blockX, blockZ);
                for (Object[] entry : selected) {
                    String name = (String) entry[0];
                    int step = (int) entry[1];
                    int index = (int) entry[2];
                    featureSeed.invoke(random, decoration, index, step);
                    if (streaming) {
                        Stream<?> current = Stream.of(blockPos.getConstructor(int.class, int.class, int.class)
                            .newInstance(blockX, 0, blockZ));
                        for (Object modifier : name.equals("atlas:variable_layer") ? List.of(variableLayer) : (List<?>) modifiers.invoke(features.get(name))) {
                            if (modifier.getClass().getName().equals("eii")) break;
                            current = current.flatMap(position -> {
                                try { return (Stream<?>) positions.invoke(modifier, context, random, position); }
                                catch (Exception error) { throw new RuntimeException(error); }
                            });
                        }
                        int[] count = {0};
                        current.forEach(position -> {
                            try {
                                System.out.printf("stream_candidate %d %d %d %d %d %s %d %d %d%n", seed,
                                    chunk[0], chunk[1], step, index, name,
                                    x.invoke(position), y.invoke(position), z.invoke(position));
                                int calls = (int) randomClass.getMethod("a", int.class).invoke(random, 7) + 1;
                                for (int call = 0; call < calls; call++) randomClass.getMethod("g").invoke(random);
                                count[0]++;
                            } catch (Exception error) { throw new RuntimeException(error); }
                        });
                        System.out.printf("stream_count %d %d %d %d %d %s %d %d%n", seed,
                            chunk[0], chunk[1], step, index, name, count[0], randomClass.getMethod("g").invoke(random));
                        continue;
                    }
                    List<?> current = List.of(blockPos.getConstructor(int.class, int.class, int.class)
                        .newInstance(blockX, 0, blockZ));
                    for (Object modifier : name.equals("atlas:variable_layer") ? List.of(variableLayer) : (List<?>) modifiers.invoke(features.get(name))) {
                        if (modifier.getClass().getName().equals("eii")) break;
                        java.util.ArrayList<Object> next = new java.util.ArrayList<>();
                        for (Object position : current) {
                            next.addAll(((Stream<?>) positions.invoke(modifier, context, random, position)).toList());
                        }
                        current = next;
                    }
                    System.out.printf("candidate_count %d %d %d %d %d %s %d %d%n", seed,
                        chunk[0], chunk[1], step, index, name, current.size(),
                        randomClass.getMethod("g").invoke(random));
                    for (Object position : current) {
                        System.out.printf("candidate %d %d %d %d %d %s %d %d %d%n", seed,
                            chunk[0], chunk[1], step, index, name,
                            x.invoke(position), y.invoke(position), z.invoke(position));
                    }
                }
            }
        }
    }
}
