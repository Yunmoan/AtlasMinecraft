package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.*;
import java.util.stream.Stream;

/** Calls the actual 1.21.1 Nether vines, glowstone and geological decorations. */
public final class ScatteredOreProbe {
    // Random tag selection requires resource order, including nested references.
    private static Set<String> orderedTagBlocks(String name) throws Exception {
        String text;
        try(var input=ScatteredOreProbe.class.getClassLoader().getResourceAsStream("data/minecraft/tags/block/"+name+".json")) {
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
        for(String[] entry:new String[][]{{"snow_layer_cannot_survive_on","cp"},{"snow_layer_can_survive_on","cq"},{"dripstone_replaceable_blocks","bt"},{"base_stone_overworld","bf"},{"moss_replaceable","bv"},{"lush_ground_replaceable","bw"},{"small_dripleaf_placeable","by"},{"azalea_grows_on","ci"},{"azalea_root_replaceable","bx"},{"features_cannot_replace","bP"},{"lava_pool_stone_cannot_replace","bQ"},{"geode_invalid_blocks","bR"},{"coral_blocks","ar"},{"corals","au"},{"wall_corals","as"},{"bamboo_plantable_on","av"},{"mushroom_grow_block","bb"},{"fire","aL"},{"sculk_replaceable","bT"},{"sculk_replaceable_world_gen","bU"},{"nylium","aM"},{"base_stone_nether","bi"}}) {
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
        for(String name:new String[]{"air","netherrack","warped_nylium","nether_wart_block","warped_wart_block","bedrock","water","lava","stone","basalt","blackstone","magma_block","soul_sand","nether_bricks","nether_brick_fence","nether_brick_stairs","nether_wart","chest","spawner","glowstone"}) {
            Object key=Class.forName("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name);
            Object block=Class.forName("jz").getMethod("a",Class.forName("akr")).invoke(blockRegistry,key);
            blocks.put(name,Class.forName("dfy").getMethod("o").invoke(block));
        }
        Class<?> pos=Class.forName("jd"),state=Class.forName("dtc");
        Method x=pos.getMethod("u"),y=pos.getMethod("v"),z=pos.getMethod("w");
        Method isAir=Class.forName("dtb$a").getMethod("i"),fluid=Class.forName("dtb$a").getMethod("u"),stateID=Class.forName("dfy").getMethod("i",state);
        for(int variant=0;variant<8;variant++) {
            int size=new int[]{0,1,8,16,64,64,64,8}[variant];
            float chance=new float[]{0,0,.25f,.5f,1,.5f,.5f,0}[variant];
            List<Object> targets=new ArrayList<>();
            Object base=Class.forName("dtb$a").getMethod("b").invoke(blocks.get(variant==6?"stone":"netherrack"));
            Object predicate=Class.forName("emv").getConstructor(Class.forName("dfy")).newInstance(base);
            targets.add(Class.forName("eeq").getMethod("a",Class.forName("enn"),state).invoke(null,predicate,blocks.get(variant==7?"netherrack":"basalt")));
            if(variant==5||variant==6) {
                base=Class.forName("dtb$a").getMethod("b").invoke(blocks.get("netherrack"));
                predicate=Class.forName("emv").getConstructor(Class.forName("dfy")).newInstance(base);
                targets.add(Class.forName("eeq").getMethod("a",Class.forName("enn"),state).invoke(null,predicate,blocks.get("blackstone")));
            }
            Object config=Class.forName("eeq").getConstructor(List.class,int.class,float.class).newInstance(targets,size,chance);
            features.put("minecraft:scatter_custom_"+variant,Class.forName("ebq").getConstructor(Class.forName("ece"),Class.forName("eei")).newInstance(Class.forName("ece").getField("af").get(null),config));
        }
        String[] names={"ore_ancient_debris_large","ore_ancient_debris_small","scatter_custom_0","scatter_custom_1","scatter_custom_2","scatter_custom_3","scatter_custom_4","scatter_custom_5","scatter_custom_6","scatter_custom_7"};
        for(String name:names)for(long seed:new long[]{0,1,-1,42,123456789,987654321,2,3})for(int scenario=0;scenario<28;scenario++) {
            final int mode=scenario;
            int[] origin=mode==1?new int[]{-1,41,-16}:mode==2?new int[]{15,41,15}:mode==18?new int[]{8,1,8}:mode==19?new int[]{8,0,8}:mode==20?new int[]{8,255,8}:mode==21?new int[]{8,128,8}:mode==22?new int[]{8,32,8}:mode==23?new int[]{8,33,8}:mode==24?new int[]{8,30,8}:new int[]{8,41,8};
            Map<String,Object> changed=new HashMap<>();
            boolean ceiling=name.equals("weeping_vines")||name.equals("glowstone_extra")||name.equals("basalt_pillar");
            java.util.function.Function<int[],Object> read=p->{
                if(p[1]<0||p[1]>=256)return blocks.get("air");
                String key=p[0]+" "+p[1]+" "+p[2];if(changed.containsKey(key))return changed.get(key);
                int dx=p[0]-origin[0],dy=p[1]-origin[1],dz=p[2]-origin[2];
                if(mode==3)return blocks.get("air");
                if(mode==4)return blocks.get("stone");
                int floor=origin[1]-(ceiling?12:1),roof=origin[1]+1;
                if(mode==15)floor+=Math.floorMod(dx+dz,5)-2;
                if(mode==25)floor+=Math.floorMod(dx*dx+dz*dz,7)-3;
                String support=switch(mode){case 5->"stone";case 6->"basalt";case 7->"blackstone";case 8->"bedrock";case 9->"nether_wart_block";case 10->"warped_wart_block";case 11->"warped_nylium";case 12->"lava";default->"netherrack";};
                if(mode==17)support=new String[]{"bedrock","magma_block","soul_sand","nether_bricks","nether_brick_fence","nether_brick_stairs","nether_wart","chest","spawner","netherrack"}[Math.floorMod(dx+dz,10)];
                if(p[1]<=floor)return blocks.get(p[1]==floor?support:"netherrack");
                if(ceiling&&p[1]==roof&&!(mode==16&&(dx!=0||dz!=0)))return blocks.get(support);
                if(mode==13&&dy==0)return blocks.get("lava");
                if(mode==14&&dy==0)return blocks.get("water");
                if(mode==26&&dy>=-6&&dy<=4&&Math.floorMod(dx*3+dz,5)==0)return blocks.get("netherrack");
                if(mode==27&&dy==0&&dx==0&&dz==0)return blocks.get("glowstone");
                return blocks.get("air");
            };
            Object world=Proxy.newProxyInstance(Class.forName("dds").getClassLoader(),new Class<?>[]{Class.forName("dds")},(proxy,method,values)->{
                String m=method.getName();if(m.equals("I_"))return 0;if(m.equals("J_"))return 256;if(m.equals("am"))return 256;if(m.equals("f_"))return true;
                if(m.equals("toString"))return "atlas-nether-decoration-world";
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
            System.out.printf("scattered_ore minecraft:%s %d %d %d %d %d %d %d %d %s%n",name,seed,scenario,origin[0],origin[1],origin[2],placed?1:0,changed.size(),Class.forName("dzx").getMethod("g").invoke(random),hash);
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
