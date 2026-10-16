package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.util.*;

/** Runs the actual seven SaplingBlock types and MangrovePropaguleBlock. */
public final class SaplingSurvivalProbe {
    public static void main(String[] args)throws Exception {
        Object version=Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a",Class.forName("ae")).invoke(null,version);
        Class.forName("akt").getMethod("a").invoke(null);
        TreeFeatureProbe.bindTreeTags();
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Class<?> pos=Class.forName("jd");
        Object registry=Class.forName("lt").getField("e").get(null);
        String[] grounds={"dirt","grass_block","podzol","coarse_dirt","mycelium","rooted_dirt","moss_block","mud","muddy_mangrove_roots","farmland","clay","stone","sand","water","air","oak_leaves","mangrove_leaves","dirt_path","snow_block","crimson_nylium","warped_nylium"};
        String[] names={"oak_sapling","spruce_sapling","birch_sapling","jungle_sapling","acacia_sapling","dark_oak_sapling","cherry_sapling","mangrove_propagule"};
        Map<String,Object> blocks=new HashMap<>();
        List<String> all=new ArrayList<>(List.of(grounds));all.addAll(List.of(names));
        for(String name:all) {
            Object block=Class.forName("jz").getMethod("a",Class.forName("akr")).invoke(registry,Class.forName("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name));
            blocks.put(name,Class.forName("dfy").getMethod("o").invoke(block));
        }
        Object origin=pos.getConstructor(int.class,int.class,int.class).newInstance(-1,41,15);
        for(String name:names)for(String ground:grounds)for(String ceiling:new String[]{"air","oak_leaves","mangrove_leaves"})for(int variant=0;variant<(name.equals("mangrove_propagule")?3:1);variant++) {
            Object state=blocks.get(name);
            if(variant>0)state=Class.forName("dte").getMethod("a",Class.forName("duf"),Comparable.class).invoke(state,Class.forName("dts").getField(variant==1?"C":"j").get(null),true);
            Object world=Proxy.newProxyInstance(Class.forName("dcz").getClassLoader(),new Class<?>[]{Class.forName("dcz")},(proxy,method,values)->{
                if(method.getName().equals("a_")) {
                    int height=(int)pos.getMethod("v").invoke(values[0]);
                    return blocks.get(height==40?ground:height==42?ceiling:"air");
                }
                throw new UnsupportedOperationException(method.toString());
            });
            boolean survives=(boolean)Class.forName("dtb$a").getMethod("a",Class.forName("dcz"),pos).invoke(state,world,origin);
            System.out.printf("sapling_survival minecraft:%s minecraft:%s minecraft:%s %d %d%n",name,ground,ceiling,variant,survives?1:0);
        }
    }
}
