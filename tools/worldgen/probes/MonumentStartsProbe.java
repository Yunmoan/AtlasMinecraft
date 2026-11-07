package atlasprobe;

import com.mojang.datafixers.util.Pair;
import java.io.*;
import java.util.*;
import java.util.function.Consumer;
import java.util.function.Predicate;
import java.util.stream.Stream;

/** Full original monument start, with fixed biomes and real Overworld climate. */
public final class MonumentStartsProbe {
    static Class<?> cls(String name)throws Exception{return Class.forName(name);}
    public static void main(String[] args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);
        cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);
        Object registries=IglooProbeSupport.registryAccess();RuinedPortalStartsProbe.bindBiomeTags(registries);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Object lookup=cls("jo$a").getMethod("b").invoke(registries);
        var registry=cls("jn$a").getMethod("b",cls("akq"));
        Object structures=registry.invoke(lookup,cls("lu").getField("aR").get(null));
        Object biomes=registry.invoke(lookup,cls("lu").getField("aF").get(null));
        Object settingsLookup=registry.invoke(lookup,cls("lu").getField("aO").get(null));
        Map<String,Object> holders=new HashMap<>();Object structure=null;
        for(Object holder:((Stream<?>)cls("jo").getMethod("b").invoke(biomes)).toList()) holders.put(cls("akq").getMethod("a").invoke(cls("jm$c").getMethod("h").invoke(holder)).toString(),holder);
        for(Object holder:((Stream<?>)cls("jo").getMethod("b").invoke(structures)).toList()) if(cls("akq").getMethod("a").invoke(cls("jm$c").getMethod("h").invoke(holder)).toString().equals("minecraft:monument")) structure=cls("jm").getMethod("a").invoke(holder);
        Object allowed=cls("ejr").getMethod("a").invoke(structure);
        Predicate<Object> predicate=holder->{try{return (Boolean)cls("jq").getMethod("a",cls("jm")).invoke(allowed,holder);}catch(Exception e){throw new IllegalStateException(e);}};
        List<Pair<Object,Object>> points=new ArrayList<>();
        var addPoints=cls("dem").getDeclaredMethod("a",Consumer.class);addPoints.setAccessible(true);
        addPoints.invoke(cls("dem").getConstructor().newInstance(),(Consumer<Pair<Object,Object>>)pair->{try { points.add(Pair.of(pair.getFirst(),holders.get(cls("akq").getMethod("a").invoke(pair.getSecond()).toString()))); } catch(Exception e){throw new IllegalStateException(e);}});
        Object natural=cls("dej").getMethod("a",cls("def$c")).invoke(null,cls("def$c").getConstructor(List.class).newInstance(points));
        Object key=cls("dzd").getField("c").get(null), settings=cls("jn").getMethod("b",cls("akq")).invoke(settingsLookup,key);
        Object heights=cls("dcy").getMethod("e",int.class,int.class).invoke(null,-64,384);
        Object oceanFloor=cls("dyy$a").getField("c").get(null);
        Object serialization=cls("ekh").getConstructor(cls("aue"),cls("ka"),cls("enu")).newInstance(null,registries,null);
        int[][] positions={{0,0},{1,-1},{-1,1},{17,-9},{-41,23},{0,43},{-24,61},{145,-62}};
        for(long seed:new long[]{0,1,-1,1234,Long.MIN_VALUE,Long.MAX_VALUE,4511451270791457643L}) {
            Object noise=cls("dzm").getMethod("a",cls("jn$a"),cls("akq"),long.class).invoke(null,lookup,key,seed);
            for(String biome:new String[]{"deep_ocean","deep_cold_ocean","deep_frozen_ocean","deep_lukewarm_ocean","ocean","river","frozen_river","plains","natural"}) {
                Object source=biome.equals("natural")?natural:cls("deh").getConstructor(cls("jm")).newInstance(holders.get("minecraft:"+biome));
                Object generator=cls("dzb").getConstructor(cls("dea"),cls("jm")).newInstance(source,settings);
                for(int[] pos:positions) {
                    Object chunk=cls("dcd").getConstructor(int.class,int.class).newInstance(pos[0],pos[1]);
                    int height=(Integer)cls("duz").getMethod("c",int.class,int.class,cls("dyy$a"),cls("dcy"),cls("dzm")).invoke(generator,pos[0]*16+8,pos[1]*16+8,oceanFloor,heights,noise);
                    Object context=cls("ejr$a").getConstructor(cls("ka"),cls("duz"),cls("dea"),cls("dzm"),cls("enu"),long.class,cls("dcd"),cls("dcy"),Predicate.class).newInstance(registries,generator,source,noise,null,seed,chunk,heights,predicate);
                    Optional<?> stub=(Optional<?>)cls("ejr").getMethod("b",cls("ejr$a")).invoke(structure,context);
                    int count=0;String saved="-";
                    if(stub.isPresent()) {
                        Object builder=cls("ejr$b").getMethod("a").invoke(stub.get());
                        Object container=cls("ekj").getMethod("a").invoke(builder);
                        count=((List<?>)cls("ekg").getMethod("c").invoke(container)).size();
                        Object start=cls("ejz").getConstructor(cls("ejr"),cls("dcd"),int.class,cls("ekg")).newInstance(structure,chunk,0,container);
                        saved=MonumentLayoutProbe.hash(IglooProbeSupport.canonicalNBT(cls("ejz").getMethod("a",cls("ekh"),cls("dcd")).invoke(start,serialization,chunk)));
                    }
                    long next=(Long)cls("ayw").getMethod("g").invoke(cls("ejr$a").getMethod("f").invoke(context));
                    System.out.printf("monument_start %d %d %d %s %d %d %s %d%n",seed,pos[0],pos[1],biome,height,count,saved,next);
                }
            }
        }
    }
}
