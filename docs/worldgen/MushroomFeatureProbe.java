package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.*;
import java.util.function.Predicate;
import java.util.stream.Stream;

/** Real configured patches, nested placements, providers and lazy top-level placements. */
public final class MushroomFeatureProbe {
    public static void main(String[] args) throws Exception {
        Object version = Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a", Class.forName("ae")).invoke(null, version);
        Class.forName("akt").getMethod("a").invoke(null);
        TreeFeatureProbe.bindTreeTags();
        Object blockRegistry=Class.forName("lt").getField("e").get(null);
        Map<Object,List<Object>> tagMap=new HashMap<>();
        for(Object tag:((Stream<?>)Class.forName("jz").getMethod("k").invoke(blockRegistry)).toList()) {
            List<Object> values=new ArrayList<>();
            Object current=Class.forName("jz").getMethod("b",Class.forName("awu")).invoke(blockRegistry,tag);
            if(current instanceof Optional<?> option&&option.isPresent())for(Object holder:(Iterable<?>)option.get())values.add(holder);
            tagMap.put(tag,values);
        }
        List<Object> growHolders=new ArrayList<>();
        for(String name:new String[]{"podzol","mycelium","crimson_nylium","warped_nylium"}) {
            Object block=Class.forName("jz").getMethod("a",Class.forName("akr")).invoke(blockRegistry,Class.forName("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name));
            growHolders.add(Class.forName("jz").getMethod("e",Object.class).invoke(blockRegistry,block));
        }
        tagMap.put(Class.forName("awe").getField("bb").get(null),growHolders);
        Class.forName("jz").getMethod("a",Map.class).invoke(blockRegistry,tagMap);
        Object registries = Class.forName("pd").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Object lookup = Class.forName("jo$a").getMethod("b").invoke(registries);
        Map<String,Object> configured = registry(lookup,"aI"), placed = registry(lookup,"aQ");
        Object settingsLookup = Class.forName("jn$a").getMethod("b",Class.forName("akq")).invoke(lookup,Class.forName("lu").getField("aO").get(null));
        Object settings = Class.forName("jn").getMethod("b",Class.forName("akq")).invoke(settingsLookup,Class.forName("dzd").getField("c").get(null));
        Object biomeLookup = Class.forName("jn$a").getMethod("b",Class.forName("akq")).invoke(lookup,Class.forName("lu").getField("aF").get(null));
        Object biome = ((Stream<?>)Class.forName("jo").getMethod("b").invoke(biomeLookup)).findFirst().orElseThrow();
        Object source = Class.forName("deh").getConstructor(Class.forName("jm")).newInstance(biome);
        Object generator = Class.forName("dzb").getConstructor(Class.forName("dea"),Class.forName("jm")).newInstance(source,settings);
        Class<?> pos = Class.forName("jd"), state = Class.forName("dtc");
        Method x = pos.getMethod("u"), y = pos.getMethod("v"), z = pos.getMethod("w");
        Method isAir = Class.forName("dtb$a").getMethod("i"), solid = Class.forName("dtb$a").getMethod("d"), fluid = Class.forName("dtb$a").getMethod("u");
        Method emptyFluid = Class.forName("epe").getMethod("c");
        Method stateID = Class.forName("dfy").getMethod("i",state);
        Map<String,Object> blocks = new HashMap<>();
        Object registry = Class.forName("lt").getField("e").get(null);
        for (String name : new String[]{"air","grass_block","water","stone","sand","podzol","short_grass","ice","mycelium","oak_leaves","glass","farmland","magma_block","red_mushroom","brown_mushroom"}) {
            Object block = Class.forName("jz").getMethod("a",Class.forName("akr")).invoke(registry,Class.forName("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name));
            blocks.put(name,Class.forName("dfy").getMethod("o").invoke(block));
        }
        Object lightPosition=pos.getConstructor(int.class,int.class,int.class).newInstance(8,41,8);
        for(int level:new int[]{-1,0,12,13,15})for(String ground:new String[]{"grass_block","podzol","mycelium","stone","oak_leaves","glass","farmland","water","magma_block"})for(String mushroom:new String[]{"brown_mushroom","red_mushroom"}) {
            Object engine=lightEngine(level);
            Object world=Proxy.newProxyInstance(Class.forName("dcz").getClassLoader(),new Class<?>[]{Class.forName("dcz")},(proxy,method,values)->{
                if(method.getName().equals("a_"))return blocks.get(ground);
                if(method.getName().equals("b")&&values.length==2&&values[1] instanceof Integer)return Class.forName("eot").getMethod("a",pos,int.class).invoke(engine,values[0],values[1]);
                throw new UnsupportedOperationException(method.toString());
            });
            boolean survives=(boolean)Class.forName("dtb$a").getMethod("a",Class.forName("dcz"),pos).invoke(blocks.get(mushroom),world,lightPosition);
            int raw=(int)Class.forName("eot").getMethod("a",pos,int.class).invoke(engine,lightPosition,0);
            System.out.printf("mushroom_survival minecraft:%s minecraft:%s %d %d %d%n",mushroom,ground,level,raw,survives?1:0);
        }
        for(int height:new int[]{-64,0,41,319})for(int darkness:new int[]{0,5,15}) {
            Object engine=lightEngine(-1),position=pos.getConstructor(int.class,int.class,int.class).newInstance(8,height,8);
            System.out.printf("feature_brightness %d %d %d%n",height,darkness,Class.forName("eot").getMethod("a",pos,int.class).invoke(engine,position,darkness));
        }
        var entries = new ArrayList<String[]>();
        for (String name : new String[]{"patch_brown_mushroom","patch_red_mushroom"}) entries.add(new String[]{name,"configured"});
        if(args.length>1 && args[1].equals("only")) entries.clear();
        // The second argument lists every runtime-supported top-level vegetation feature with its actual index.
        if (args.length > 0) for (String line : java.nio.file.Files.readAllLines(java.nio.file.Path.of(args[0]))) {
            String[] fields = line.trim().split(" +");
            if (fields.length == 2) entries.add(fields);
        }
        for (String[] entry : entries) {
            String name = "minecraft:"+entry[0];
            boolean primitive = entry[1].equals("configured");
            int index = primitive ? -1 : Integer.parseInt(entry[1]);
            Object feature = primitive ? configured.get(name) : Class.forName("jm").getMethod("a").invoke(Class.forName("eiv").getMethod("b").invoke(placed.get(name)));
            Set<Long> defaultSeeds=Set.of(0L,1L,-1L,123456789L,2L,3L,42L,987654321L);
            List<Long> seeds=new ArrayList<>(List.of(0L,1L,-1L,123456789L,2L,3L,42L,987654321L));
            if(!primitive) {
                int chance=Map.of("brown_mushroom_old_growth",4,"red_mushroom_old_growth",171,"brown_mushroom_swamp",1,"red_mushroom_swamp",64,"brown_mushroom_normal",256,"red_mushroom_normal",512,"brown_mushroom_taiga",4,"red_mushroom_taiga",256).get(entry[0]);
                // Select independent seeds with a successful first rarity draw;
                // each resulting vector still executes the actual placed feature.
                for(long trial=100;seeds.size()<11&&trial<100000;trial++) {
                    Object rng=Class.forName("dzx").getConstructor(Class.forName("ayw")).newInstance(Class.forName("dzz").getConstructor(long.class).newInstance(0L));
                    long decoration=(long)Class.forName("dzx").getMethod("a",long.class,int.class,int.class).invoke(rng,trial,0,0);
                    Class.forName("dzx").getMethod("b",long.class,int.class,int.class).invoke(rng,decoration,index,9);
                    if((float)Class.forName("dzx").getMethod("i").invoke(rng)<1.0f/chance)seeds.add(trial);
                }
                if(seeds.size()!=11)throw new IllegalStateException("failed to find positive rarity seeds");
            }
            for (long seed : seeds) {
                for (int scenario = 0; scenario < (args.length>1 ? 14 : primitive ? 12 : 6); scenario++) {
                    if(!defaultSeeds.contains(seed)&&scenario!=4)continue;
                    int mode = scenario;
                    int[] origin = mode == 1 ? new int[]{-1,41,-16} : mode == 2 ? new int[]{15,41,15}
                        : mode == 8 ? new int[]{8,319,8} : mode == 9 ? new int[]{8,-64,8}
                        : mode == 10 ? new int[]{29999968,41,29999968} : mode == 11 ? new int[]{-29999968,41,-29999968} : new int[]{8,41,8};
                    if (!primitive) { origin[0] = mode % 2 == 0 ? 0 : -272; origin[2] = mode % 2 == 0 ? 0 : 368; }
                    Map<String,Object> changed = new HashMap<>();
                    java.util.function.Function<int[],Object> read = p -> {
                        if (p[1] < -64 || p[1] >= 320) return blocks.get("air");
                        String key = p[0]+" "+p[1]+" "+p[2];
                        if (changed.containsKey(key)) return changed.get(key);
                        if(mode==12 && p[1]==40) return blocks.get(Math.floorMod(p[0],3)==0 ? "water" : "sand");
                        if(mode==13 && p[1]<=40) return blocks.get("ice");
                        int floor = mode == 8 ? 318 : mode == 9 ? -65 : mode == 1 ? 38+Math.floorMod(p[0]+p[2],5) : 40;
                        if (p[1] <= floor) return blocks.get(mode == 3 ? "sand" : mode == 4 ? "podzol" : mode == 5 ? "stone" : "grass_block");
                        if (mode == 6 && p[1] <= 42) return blocks.get("water");
                        if (mode == 7 && p[1] == 42) return blocks.get("stone");
                        return blocks.get("air");
                    };
                    Object lightEngine=lightEngine(-1);
                    Object world = Proxy.newProxyInstance(Class.forName("dds").getClassLoader(),new Class<?>[]{Class.forName("dds")},(proxy,method,values) -> {
                        String m = method.getName();
                        if (m.equals("I_")) return -64;
                        if (m.equals("J_")) return 384;
                        if (m.equals("am")) return 320;
                        if (m.equals("d") && values.length == 1 && values[0] instanceof Integer) return (int)values[0] < -64 || (int)values[0] >= 320;
                        if (m.equals("f_")) return true;
                        if (m.equals("toString")) return "atlas-vegetation-world";
                        if (m.equals("a") && values.length == 3 && Class.forName("dyy$a").isInstance(values[0])) {
                            boolean oceanFloor = values[0] == Class.forName("dyy$a").getField("d").get(null);
                            boolean motion = values[0] == Class.forName("dyy$a").getField("e").get(null);
                            for (int height=319;height>=-64;height--) {
                                Object s=read.apply(new int[]{(int)values[1],height,(int)values[2]});
                                if (oceanFloor ? (boolean)solid.invoke(s) : motion ? (boolean)solid.invoke(s) || !(boolean)emptyFluid.invoke(fluid.invoke(s)) : !(boolean)isAir.invoke(s)) return height+1;
                            }
                            return -64;
                        }
                        if (values != null && values.length > 0 && pos.isInstance(values[0])) {
                            int[] p={(int)x.invoke(values[0]),(int)y.invoke(values[0]),(int)z.invoke(values[0])}; Object s=read.apply(p);
                            if(m.equals("b")&&values.length==2&&values[1] instanceof Integer)return Class.forName("eot").getMethod("a",pos,int.class).invoke(lightEngine,values[0],values[1]);
                            if (m.equals("a_")) return s;
                            if (m.equals("u")) return (boolean)isAir.invoke(s);
                            if (m.equals("b_")) return fluid.invoke(s);
                            if (m.equals("a") && values.length==2 && values[1] instanceof Predicate<?>) return ((Predicate<Object>)values[1]).test(s);
                            if (m.equals("b") && values.length==2 && values[1] instanceof Predicate<?>) return ((Predicate<Object>)values[1]).test(fluid.invoke(s));
                            if (m.equals("a") && values.length==3 && state.isInstance(values[1])) { if (p[1]<-64 || p[1]>=320) return false; changed.put(p[0]+" "+p[1]+" "+p[2],values[1]); return true; }
                            if (method.getReturnType()==void.class) return null;
                        }
                        throw new UnsupportedOperationException(method.toString());
                    });
                    Object random = Class.forName("dzx").getConstructor(Class.forName("ayw")).newInstance(Class.forName("dzz").getConstructor(long.class).newInstance(primitive ? seed : 0));
                    if (!primitive) {
                        long decoration=(long)Class.forName("dzx").getMethod("a",long.class,int.class,int.class).invoke(random,seed,origin[0],origin[2]);
                        Class.forName("dzx").getMethod("b",long.class,int.class,int.class).invoke(random,decoration,index,9);
                    }
                    Stream<?> positions=Stream.of(pos.getConstructor(int.class,int.class,int.class).newInstance(origin[0],primitive ? origin[1] : 0,origin[2]));
                    if (!primitive) {
                        Object context=Class.forName("eiw").getConstructor(Class.forName("dds"),Class.forName("duz"),Optional.class).newInstance(world,generator,Optional.empty());
                        Method positionsMethod=Class.forName("eiy").getMethod("a_",Class.forName("eiw"),Class.forName("ayw"),pos);
                        for(Object modifier:(List<?>)Class.forName("eiv").getMethod("c").invoke(placed.get(name))) {
                            if(modifier.getClass().getName().equals("eii")) continue; // Accepted biome, isolated from biome generation.
                            positions=positions.flatMap(p->{try{return (Stream<?>)positionsMethod.invoke(modifier,context,random,p);}catch(Exception e){throw new RuntimeException(e);}});
                        }
                    }
                    int[] count={0,0};
                    positions.forEach(p->{try{count[0]++; if((boolean)Class.forName("ebq").getMethod("a",Class.forName("dds"),Class.forName("duz"),Class.forName("ayw"),pos).invoke(feature,world,generator,random,p)) count[1]++;}catch(Exception e){throw new RuntimeException(e);}});
                    List<String> output=new ArrayList<>();
                    for(var change:changed.entrySet()) output.add(change.getKey()+" "+stateID.invoke(null,change.getValue()));
                    output.sort(String::compareTo);
                    String hash=HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(String.join("\n",output).getBytes(StandardCharsets.UTF_8)));
                    System.out.printf("vegetation %s %s %d %d %d %d %d %d %d %d %d %s%n",primitive?"configured":"placed",name,seed,scenario,origin[0],origin[1],origin[2],count[0],count[1],changed.size(),Class.forName("dzx").getMethod("g").invoke(random),hash);
                }
            }
        }
    }
    // This is the real official light engine. A missing column is deliberately
    // left uninitialized; explicit test layers are queued through its API.
    private static Object lightEngine(int brightness)throws Exception {
        Object getter=Proxy.newProxyInstance(Class.forName("dcc").getClassLoader(),new Class<?>[]{Class.forName("dcc")},(p,m,v)->{
            if(m.getName().equals("I_"))return -64;
            if(m.getName().equals("J_"))return 384;
            if(m.isDefault())return InvocationHandler.invokeDefault(p,m,v);
            throw new UnsupportedOperationException(m.toString());
        });
        Object provider=Proxy.newProxyInstance(Class.forName("dvl").getClassLoader(),new Class<?>[]{Class.forName("dvl")},(p,m,v)->{
            if(m.getName().equals("q"))return getter;
            if(m.getName().equals("c"))return null;
            if(m.getReturnType()==void.class)return null;
            throw new UnsupportedOperationException(m.toString());
        });
        Class<?> type=Class.forName("eot"),section=Class.forName("kf");
        Object engine=type.getConstructor(Class.forName("dvl"),boolean.class,boolean.class).newInstance(provider,true,true);
        if(brightness>=0) {
            Object sp=section.getMethod("a",int.class,int.class,int.class).invoke(null,0,2,0);
            type.getMethod("a",section,boolean.class).invoke(engine,sp,false);
            Object sky=Class.forName("ddf").getField("a").get(null);
            Object data=Class.forName("dvd").getConstructor(int.class).newInstance(brightness);
            type.getMethod("a",Class.forName("ddf"),section,Class.forName("dvd")).invoke(engine,sky,sp,data);
            type.getMethod("a").invoke(engine);
        }
        return engine;
    }
    private static Map<String,Object> registry(Object lookup,String field) throws Exception {
        Object values=Class.forName("jn$a").getMethod("b",Class.forName("akq")).invoke(lookup,Class.forName("lu").getField(field).get(null));
        Map<String,Object> result=new HashMap<>();
        for(Object holder:((Stream<?>)Class.forName("jo").getMethod("b").invoke(values)).toList()) {
            Object key=((Optional<?>)Class.forName("jm").getMethod("e").invoke(holder)).orElseThrow();
            result.put(Class.forName("akq").getMethod("a").invoke(key).toString(),Class.forName("jm").getMethod("a").invoke(holder));
        }
        return result;
    }
}
