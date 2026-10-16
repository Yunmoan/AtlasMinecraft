package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.*;
import java.util.stream.Stream;

/** Calls the actual 1.21.1 Nether vines, glowstone and geological decorations. */
public final class EndFeatureProbe {
    // Random tag selection requires resource order, including nested references.
    private static Set<String> orderedTagBlocks(String name) throws Exception {
        String text;
        try(var input=EndFeatureProbe.class.getClassLoader().getResourceAsStream("data/minecraft/tags/block/"+name+".json")) {
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
        Object settings=Class.forName("jn").getMethod("b",Class.forName("akq")).invoke(settingsLookup,Class.forName("dzd").getField("g").get(null));
        Object biomeHolder=Class.forName("jm").getMethod("a",Object.class).invoke(null,biomes.get("minecraft:the_end"));
        Object biomeSource=Class.forName("deh").getConstructor(Class.forName("jm")).newInstance(biomeHolder);
        Object generator=Class.forName("dzb").getConstructor(Class.forName("dea"),Class.forName("jm")).newInstance(biomeSource,settings);
        Map<String,Object> blocks=new HashMap<>();
        for(String name:new String[]{"air","end_stone","stone","water","lava","obsidian","bedrock","iron_bars","end_gateway","chorus_plant","chorus_flower"}) {
            Object key=Class.forName("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name);
            Object block=Class.forName("jz").getMethod("a",Class.forName("akr")).invoke(blockRegistry,key);
            blocks.put(name,Class.forName("dfy").getMethod("o").invoke(block));
        }
        Class<?> pos=Class.forName("jd"),state=Class.forName("dtc");
        Method x=pos.getMethod("u"),y=pos.getMethod("v"),z=pos.getMethod("w");
        Method isAir=Class.forName("dtb$a").getMethod("i"),fluid=Class.forName("dtb$a").getMethod("u"),stateID=Class.forName("dfy").getMethod("i",state);
        Object gatewayNone=Class.forName("eeh").getMethod("a").invoke(null);
        Object gatewayExit=pos.getConstructor(int.class,int.class,int.class).newInstance(-100,50,123);
        Object gatewayInexact=Class.forName("eeh").getMethod("a",pos,boolean.class).invoke(null,gatewayExit,false);
        var configured=Class.forName("ebq").getConstructor(Class.forName("ece"),Class.forName("eei"));
        features.put("minecraft:end_gateway_none",configured.newInstance(Class.forName("ece").getField("M").get(null),gatewayNone));
        features.put("minecraft:end_gateway_inexact",configured.newInstance(Class.forName("ece").getField("M").get(null),gatewayInexact));
        Object guarded=Class.forName("edo$a").getConstructor(int.class,int.class,int.class,int.class,boolean.class).newInstance(8,8,2,76,true);
        Object small=Class.forName("edo$a").getConstructor(int.class,int.class,int.class,int.class,boolean.class).newInstance(10,10,3,80,false);
        Object beam=pos.getConstructor(int.class,int.class,int.class).newInstance(0,128,0);
        Object guardConfig=Class.forName("efc").getConstructor(boolean.class,List.class,pos).newInstance(false,List.of(guarded),null);
        Object beamConfig=Class.forName("efc").getConstructor(boolean.class,List.class,pos).newInstance(true,List.of(guarded,small),beam);
        features.put("minecraft:end_spike_guarded",configured.newInstance(Class.forName("ece").getField("K").get(null),guardConfig));
        features.put("minecraft:end_spike_beam",configured.newInstance(Class.forName("ece").getField("K").get(null),beamConfig));
        Field unsafeField=sun.misc.Unsafe.class.getDeclaredField("theUnsafe");unsafeField.setAccessible(true);
        var unsafe=(sun.misc.Unsafe)unsafeField.get(null);
        Object server=unsafe.allocateInstance(Class.forName("apn")),level=unsafe.allocateInstance(Class.forName("aqu"));
        Object worldData=Proxy.newProxyInstance(Class.forName("erl").getClassLoader(),new Class<?>[]{Class.forName("erl")},(proxy,method,values)->{if(method.getName().equals("K"))return Class.forName("cpn").getField("g").get(null);throw new UnsupportedOperationException(method.toString());});
        Field serverData=Class.forName("net.minecraft.server.MinecraftServer").getDeclaredField("j");serverData.setAccessible(true);serverData.set(server,worldData);
        Field serverField=Class.forName("aqu").getDeclaredField("J");serverField.setAccessible(true);serverField.set(level,server);
        long[] seeds={0,1,-1,42,123456789,987654321,2,3};
        Map<Long,List<?>> layouts=new HashMap<>();
        for(long seed:seeds){Object seedWorld=Proxy.newProxyInstance(Class.forName("dds").getClassLoader(),new Class<?>[]{Class.forName("dds")},(proxy,method,values)->{if(method.getName().equals("C"))return seed;throw new UnsupportedOperationException(method.toString());});List<?> layout=(List<?>)Class.forName("edo").getMethod("a",Class.forName("dds")).invoke(null,seedWorld);layouts.put(seed,layout);for(int i=0;i<layout.size();i++){Object s=layout.get(i);System.out.printf("end_spike_layout %d %d %s %s %s %s %s%n",seed,i,Class.forName("edo$a").getMethod("a").invoke(s),Class.forName("edo$a").getMethod("b").invoke(s),Class.forName("edo$a").getMethod("c").invoke(s),Class.forName("edo$a").getMethod("d").invoke(s),(boolean)Class.forName("edo$a").getMethod("e").invoke(s)?1:0);}}
        String[] names={"end_island","chorus_plant","end_gateway_return","end_platform","end_spike","end_gateway_none","end_gateway_inexact","end_spike_guarded","end_spike_beam"};
        for(String name:names)for(long seed:seeds)for(int scenario=0;scenario<18;scenario++) {
            final int mode=scenario;
            int[] origin=mode==1?new int[]{-1,70,-16}:mode==2?new int[]{15,70,15}:mode==12?new int[]{8,1,8}:mode==13?new int[]{8,0,8}:mode==14?new int[]{8,255,8}:mode==15?new int[]{8,254,8}:new int[]{8,70,8};
            if(name.equals("end_spike")&&mode<10){Object s=layouts.get(seed).get(mode);origin[0]=(int)Class.forName("edo$a").getMethod("a").invoke(s);origin[2]=(int)Class.forName("edo$a").getMethod("b").invoke(s);}
            Map<String,Object> changed=new HashMap<>(),entities=new HashMap<>();
            List<Object> crystals=new ArrayList<>();
            java.util.function.Function<int[],Object> read=p->{
                if(p[1]<0||p[1]>=256)return blocks.get("air");
                String key=p[0]+" "+p[1]+" "+p[2];if(changed.containsKey(key))return changed.get(key);
                int dx=p[0]-origin[0],dy=p[1]-origin[1],dz=p[2]-origin[2];
                if(mode==3)return blocks.get("air");
                if(mode==4)return blocks.get("stone");
                int floor=origin[1]-1;if(name.startsWith("end_spike"))floor=66;
                if(mode==10)floor+=Math.floorMod(dx+dz,5)-2;
                if(p[1]<=floor)return blocks.get(mode==5?"stone":mode==6?"water":mode==7?"obsidian":"end_stone");
                if(mode==8&&dy==0)return blocks.get("water");
                if(mode==9&&dy==0)return blocks.get("lava");
                if(mode==11&&dy>=2&&dy<=5&&Math.abs(dx)<=2&&Math.abs(dz)<=2)return blocks.get("stone");
                if(mode==16&&dy>=0&&dy<=6&&dx==2)return blocks.get("chorus_plant");
                if(mode==17&&dy>=0&&dy<=6&&Math.floorMod(dx+dz,3)==0)return blocks.get("end_stone");
                return blocks.get("air");
            };
            Object world=Proxy.newProxyInstance(Class.forName("dds").getClassLoader(),new Class<?>[]{Class.forName("dds")},(proxy,method,values)->{
                String m=method.getName();if(m.equals("I_"))return 0;if(m.equals("J_"))return 256;if(m.equals("am"))return 256;if(m.equals("f_"))return true;
                if(m.equals("E"))return level;if(m.equals("C"))return seed;if(m.equals("b")&&values.length==1&&Class.forName("bsr").isInstance(values[0])){crystals.add(values[0]);return true;}
                if(m.equals("toString"))return "atlas-nether-decoration-world";
                if(values!=null&&values.length>0&&pos.isInstance(values[0])) {
                    int[] p={(int)x.invoke(values[0]),(int)y.invoke(values[0]),(int)z.invoke(values[0])};Object s=read.apply(p);
                    if(m.equals("c_"))return entities.get(p[0]+" "+p[1]+" "+p[2]);
                    if(m.equals("a_"))return s;if(m.equals("u"))return (boolean)isAir.invoke(s);if(m.equals("b_"))return fluid.invoke(s);
                    if(m.equals("a")&&values.length==2&&values[1] instanceof java.util.function.Predicate<?>)return ((java.util.function.Predicate<Object>)values[1]).test(s);
                    if((m.equals("a")&&values.length==3&&state.isInstance(values[1]))||(m.equals("b")&&values.length==2&&values[1] instanceof Boolean)) {
                        if(p[1]<0||p[1]>=256)return false;
                        changed.put(p[0]+" "+p[1]+" "+p[2],m.equals("b")?blocks.get("air"):values[1]);
                        if(m.equals("a")&&values[1]==blocks.get("end_gateway"))entities.put(p[0]+" "+p[1]+" "+p[2],Class.forName("dry").getConstructor(pos,state).newInstance(values[0],values[1]));
                        return true;
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
            System.out.printf("end_feature minecraft:%s %d %d %d %d %d %d %d %d %s%n",name,seed,scenario,origin[0],origin[1],origin[2],placed?1:0,changed.size(),Class.forName("dzx").getMethod("g").invoke(random),hash);
            for(var entry:entities.entrySet()) {
                Object saved=Class.forName("dqh").getMethod("b",Class.forName("jo$a")).invoke(entry.getValue(),registries);
                var bytes=io.netty.buffer.Unpooled.buffer();Object buffer=Class.forName("vw").getConstructor(io.netty.buffer.ByteBuf.class).newInstance(bytes);Class.forName("vw").getMethod("a",Class.forName("uy")).invoke(buffer,saved);byte[] data=new byte[bytes.readableBytes()];bytes.readBytes(data);bytes.release();
                System.out.printf("end_gateway_nbt minecraft:%s %d %d %s %s%n",name,seed,scenario,entry.getKey(),HexFormat.of().formatHex(data));
            }
            for(Object crystal:crystals) {
                Object target=Class.forName("chz").getMethod("p").invoke(crystal);String beamText=target==null?"none":x.invoke(target)+","+y.invoke(target)+","+z.invoke(target);
                System.out.printf("end_crystal minecraft:%s %d %d %s %s %s %s %s %s%n",name,seed,scenario,Class.forName("bsr").getMethod("dt").invoke(crystal),Class.forName("bsr").getMethod("dv").invoke(crystal),Class.forName("bsr").getMethod("dz").invoke(crystal),Float.floatToRawIntBits((float)Class.forName("bsr").getMethod("dE").invoke(crystal)),(boolean)Class.forName("bsr").getMethod("cv").invoke(crystal)?1:0,beamText);
            }

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
