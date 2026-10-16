package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.*;
import java.util.stream.Stream;

/** Calls the actual 1.21.1 Nether forest vegetation and huge fungi. */
public final class NetherSurvivalProbe {
    // Random tag selection requires resource order, including nested references.
    private static Set<String> orderedTagBlocks(String name) throws Exception {
        String text;
        try(var input=NetherSurvivalProbe.class.getClassLoader().getResourceAsStream("data/minecraft/tags/block/"+name+".json")) {
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
        for(String[] entry:new String[][]{{"snow_layer_cannot_survive_on","cp"},{"snow_layer_can_survive_on","cq"},{"dripstone_replaceable_blocks","bt"},{"base_stone_overworld","bf"},{"moss_replaceable","bv"},{"lush_ground_replaceable","bw"},{"small_dripleaf_placeable","by"},{"azalea_grows_on","ci"},{"azalea_root_replaceable","bx"},{"features_cannot_replace","bP"},{"lava_pool_stone_cannot_replace","bQ"},{"geode_invalid_blocks","bR"},{"coral_blocks","ar"},{"corals","au"},{"wall_corals","as"},{"bamboo_plantable_on","av"},{"mushroom_grow_block","bb"},{"fire","aL"},{"sculk_replaceable","bT"},{"sculk_replaceable_world_gen","bU"},{"nylium","aM"},{"soul_fire_base_blocks","aU"}}) {
            var values=new ArrayList<Object>();
            for(String name:orderedTagBlocks(entry[0])) {
                Object key=Class.forName("akr").getMethod("a",String.class).invoke(null,name);
                Object block=Class.forName("jz").getMethod("a",Class.forName("akr")).invoke(blockRegistry,key);
                values.add(Class.forName("jz").getMethod("e",Object.class).invoke(blockRegistry,block));
            }
            tags.put(Class.forName("awe").getField(entry[1]).get(null),values);
        }
        Class.forName("jz").getMethod("a",Map.class).invoke(blockRegistry,tags);
        Class<?> pos=Class.forName("jd"),state=Class.forName("dtc");
        Object origin=pos.getConstructor(int.class,int.class,int.class).newInstance(-1,41,15);
        Method stateID=Class.forName("dfy").getMethod("i",state);
        Object fire=Class.forName("dga").getField("cr").get(null);
        Method canBurn=Class.forName("diy").getDeclaredMethod("f",state);canBurn.setAccessible(true);
        byte[] flags=new byte[(26684+7)/8];
        for(int id=0;id<26684;id++)if((boolean)canBurn.invoke(fire,Class.forName("dfy").getMethod("a",int.class).invoke(null,id)))flags[id/8]|=1<<(id%8);
        System.out.printf("fire_burnable_states 26684 %s%n",Base64.getEncoder().encodeToString(flags));
        String[] grounds={"dirt","grass_block","podzol","coarse_dirt","mycelium","rooted_dirt","moss_block","mud","muddy_mangrove_roots","farmland","clay","stone","sand","water","air","oak_leaves","mangrove_leaves","dirt_path","snow_block","crimson_nylium","warped_nylium","soul_sand","soul_soil","netherrack","oak_planks","crimson_planks","glass","oak_slab","oak_stairs","magma_block","bedrock","nether_wart_block","warped_wart_block","red_carpet","white_wool","short_grass","warped_roots","snow","lava"};
        String[] sides={"air","oak_planks","oak_leaves","crimson_planks","warped_wart_block","white_wool","red_carpet","oak_slab","oak_stairs","short_grass","warped_roots","stone","water"};
        String[] names={"fire","soul_fire","crimson_roots","warped_roots","crimson_fungus","warped_fungus","nether_sprouts"};
        Map<String,Object> blocks=new HashMap<>();List<String> all=new ArrayList<>(List.of(grounds));all.addAll(List.of(sides));all.addAll(List.of(names));
        for(String name:all){Object block=Class.forName("jz").getMethod("a",Class.forName("akr")).invoke(blockRegistry,Class.forName("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name));blocks.put(name,Class.forName("dfy").getMethod("o").invoke(block));}
        for(String name:names)for(String ground:grounds)for(String side:sides)for(int variant=0;variant<2;variant++) {
            Object neighbor=blocks.get(side),below=blocks.get(ground);
            if(variant==1) {
                Object waterlogged=Class.forName("dts").getField("C").get(null);
                if((boolean)Class.forName("dte").getMethod("b",Class.forName("duf")).invoke(neighbor,waterlogged))neighbor=Class.forName("dte").getMethod("a",Class.forName("duf"),Comparable.class).invoke(neighbor,waterlogged,true);
                if((boolean)Class.forName("dte").getMethod("b",Class.forName("duf")).invoke(below,waterlogged))below=Class.forName("dte").getMethod("a",Class.forName("duf"),Comparable.class).invoke(below,waterlogged,true);
            }
            final Object neighborState=neighbor,groundState=below;
            Object world=Proxy.newProxyInstance(Class.forName("dcz").getClassLoader(),new Class<?>[]{Class.forName("dcz")},(proxy,method,values)->{
                if(method.getName().equals("a_")){int xx=(int)pos.getMethod("u").invoke(values[0]),yy=(int)pos.getMethod("v").invoke(values[0]),zz=(int)pos.getMethod("w").invoke(values[0]);if(xx==-1&&yy==40&&zz==15)return groundState;if(xx==0&&yy==41&&zz==15)return neighborState;return blocks.get("air");}
                throw new UnsupportedOperationException(method.toString());
            });
            boolean survives=(boolean)Class.forName("dtb$a").getMethod("a",Class.forName("dcz"),pos).invoke(blocks.get(name),world,origin);
            System.out.printf("nether_survival minecraft:%s %d %d %d%n",name,stateID.invoke(null,below),stateID.invoke(null,neighbor),survives?1:0);
        }
    }
}
