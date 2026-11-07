package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.lang.reflect.Method;
import java.lang.reflect.Proxy;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.*;
import java.util.function.Predicate;
import java.util.stream.Stream;

/** Lazy top-level placement plus real nested configured/placed tree features. */
public final class TreePlacementProbe {
    public static void main(String[] args) throws Exception {
        Object version = Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a", Class.forName("ae")).invoke(null, version);
        Class.forName("akt").getMethod("a").invoke(null);
        TreeFeatureProbe.bindTreeTags();
        Object registries = Class.forName("pd").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Object lookup = Class.forName("jo$a").getMethod("b").invoke(registries);
        Object settingsLookup = Class.forName("jn$a").getMethod("b",Class.forName("akq"))
            .invoke(lookup,Class.forName("lu").getField("aO").get(null));
        Object settings = Class.forName("jn").getMethod("b",Class.forName("akq"))
            .invoke(settingsLookup,Class.forName("dzd").getField("c").get(null));
        Object biomeLookup = Class.forName("jn$a").getMethod("b",Class.forName("akq"))
            .invoke(lookup,Class.forName("lu").getField("aF").get(null));
        Object biome = ((Stream<?>)Class.forName("jo").getMethod("b").invoke(biomeLookup)).findFirst().orElseThrow();
        Object source = Class.forName("deh").getConstructor(Class.forName("jm")).newInstance(biome);
        Object generator = Class.forName("dzb").getConstructor(Class.forName("dea"),Class.forName("jm")).newInstance(source,settings);
        Object placedLookup = Class.forName("jn$a").getMethod("b",Class.forName("akq"))
            .invoke(lookup,Class.forName("lu").getField("aQ").get(null));
        Map<String,Object> features = new HashMap<>();
        for (Object holder : ((Stream<?>)Class.forName("jo").getMethod("b").invoke(placedLookup)).toList()) {
            Object key = ((Optional<?>)Class.forName("jm").getMethod("e").invoke(holder)).orElseThrow();
            features.put(Class.forName("akq").getMethod("a").invoke(key).toString(),Class.forName("jm").getMethod("a").invoke(holder));
        }
        Class<?> pos = Class.forName("jd"), state = Class.forName("dtc");
        Method x = pos.getMethod("u"), y = pos.getMethod("v"), z = pos.getMethod("w");
        Method isAir = Class.forName("dtb$a").getMethod("i"), solid = Class.forName("dtb$a").getMethod("d"), fluid = Class.forName("dtb$a").getMethod("u");
        Method stateID = Class.forName("dfy").getMethod("i",state);
        Map<String,Object> blocks = new HashMap<>();
        Object registry = Class.forName("lt").getField("e").get(null);
        for (String name : new String[]{"air","dirt","grass_block","water","stone","snow_block","powder_snow"}) {
            Object block = Class.forName("jz").getMethod("a",Class.forName("akr"))
                .invoke(registry,Class.forName("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name));
            blocks.put(name,Class.forName("dfy").getMethod("o").invoke(block));
        }
        Object[][] selected = {{"trees_taiga",41},{"trees_grove",42},{"trees_snowy",44},{"trees_water",48},
            {"trees_windswept_forest",37},{"trees_windswept_hills",43},{"trees_badlands",5},
            {"pine_on_snow",-1},{"spruce_on_snow",-1}};
        for (Object[] entry : selected) {
            String name = "minecraft:"+entry[0]; int index = (int)entry[1]; Object placed = features.get(name);
            Object configured = Class.forName("jm").getMethod("a").invoke(Class.forName("eiv").getMethod("b").invoke(placed));
            for (long seed : new long[]{0,1,-1,123456789,2,3,42,987654321}) {
                for (int[] chunk : new int[][]{{0,0},{-17,23}}) {
                    for (int scenario = 0; scenario < 7; scenario++) {
                        int mode = scenario;
                        Map<String,Object> changed = new HashMap<>();
                        java.util.function.Function<int[],Object> read = p -> {
                            String key = p[0]+" "+p[1]+" "+p[2];
                            if (changed.containsKey(key)) return changed.get(key);
                            int floor = mode == 1 ? 38 + Math.floorMod(p[0]+p[2],5) : 40;
                            if (p[1] < floor) return blocks.get("dirt");
                            if (p[1] == floor) return blocks.get(mode == 3 ? "stone" : mode >= 4 ? "snow_block" : "grass_block");
                            if (mode == 2 && p[1] <= 42) return blocks.get("water");
                            if (mode == 4 && p[1] <= 41 + Math.floorMod(p[0]+p[2],3)) return blocks.get("powder_snow");
                            if (mode == 6 && p[1] <= 51 + Math.floorMod(p[0]+p[2],3)) return blocks.get("powder_snow");
                            return blocks.get("air");
                        };
                        Object world = Proxy.newProxyInstance(Class.forName("dds").getClassLoader(),new Class<?>[]{Class.forName("dds")},(proxy,method,values) -> {
                            String m = method.getName();
                            if (m.equals("I_")) return -64;
                            if (m.equals("J_")) return 384;
                            if (m.equals("am")) return 320;
                            if (m.equals("d") && values.length == 1 && values[0] instanceof Integer) return (int)values[0] < -64 || (int)values[0] >= 320;
                            if (m.equals("f_")) return true;
                            if (m.equals("toString")) return "atlas-tree-placement-world";
                            if (m.equals("a") && values.length == 3 && Class.forName("dyy$a").isInstance(values[0])) {
                                boolean oceanFloor = values[0] == Class.forName("dyy$a").getField("d").get(null);
                                for (int height = 319; height >= -64; height--) {
                                    Object s = read.apply(new int[]{(int)values[1],height,(int)values[2]});
                                    if (oceanFloor ? (boolean)solid.invoke(s) : !(boolean)isAir.invoke(s)) return height+1;
                                }
                                return -64;
                            }
                            if (values != null && values.length > 0 && pos.isInstance(values[0])) {
                                int[] p = {(int)x.invoke(values[0]),(int)y.invoke(values[0]),(int)z.invoke(values[0])};
                                Object s = read.apply(p);
                                if (m.equals("a_")) return s;
                                if (m.equals("a") && values.length == 2 && values[1] instanceof Predicate<?>) return ((Predicate<Object>)values[1]).test(s);
                                if (m.equals("b") && values.length == 2 && values[1] instanceof Predicate<?>) return ((Predicate<Object>)values[1]).test(fluid.invoke(s));
                                if (m.equals("a") && values.length == 3 && state.isInstance(values[1])) { changed.put(p[0]+" "+p[1]+" "+p[2],values[1]); return true; }
                                if (method.getReturnType() == void.class) return null;
                            }
                            throw new UnsupportedOperationException(method.toString());
                        });
                        Object random = Class.forName("dzx").getConstructor(Class.forName("ayw"))
                            .newInstance(Class.forName("dzz").getConstructor(long.class).newInstance(index < 0 ? seed : 0));
                        if (index >= 0) {
                            long decoration = (long)Class.forName("dzx").getMethod("a",long.class,int.class,int.class).invoke(random,seed,chunk[0]*16,chunk[1]*16);
                            Class.forName("dzx").getMethod("b",long.class,int.class,int.class).invoke(random,decoration,index,9);
                        }
                        Object context = Class.forName("eiw").getConstructor(Class.forName("dds"),Class.forName("duz"),Optional.class).newInstance(world,generator,Optional.empty());
                        Stream<?> positions = Stream.of(pos.getConstructor(int.class,int.class,int.class).newInstance(chunk[0]*16,index < 0 ? 41 : 0,chunk[1]*16));
                        Method positionsMethod = Class.forName("eiy").getMethod("a_",Class.forName("eiw"),Class.forName("ayw"),pos);
                        for (Object modifier : (List<?>)Class.forName("eiv").getMethod("c").invoke(placed)) {
                            if (modifier.getClass().getName().equals("eii")) continue; // Fixed accepted biome isolates placement/trees.
                            positions = positions.flatMap(p -> { try { return (Stream<?>)positionsMethod.invoke(modifier,context,random,p); } catch(Exception e) { throw new RuntimeException(e); } });
                        }
                        int[] count = {0,0};
                        positions.forEach(p -> {
                            try { count[0]++; if ((boolean)Class.forName("ebq").getMethod("a",Class.forName("dds"),Class.forName("duz"),Class.forName("ayw"),pos).invoke(configured,world,generator,random,p)) count[1]++; }
                            catch(Exception e) { throw new RuntimeException(e); }
                        });
                        List<String> output = new ArrayList<>();
                        for (var change : changed.entrySet()) output.add(change.getKey()+" "+stateID.invoke(null,change.getValue()));
                        output.sort(String::compareTo);
                        String hash = HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(String.join("\n",output).getBytes(StandardCharsets.UTF_8)));
                        System.out.printf("tree_placement %s %d %d %d %d %d %d %d %d %s%n",name,seed,chunk[0],chunk[1],scenario,count[0],count[1],changed.size(),Class.forName("dzx").getMethod("g").invoke(random),hash);
                    }
                }
            }
        }
    }
}
