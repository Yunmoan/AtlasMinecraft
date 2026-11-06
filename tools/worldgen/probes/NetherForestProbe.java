package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.*;
import java.util.stream.Stream;

/** Calls the actual 1.21.1 Nether forest vegetation and huge fungi. */
public final class NetherForestProbe {
    // Random tag selection requires resource order, including nested references.
    private static Set<String> orderedTagBlocks(String name) throws Exception {
        String text;
        try(var input=NetherForestProbe.class.getClassLoader().getResourceAsStream("data/minecraft/tags/block/"+name+".json")) {
            if(input==null)throw new IllegalStateException("missing tag "+name);
            text=new String(input.readAllBytes(),StandardCharsets.UTF_8);
        }
        var values=java.util.regex.Pattern.compile("\"(#?minecraft:[^\"]+)\"").matcher(text);
        var result=new LinkedHashSet<String>();
        while(values.find()){String v=values.group(1);if(v.startsWith("#"))result.addAll(orderedTagBlocks(v.substring(11)));else result.add(v);}
        return result;
    }
    public static void main(String[] args) throws Exception {
        Object version = Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a",Class.forName("ae")).invoke(null,version);
        Class.forName("akt").getMethod("a").invoke(null);
        TreeFeatureProbe.bindTreeTags();
        // SnowLayerBlock uses two further tags before checking collision faces.
        Object blockRegistry=Class.forName("lt").getField("e").get(null);
        Map<Object,List<Object>> tags=new HashMap<>();
        for(Object tag:((Stream<?>)Class.forName("jz").getMethod("k").invoke(blockRegistry)).toList()) {
            List<Object> values=new ArrayList<>();
            Object current=Class.forName("jz").getMethod("b",Class.forName("awu")).invoke(blockRegistry,tag);
            if(current instanceof Optional<?> option && option.isPresent()) for(Object holder:(Iterable<?>)option.get()) values.add(holder);
            tags.put(tag,values);
        }
        for(String[] entry:new String[][]{{"snow_layer_cannot_survive_on","cp"},{"snow_layer_can_survive_on","cq"},{"dripstone_replaceable_blocks","bt"},{"base_stone_overworld","bf"},{"moss_replaceable","bv"},{"lush_ground_replaceable","bw"},{"small_dripleaf_placeable","by"},{"azalea_grows_on","ci"},{"azalea_root_replaceable","bx"},{"features_cannot_replace","bP"},{"lava_pool_stone_cannot_replace","bQ"},{"geode_invalid_blocks","bR"},{"coral_blocks","ar"},{"corals","au"},{"wall_corals","as"},{"bamboo_plantable_on","av"},{"mushroom_grow_block","bb"},{"fire","aL"},{"sculk_replaceable","bT"},{"sculk_replaceable_world_gen","bU"},{"nylium","aM"}}) {
            var values=new ArrayList<Object>();
            for(String name:orderedTagBlocks(entry[0])) {
                Object key=Class.forName("akr").getMethod("a",String.class).invoke(null,name);
                Object block=Class.forName("jz").getMethod("a",Class.forName("akr")).invoke(blockRegistry,key);
                values.add(Class.forName("jz").getMethod("e",Object.class).invoke(blockRegistry,block));
            }
            tags.put(Class.forName("awe").getField(entry[1]).get(null),values);
        }
        Class.forName("jz").getMethod("a",Map.class).invoke(blockRegistry,tags);
        Object registries=Class.forName("pd").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        System.setErr(new PrintStream(new FileOutputStream(FileDescriptor.err),true));
        Object lookup=Class.forName("jo$a").getMethod("b").invoke(registries);
        Map<String,Object> features=registry(lookup,"aI"),biomes=registry(lookup,"aF");
        Map<String,Object> placedFeatures=registry(lookup,"aQ");
        Object biomeRegistry=Class.forName("ju").getConstructor(Class.forName("akq"),com.mojang.serialization.Lifecycle.class).newInstance(Class.forName("lu").getField("aF").get(null),com.mojang.serialization.Lifecycle.stable());
        Object biomeLookup=Class.forName("jn$a").getMethod("b",Class.forName("akq")).invoke(lookup,Class.forName("lu").getField("aF").get(null));
        for(Object holder:((Stream<?>)Class.forName("jo").getMethod("b").invoke(biomeLookup)).toList()) Class.forName("jz").getMethod("a",Class.forName("jz"),Class.forName("akq"),Object.class).invoke(null,biomeRegistry,Class.forName("jm$c").getMethod("h").invoke(holder),Class.forName("jm").getMethod("a").invoke(holder));
        Object worldHeight=Class.forName("dcy").getMethod("e",int.class,int.class).invoke(null,-64,384);
        var postConstructor=Class.forName("dvs").getConstructor(Class.forName("dcd"),Class.forName("dvv"),Class.forName("dcy"),Class.forName("jz"),Class.forName("eab"));
        Object settingsLookup=Class.forName("jn$a").getMethod("b",Class.forName("akq")).invoke(lookup,Class.forName("lu").getField("aO").get(null));
        Object settings=Class.forName("jn").getMethod("b",Class.forName("akq")).invoke(settingsLookup,Class.forName("dzd").getField("f").get(null));
        Object biomeHolder=Class.forName("jm").getMethod("a",Object.class).invoke(null,biomes.get("minecraft:nether_wastes"));
        Object biomeSource=Class.forName("deh").getConstructor(Class.forName("jm")).newInstance(biomeHolder);
        Object generator=Class.forName("dzb").getConstructor(Class.forName("dea"),Class.forName("jm")).newInstance(biomeSource,settings);
        Map<String,Object> blocks=new HashMap<>();
        for(String name:new String[]{"air","netherrack","crimson_nylium","warped_nylium","nether_wart_block","warped_wart_block","bedrock","water","lava","crimson_roots","warped_roots","crimson_fungus","warped_fungus","short_grass","dirt","soul_soil","mycelium","stone"}) {
            Object key=Class.forName("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name);
            Object block=Class.forName("jz").getMethod("a",Class.forName("akr")).invoke(blockRegistry,key);
            blocks.put(name,Class.forName("dfy").getMethod("o").invoke(block));
        }
        Class<?> pos=Class.forName("jd"),state=Class.forName("dtc");
        Method x=pos.getMethod("u"),y=pos.getMethod("v"),z=pos.getMethod("w");
        Method isAir=Class.forName("dtb$a").getMethod("i"),fluid=Class.forName("dtb$a").getMethod("u"),stateID=Class.forName("dfy").getMethod("i",state);
        String[] names={"crimson_forest_vegetation","warped_forest_vegetation","nether_sprouts","crimson_forest_vegetation_bonemeal","warped_forest_vegetation_bonemeal","nether_sprouts_bonemeal","crimson_fungus","warped_fungus","crimson_fungus_planted","warped_fungus_planted"};
        for(String name:names)for(long seed:new long[]{0,1,-1,42,123456789,987654321,2,3})for(int scenario=0;scenario<23;scenario++) {
            final int mode=scenario;
            int[] origin=mode==1?new int[]{-1,41,-16}:mode==2?new int[]{15,41,15}:mode==14?new int[]{8,1,8}:mode==15?new int[]{8,127,8}:mode==16?new int[]{8,254,8}:mode==17?new int[]{8,0,8}:new int[]{8,41,8};
            Map<String,Object> changed=new HashMap<>();
            String nylium=name.contains("warped")?"warped_nylium":"crimson_nylium";
            java.util.function.Function<int[],Object> read=p->{
                if(p[1]<0||p[1]>=256)return blocks.get("air");
                String key=p[0]+" "+p[1]+" "+p[2];if(changed.containsKey(key))return changed.get(key);
                int dx=p[0]-origin[0],dy=p[1]-origin[1],dz=p[2]-origin[2];
                if(mode==7)return blocks.get("air");
                int floor=origin[1]-1;
                if(mode==13)floor+=Math.floorMod(dx+dz,3);
                if(p[1]<=floor) {
                    if(p[1]<floor)return blocks.get("netherrack");
                    if(mode==3)return blocks.get("netherrack");
                    if(mode==4)return blocks.get(nylium.equals("crimson_nylium")?"warped_nylium":"crimson_nylium");
                    if(mode==8)return blocks.get("bedrock");
                    if(mode==21&&(dx!=0||dz!=0))return blocks.get(new String[]{nylium,"dirt","soul_soil","mycelium","netherrack"}[Math.floorMod(dx+dz,5)]);
                    return blocks.get(nylium);
                }
                if(mode==5&&dy==0)return blocks.get("water");
                if(mode==6&&dy>=0&&dy<=5)return blocks.get("crimson_roots");
                if(mode==9&&dy>=2&&Math.abs(dx)<=2&&Math.abs(dz)<=2)return blocks.get("netherrack");
                if(mode==10&&dy>=2&&Math.abs(dx)<=2&&Math.abs(dz)<=2)return blocks.get("nether_wart_block");
                if(mode==11&&dy>=0&&dy<=12)return blocks.get("lava");
                if(mode==12&&dy==0)return blocks.get("warped_fungus");
                if(mode==18&&dy>=2&&Math.abs(dx)<=2&&Math.abs(dz)<=2)return blocks.get("stone");
                if(mode==19&&dx==0&&dz==0&&dy==0)return blocks.get("crimson_fungus");
                if(mode==20&&dy==3)return blocks.get("netherrack");
                if(mode==22&&dy>=2&&Math.abs(dx)<=2&&Math.abs(dz)<=2)return blocks.get("short_grass");
                return blocks.get("air");
            };
            Object world=Proxy.newProxyInstance(Class.forName("dds").getClassLoader(),new Class<?>[]{Class.forName("dds")},(proxy,method,values)->{
                String m=method.getName();if(m.equals("I_"))return 0;if(m.equals("J_"))return 256;if(m.equals("am"))return 256;if(m.equals("f_"))return true;
                if(m.equals("toString"))return "atlas-nether-forest-world";
                if(values!=null&&values.length>0&&pos.isInstance(values[0])) {
                    int[] p={(int)x.invoke(values[0]),(int)y.invoke(values[0]),(int)z.invoke(values[0])};Object s=read.apply(p);
                    if(m.equals("a_"))return s;if(m.equals("u"))return (boolean)isAir.invoke(s);if(m.equals("b_"))return fluid.invoke(s);
                    if(m.equals("a")&&values.length==2&&values[1] instanceof java.util.function.Predicate<?>)return ((java.util.function.Predicate<Object>)values[1]).test(s);
                    if((m.equals("a")&&values.length==3&&state.isInstance(values[1]))||(m.equals("b")&&values.length==2&&values[1] instanceof Boolean)) {
                        if(p[1]<0||p[1]>=256)return false;
                        changed.put(p[0]+" "+p[1]+" "+p[2],m.equals("b")?blocks.get("air"):values[1]);return true;
                    }
                    if((m.equals("s")||m.equals("d"))&&values.length==1)return p[1]<0||p[1]>=256;
                }
                if(method.getReturnType()==void.class)return null;
                throw new UnsupportedOperationException(method.toString());
            });
            Object random=Class.forName("dzx").getConstructor(Class.forName("ayw")).newInstance(Class.forName("dzz").getConstructor(long.class).newInstance(seed));
            Object start=pos.getConstructor(int.class,int.class,int.class).newInstance(origin[0],origin[1],origin[2]);
            boolean placed=(boolean)Class.forName("ebq").getMethod("a",Class.forName("dds"),Class.forName("duz"),Class.forName("ayw"),pos).invoke(features.get("minecraft:"+name),world,generator,random,start);
            List<String> output=new ArrayList<>();for(var entry:changed.entrySet())output.add(entry.getKey()+" "+stateID.invoke(null,entry.getValue()));output.sort(String::compareTo);
            String hash=HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(String.join("\n",output).getBytes(StandardCharsets.UTF_8)));
            System.out.printf("nether_forest minecraft:%s %d %d %d %d %d %d %d %d %s%n",name,seed,scenario,origin[0],origin[1],origin[2],placed?1:0,changed.size(),Class.forName("dzx").getMethod("g").invoke(random),hash);
        }
    }
    private static Map<String,Object> registry(Object lookup,String field)throws Exception {
        Object values=Class.forName("jn$a").getMethod("b",Class.forName("akq")).invoke(lookup,Class.forName("lu").getField(field).get(null));
        Map<String,Object> result=new HashMap<>();
        for(Object holder:((Stream<?>)Class.forName("jo").getMethod("b").invoke(values)).toList()){
            Object key=((Optional<?>)Class.forName("jm").getMethod("e").invoke(holder)).orElseThrow();result.put(Class.forName("akq").getMethod("a").invoke(key).toString(),Class.forName("jm").getMethod("a").invoke(holder));
        }return result;
    }
}
