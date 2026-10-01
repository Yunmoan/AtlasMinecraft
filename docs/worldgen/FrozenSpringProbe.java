package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.*;
import java.util.stream.Stream;

/** Calls the actual 1.21.1 ice/snow/spring features against independently defined terrain. */
public final class FrozenSpringProbe {
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
        for(String[] entry:new String[][]{{"snow_layer_cannot_survive_on","cp"},{"snow_layer_can_survive_on","cq"},{"dripstone_replaceable_blocks","bt"},{"base_stone_overworld","bf"},{"moss_replaceable","bv"},{"lush_ground_replaceable","bw"},{"small_dripleaf_placeable","by"},{"azalea_grows_on","ci"},{"azalea_root_replaceable","bx"},{"features_cannot_replace","bP"},{"lava_pool_stone_cannot_replace","bQ"},{"geode_invalid_blocks","bR"}}) {
            var values=new ArrayList<Object>();
            for(String name:TreeFeatureProbe.tagBlocks(entry[0])) {
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
        for(String name:new String[]{"air","stone","grass_block","podzol","sand","water","snow_block","ice","packed_ice","soul_sand","honey_block","oak_leaves","netherrack","dripstone_block","dirt","calcite","lava","deepslate","clay","moss_block","moss_carpet","bedrock","amethyst_block"}) {
            Object key=Class.forName("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name);
            Object block=Class.forName("jz").getMethod("a",Class.forName("akr")).invoke(blockRegistry,key);
            blocks.put(name,Class.forName("dfy").getMethod("o").invoke(block));
        }
        Class<?> pos=Class.forName("jd"),state=Class.forName("dtc");
        Method x=pos.getMethod("u"),y=pos.getMethod("v"),z=pos.getMethod("w");
        Method isAir=Class.forName("dtb$a").getMethod("i"),solid=Class.forName("dtb$a").getMethod("d"),fluid=Class.forName("dtb$a").getMethod("u");
        Method emptyFluid=Class.forName("epe").getMethod("c"),stateID=Class.forName("dfy").getMethod("i",state);
        String[] names={"ice_spike","freeze_top_layer","spring_water","spring_lava_overworld","spring_lava_frozen","spring_nether_open","spring_nether_closed"};
        boolean caves=args.length>0&&args[0].startsWith("caves");
        boolean placedMode=caves&&args[0].equals("caves_placed");
        if(caves) names=new String[]{"pointed_dripstone","underwater_magma","dripstone_cluster","large_dripstone","glow_lichen","cave_vine","cave_vine_in_moss","spore_blossom","moss_patch","moss_patch_ceiling","moss_patch_bonemeal","lush_caves_clay","clay_with_dripleaves","clay_pool_with_dripleaves","moss_vegetation","dripleaf","azalea_tree","rooted_azalea_tree","vines","lake_lava","amethyst_geode"};
        Map<String,int[]> order=new HashMap<>();
        order.put("large_dripstone",new int[]{2,3});order.put("underwater_magma",new int[]{6,26});order.put("dripstone_cluster",new int[]{7,2});order.put("pointed_dripstone",new int[]{7,3});order.put("glow_lichen",new int[]{9,0});order.put("lush_caves_ceiling_vegetation",new int[]{9,22});order.put("cave_vines",new int[]{9,23});order.put("lush_caves_clay",new int[]{9,24});order.put("lush_caves_vegetation",new int[]{9,25});order.put("spore_blossom",new int[]{9,27});
        if(placedMode)names=new String[]{"large_dripstone","underwater_magma","dripstone_cluster","pointed_dripstone","glow_lichen","lush_caves_ceiling_vegetation","cave_vines","lush_caves_clay","lush_caves_vegetation","spore_blossom"};
        for(String name:names) for(long seed:new long[]{0,1,-1,123456789,2,3,42,987654321}) for(int scenario=0;scenario<(caves?64:20);scenario++) {
            if(placedMode && (seed!=0&&seed!=1&&seed!=-1&&seed!=42 || !Set.of(5,20,21,24,30,32,40,47).contains(scenario)))continue;
            int mode=scenario;
            int[] origin=mode==62?new int[]{8,-60,8}:mode>=56?new int[]{8,41,8}:mode==50?new int[]{-1,45,-16}:mode==49||mode==51?new int[]{8,51,8}:mode>=48?new int[]{8,45,8}:mode==1||mode==32 ? new int[]{-1,41,-16}:mode==2||mode==33 ? new int[]{15,41,15}:mode==8||mode==34 ? new int[]{8,319,8}:mode==9||mode==35 ? new int[]{8,-64,8}:mode==10 ? new int[]{29999952,41,29999952}:mode==11 ? new int[]{-29999952,41,-29999952}:new int[]{8,41,8};
            Map<String,Object> changed=new HashMap<>();Set<String> ticks=new TreeSet<>();
            Map<String,Object> postChunks=new HashMap<>();
            java.util.function.Function<int[],Object> read=p->{
                if(p[1]<-64||p[1]>=320)return blocks.get("air");
                String key=p[0]+" "+p[1]+" "+p[2];if(changed.containsKey(key))return changed.get(key);
                if(mode>=56) {
                    if(mode==56&&Arrays.equals(p,origin))return blocks.get("air");
                    return blocks.get(mode==57?"deepslate":mode==58?"water":mode==59?"bedrock":mode==60?"amethyst_block":mode==61?"lava":mode==63&&p[1]==41?"water":"stone");
                }
                if(mode>=48) {
                    if(p[1]<=40)return blocks.get("stone");
                    if(p[1]<46)return blocks.get("air");
                    if(p[1]<50)return blocks.get("stone");
                    if(p[1]==50)return blocks.get(mode==52?"lava":mode==53?"clay":mode==54?"oak_leaves":mode==55?"sand":"grass_block");
                    if(mode==51&&p[1]<=52)return blocks.get("water");
                    return blocks.get("air");
                }
                if(mode>=40) {
                    int floor=mode==47?40+Math.floorMod(p[0]+p[2],3):40;
                    if(p[1]<floor)return blocks.get("stone");
                    if(p[1]==floor)return blocks.get(mode==40||mode==41?"clay":mode==42?"moss_block":mode==44?"dirt":mode==46?"sand":"stone");
                    if(p[1]>=46)return blocks.get("stone");
                    if(mode==41&&p[1]<=43)return blocks.get("water");
                    if(mode==45&&p[1]==42)return blocks.get("stone");
                    return blocks.get("air");
                }
                if(mode>=20) {
                    int floor=mode==20?40:mode==22?20:mode==23?10:mode==34?318:mode==35?-65:mode==36?37:mode==37?36:mode==38?35:36;
                    int ceiling=mode==20?44:mode==22?64:mode==23?100:mode==28?42:mode==34?321:mode==35?-58:46;
                    if(mode==30){floor=34+Math.floorMod(p[0]+p[2],5);ceiling=45+Math.floorMod(p[0]-p[2],7);}
                    if(mode==31)return blocks.get("air");
                    if(mode==39&&Arrays.equals(p,origin))return blocks.get("stone");
                    if(p[1]<=floor&&mode!=28)return blocks.get(mode==25&&p[1]==floor?"lava":mode==26&&p[1]==floor?"dripstone_block":mode==27?"dirt":mode==29?"deepslate":"stone");
                    if(p[1]>=ceiling)return blocks.get(mode==27?"calcite":"stone");
                    return blocks.get(mode==24||mode>=36&&mode<=38?"water":"air");
                }
                if(mode>=16) {if(p[0]==origin[0]&&p[1]==origin[1]&&p[2]==origin[2])return blocks.get("air");if(p[0]==origin[0]+1&&p[1]==origin[1]&&p[2]==origin[2]&&mode!=17)return blocks.get("air");return blocks.get(mode==18 ? "packed_ice":mode==19 ? "netherrack":"stone");}
                int floor=mode==8 ? 318:mode==9 ? -65:mode==1 ? 38+Math.floorMod(p[0]+p[2],5):40;
                if(p[1]<floor)return blocks.get("stone");
                if(p[1]==floor)return blocks.get(mode==3 ? "sand":mode==4 ? "podzol":mode==5 ? "stone":mode==12 ? "soul_sand":mode==13 ? "honey_block":mode==14 ? "ice":mode==15 ? "oak_leaves":mode==7 ? "snow_block":"grass_block");
                if(mode==6&&p[1]<=44)return blocks.get("water");
                return blocks.get("air");
            };
            Object biome=biomes.get(mode==0 ? "minecraft:plains":"minecraft:snowy_plains");
            Object world=Proxy.newProxyInstance(Class.forName("dds").getClassLoader(),new Class<?>[]{Class.forName("dds")},(proxy,method,values)->{
                String m=method.getName();if(m.equals("I_"))return -64;if(m.equals("J_"))return 384;if(m.equals("am"))return 320;if(m.equals("f_"))return true;
                if(m.equals("C"))return seed;
                if(m.equals("toString"))return "atlas-frozen-spring-world";
                if(m.equals("y")) {
                    int cx=(int)x.invoke(values[0])>>4,cz=(int)z.invoke(values[0])>>4;String key=cx+" "+cz;
                    if(!postChunks.containsKey(key))postChunks.put(key,postConstructor.newInstance(Class.forName("dcd").getConstructor(int.class,int.class).newInstance(cx,cz),Class.forName("dvv").getField("a").get(null),worldHeight,biomeRegistry,null));
                    return postChunks.get(key);
                }
                if(m.equals("d")&&values.length==1&&values[0] instanceof Integer)return (int)values[0]<-64||(int)values[0]>=320;
                if(m.equals("a")&&values.length==2&&Class.forName("ddf").isInstance(values[0]))return 0;
                if(m.equals("a")&&values.length==3&&Class.forName("dyy$a").isInstance(values[0])) {
                    boolean ocean=values[0]==Class.forName("dyy$a").getField("d").get(null);
                    boolean motion=values[0]==Class.forName("dyy$a").getField("e").get(null);
                    for(int height=319;height>=-64;height--){Object s=read.apply(new int[]{(int)values[1],height,(int)values[2]});if(ocean ? (boolean)solid.invoke(s):motion ? (boolean)solid.invoke(s)||!(boolean)emptyFluid.invoke(fluid.invoke(s)):!(boolean)isAir.invoke(s))return height+1;}return -64;
                }
                if(values!=null&&values.length>0&&pos.isInstance(values[0])) {
                    int[] p={(int)x.invoke(values[0]),(int)y.invoke(values[0]),(int)z.invoke(values[0])};Object s=read.apply(p);
                    if(m.equals("a_"))return s;if(m.equals("u"))return (boolean)isAir.invoke(s);if(m.equals("b_"))return fluid.invoke(s);
                    if(m.equals("a")&&values.length==2&&values[1] instanceof java.util.function.Predicate<?>)return ((java.util.function.Predicate<Object>)values[1]).test(s);
                    if(m.equals("b")&&values.length==2&&values[1] instanceof java.util.function.Predicate<?>)return ((java.util.function.Predicate<Object>)values[1]).test(fluid.invoke(s));
                    if(m.equals("z"))return !(boolean)emptyFluid.invoke(fluid.invoke(s));
                    if(m.equals("d")&&values.length==1)return p[1]<-64||p[1]>=320;
                    if(m.equals("t"))return Class.forName("jm").getMethod("a",Object.class).invoke(null,biome);
                    if(m.equals("a")&&values.length==3&&state.isInstance(values[1])){if(p[1]<-64||p[1]>=320)return false;changed.put(p[0]+" "+p[1]+" "+p[2],values[1]);return true;}
                    // Tree shape notifications schedule gameplay leaf/fluid ticks;
                    // the tree oracle checks generation edits, as TreeFeatureProbe does.
                    if(method.getReturnType()==void.class){if(!name.equals("azalea_tree")&&!name.equals("rooted_azalea_tree")&&!(values.length>1&&Class.forName("dfy").isInstance(values[1])))ticks.add(p[0]+" "+p[1]+" "+p[2]);return null;}
                }
                throw new UnsupportedOperationException(method.toString());
            });
            Object random=Class.forName("dzx").getConstructor(Class.forName("ayw")).newInstance(Class.forName("dzz").getConstructor(long.class).newInstance(seed));
            Object start=pos.getConstructor(int.class,int.class,int.class).newInstance(origin[0],origin[1],origin[2]);
            boolean placed;
            if(placedMode) {
                int cx=origin[0]>>4,cz=origin[2]>>4;int[] index=order.get(name);
                long decoration=(long)Class.forName("dzx").getMethod("a",long.class,int.class,int.class).invoke(random,seed,cx*16,cz*16);
                Class.forName("dzx").getMethod("b",long.class,int.class,int.class).invoke(random,decoration,index[1],index[0]);
                Object top=placedFeatures.get("minecraft:"+name),configured=Class.forName("jm").getMethod("a").invoke(Class.forName("eiv").getMethod("b").invoke(top));
                Object context=Class.forName("eiw").getConstructor(Class.forName("dds"),Class.forName("duz"),Optional.class).newInstance(world,generator,Optional.empty());
                Stream<?> positions=Stream.of(pos.getConstructor(int.class,int.class,int.class).newInstance(cx*16,0,cz*16));
                Method positionsMethod=Class.forName("eiy").getMethod("a_",Class.forName("eiw"),Class.forName("ayw"),pos);
                for(Object modifier:(List<?>)Class.forName("eiv").getMethod("c").invoke(top)) {
                    if(modifier.getClass().getName().equals("eii"))continue;
                    positions=positions.flatMap(p->{try{return (Stream<?>)positionsMethod.invoke(modifier,context,random,p);}catch(Exception ex){throw new RuntimeException(ex);}});
                }
                boolean[] result={false};
                positions.forEach(p->{try{result[0]|=(boolean)Class.forName("ebq").getMethod("a",Class.forName("dds"),Class.forName("duz"),Class.forName("ayw"),pos).invoke(configured,world,generator,random,p);}catch(Exception ex){throw new RuntimeException(ex);}});
                placed=result[0];
            } else placed=(boolean)Class.forName("ebq").getMethod("a",Class.forName("dds"),Class.forName("duz"),Class.forName("ayw"),pos).invoke(features.get("minecraft:"+name),world,generator,random,start);
            for(var entry:postChunks.entrySet()) {
                String[] coords=entry.getKey().split(" ");int cx=Integer.parseInt(coords[0]),cz=Integer.parseInt(coords[1]);
                Object[] sections=(Object[])Class.forName("duy").getMethod("n").invoke(entry.getValue());
                for(int section=0;section<sections.length;section++)if(sections[section]!=null)for(Object packed:(Iterable<?>)sections[section]) {int v=((Short)packed)&65535;ticks.add((cx*16+(v&15))+" "+(-64+section*16+((v>>4)&15))+" "+(cz*16+((v>>8)&15)));}
            }
            List<String> output=new ArrayList<>();for(var entry:changed.entrySet())output.add(entry.getKey()+" "+stateID.invoke(null,entry.getValue()));output.sort(String::compareTo);
            String hash=HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(String.join("\n",output).getBytes(StandardCharsets.UTF_8)));
            System.out.printf("%s minecraft:%s %d %d %d %d %d %d %d %d %s %d%n",placedMode?"cave_placement":caves?"cave_feature":"frozen_spring",name,seed,scenario,origin[0],origin[1],origin[2],placed?1:0,changed.size(),Class.forName("dzx").getMethod("g").invoke(random),hash,ticks.size());
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
