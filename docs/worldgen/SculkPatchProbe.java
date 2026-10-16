package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.*;
import java.util.stream.Stream;

/** Calls the actual 1.21.1 sculk patches against independently defined terrain. */
public final class SculkPatchProbe {
    // Random tag selection requires resource order, including nested references.
    private static Set<String> orderedTagBlocks(String name) throws Exception {
        String text;
        try(var input=SculkPatchProbe.class.getClassLoader().getResourceAsStream("data/minecraft/tags/block/"+name+".json")) {
            if(input==null)throw new IllegalStateException("missing tag "+name);
            text=new String(input.readAllBytes(),StandardCharsets.UTF_8);
        }
        var values=java.util.regex.Pattern.compile("\"(#?minecraft:[^\"]+)\"").matcher(text);
        var result=new LinkedHashSet<String>();
        while(values.find()){String v=values.group(1);if(v.startsWith("#"))result.addAll(orderedTagBlocks(v.substring(11)));else result.add(v);}
        return result;
    }
    public static void main(String[] args) throws Exception {
        boolean featurePool=args.length>0&&args[0].equals("feature_pool");
        int poolRotation=featurePool&&args.length>1?Integer.parseInt(args[1]):0;
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
        for(String[] entry:new String[][]{{"snow_layer_cannot_survive_on","cp"},{"snow_layer_can_survive_on","cq"},{"dripstone_replaceable_blocks","bt"},{"base_stone_overworld","bf"},{"moss_replaceable","bv"},{"lush_ground_replaceable","bw"},{"small_dripleaf_placeable","by"},{"azalea_grows_on","ci"},{"azalea_root_replaceable","bx"},{"features_cannot_replace","bP"},{"lava_pool_stone_cannot_replace","bQ"},{"geode_invalid_blocks","bR"},{"coral_blocks","ar"},{"corals","au"},{"wall_corals","as"},{"bamboo_plantable_on","av"},{"mushroom_grow_block","bb"},{"fire","aL"},{"sculk_replaceable","bT"},{"sculk_replaceable_world_gen","bU"}}) {
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
        Object settings=Class.forName("jn").getMethod("b",Class.forName("akq")).invoke(settingsLookup,Class.forName("dzd").getField("c").get(null));
        Object biomeHolder=Class.forName("jm").getMethod("a",Object.class).invoke(null,biomes.get("minecraft:plains"));
        Object biomeSource=Class.forName("deh").getConstructor(Class.forName("jm")).newInstance(biomeHolder);
        Object generator=Class.forName("dzb").getConstructor(Class.forName("dea"),Class.forName("jm")).newInstance(biomeSource,settings);
        Map<String,Object> blocks=new HashMap<>();
        for(String name:new String[]{"air","stone","grass_block","podzol","sand","water","snow_block","ice","packed_ice","soul_sand","honey_block","oak_leaves","netherrack","dripstone_block","dirt","calcite","lava","deepslate","clay","moss_block","moss_carpet","bedrock","amethyst_block","magma_block","kelp","kelp_plant","seagrass","tall_seagrass","sea_pickle","tube_coral","tube_coral_block","oak_slab","bamboo","bamboo_sapling","dirt_path","mycelium","farmland","gravel","chest","spawner","sculk","sculk_catalyst","moving_piston","sculk_vein","fire","short_grass","bubble_column","sculk_sensor","sculk_shrieker","deepslate_bricks","glass"}) {
            Object key=Class.forName("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name);
            Object block=Class.forName("jz").getMethod("a",Class.forName("akr")).invoke(blockRegistry,key);
            blocks.put(name,Class.forName("dfy").getMethod("o").invoke(block));
        }
        blocks.put("flowing_water",Class.forName("dte").getMethod("a",Class.forName("duf"),Comparable.class).invoke(blocks.get("water"),Class.forName("dts").getField("aP").get(null),1));
        Class<?> pos=Class.forName("jd"),state=Class.forName("dtc");
        Method x=Class.forName("kh").getMethod("u"),y=Class.forName("kh").getMethod("v"),z=Class.forName("kh").getMethod("w");
        Method isAir=Class.forName("dtb$a").getMethod("i"),fluid=Class.forName("dtb$a").getMethod("u"),stateID=Class.forName("dfy").getMethod("i",state);
        long[] seeds={0,1,-1,42,123456789,987654321};
        for(String name:new String[]{"sculk_patch_deep_dark","sculk_patch_ancient_city","sculk_patch_growth_probe","sculk_patch_dense_probe"})for(long seed:seeds)for(int mode=0;mode<20;mode++) {
            if(featurePool&&!name.equals("sculk_patch_ancient_city"))continue;
            boolean custom=name.endsWith("_probe");
            if(custom&&(seed!=0&&seed!=1&&seed!=42||!Set.of(6,7,9,11).contains(mode)))continue;
            final int scenario=mode;
            int[] origin=mode==12?new int[]{-1,41,-16}:mode==13?new int[]{15,41,15}:mode==14?new int[]{8,319,8}:mode==15?new int[]{8,-64,8}:new int[]{8,41,8};
            Map<String,Object> changed=new HashMap<>(),postChunks=new HashMap<>();
            Set<String> marked=new HashSet<>();
            java.util.function.Function<int[],Object> read=p->{
                if(p[1]<-64||p[1]>=320)return blocks.get("air");
                String key=p[0]+" "+p[1]+" "+p[2];if(changed.containsKey(key))return changed.get(key);
                int dy=p[1]-origin[1],dx=p[0]-origin[0],dz=p[2]-origin[2];
                if(scenario==0)return blocks.get("air");
                if(scenario==1)return blocks.get("stone");
                if(scenario==2)return blocks.get("water");
                if(scenario==3&&dy==0)return blocks.get("flowing_water");
                if(scenario==4&&dy==0)return blocks.get("lava");
                if(scenario==5&&dy==-1)return blocks.get("sculk");
                if(scenario==6&&dy<=-1)return blocks.get("deepslate_bricks");
                if(scenario==7&&dy==0&&dx==0&&dz==0)return blocks.get("sculk");
                if(scenario==8&&dy<=-1)return blocks.get("oak_slab");
                if(scenario==9&&dy<=-1)return blocks.get("bedrock");
                if(scenario==10&&dy<=-1)return blocks.get("sand");
                if(scenario==11&&dy>=0&&dy<=2)return blocks.get("water");
                if(scenario>=16&&scenario<=19&&dy<=-1)return blocks.get(scenario==16?"oak_leaves":scenario==17?"glass":scenario==18?"farmland":"magma_block");
                if(dy<=-1||dy>=4)return blocks.get("stone");
                return blocks.get("air");
            };
            Object world=Proxy.newProxyInstance(Class.forName("dds").getClassLoader(),new Class<?>[]{Class.forName("dds")},(proxy,method,values)->{
                String m=method.getName();if(m.equals("I_"))return -64;if(m.equals("J_"))return 384;if(m.equals("am"))return 320;if(m.equals("f_"))return true;
                if(m.equals("toString"))return "atlas-sculk-patch-world";
                if(m.equals("y")) {
                    int cx=(int)x.invoke(values[0])>>4,cz=(int)z.invoke(values[0])>>4;String key=cx+" "+cz;
                    if(!postChunks.containsKey(key))postChunks.put(key,postConstructor.newInstance(Class.forName("dcd").getConstructor(int.class,int.class).newInstance(cx,cz),Class.forName("dvv").getField("a").get(null),worldHeight,biomeRegistry,null));
                    return postChunks.get(key);
                }
                if(values!=null&&values.length>0&&pos.isInstance(values[0])) {
                    int[] p={(int)x.invoke(values[0]),(int)y.invoke(values[0]),(int)z.invoke(values[0])};
                    Object s=read.apply(p);
                    if(m.equals("a_"))return s;if(m.equals("b_"))return fluid.invoke(s);
                    if(m.equals("a")&&values.length==3&&state.isInstance(values[1])){
                        if(p[1]<-64||p[1]>=320)return false;
                        changed.put(p[0]+" "+p[1]+" "+p[2],values[1]);return true;
                    }
                }
                if(method.getReturnType()==void.class)return null;
                throw new UnsupportedOperationException(method.toString());
            });
            Object random=Class.forName("dzx").getConstructor(Class.forName("ayw")).newInstance(Class.forName("dzz").getConstructor(long.class).newInstance(seed));
            Object start=pos.getConstructor(int.class,int.class,int.class).newInstance(origin[0],origin[1],origin[2]);
            Object configured=features.get("minecraft:"+name);
            if(custom) {
                boolean dense=name.contains("dense");
                Object rare=Class.forName("bpt").getMethod("a",int.class).invoke(null,dense?0:2);
                Object config=Class.forName("eez").getConstructor(int.class,int.class,int.class,int.class,int.class,Class.forName("bpw"),float.class).newInstance(dense?32:2,dense?500:120,64,dense?1:2,2,rare,dense?1.0f:0.0f);
                Object feature=Class.forName("ebq").getMethod("b").invoke(features.get("minecraft:sculk_patch_deep_dark"));
                configured=Class.forName("ebq").getConstructor(Class.forName("ece"),Class.forName("eei")).newInstance(feature,config);
            }
            boolean placed;
            if(featurePool) {
                Object holder=Class.forName("jm").getMethod("a",Object.class).invoke(null,placedFeatures.get("minecraft:"+name));
                var constructor=Class.forName("ekt").getDeclaredConstructor(Class.forName("jm"),Class.forName("elb$a"));constructor.setAccessible(true);
                Object element=constructor.newInstance(holder,Class.forName("elb$a").getEnumConstants()[0]);
                Object farClip=Class.forName("ejj").getConstructor(int.class,int.class,int.class,int.class,int.class,int.class).newInstance(256,-64,256,271,319,271);
                Object pivot=pos.getConstructor(int.class,int.class,int.class).newInstance(999,999,999);
                placed=(boolean)Class.forName("ekt").getMethod("a",Class.forName("enu"),Class.forName("dds"),Class.forName("ddq"),Class.forName("duz"),pos,pos,Class.forName("dmm"),Class.forName("ejj"),Class.forName("ayw"),Class.forName("end"),boolean.class).invoke(element,null,world,null,generator,start,pivot,Class.forName("dmm").getEnumConstants()[poolRotation],farClip,random,Class.forName("end").getEnumConstants()[0],false);
            } else {
                placed=(boolean)Class.forName("ebq").getMethod("a",Class.forName("dds"),Class.forName("duz"),Class.forName("ayw"),pos).invoke(configured,world,generator,random,start);
            }
            for(var entry:postChunks.entrySet()) {
                String[] coords=entry.getKey().split(" ");int cx=Integer.parseInt(coords[0]),cz=Integer.parseInt(coords[1]);
                Object[] sections=(Object[])Class.forName("duy").getMethod("n").invoke(entry.getValue());
                for(int section=0;section<sections.length;section++)if(sections[section]!=null)for(Object packed:(Iterable<?>)sections[section]) {int v=((Short)packed)&65535;marked.add((cx*16+(v&15))+" "+(-64+section*16+((v>>4)&15))+" "+(cz*16+((v>>8)&15)));}
            }
            List<String> output=new ArrayList<>();for(var entry:changed.entrySet())output.add(entry.getKey()+" "+stateID.invoke(null,entry.getValue()));output.sort(String::compareTo);
            String hash=HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(String.join("\n",output).getBytes(StandardCharsets.UTF_8)));
            System.out.printf("%s minecraft:%s %d %d %d %d %d %d %d %d %s %d%s%n",featurePool?"jigsaw_feature":"sculk_patch",name,seed,scenario,origin[0],origin[1],origin[2],placed?1:0,changed.size(),Class.forName("dzx").getMethod("g").invoke(random),hash,marked.size(),featurePool?" "+poolRotation:"");
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
