package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.HexFormat;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.function.Predicate;
import java.util.jar.JarFile;
import java.util.stream.Stream;

/** Executes the original deferred Jigsaw assembly, then serializes each full
 * original StructureStart and its post-expansion WorldgenRandom continuation. */
public final class JigsawStartsProbe {
    static Class<?> cls(String name)throws Exception{return Class.forName(name);}
    static String hash(Object nbt)throws Exception{return HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(IglooProbeSupport.canonicalNBT(nbt).getBytes(StandardCharsets.UTF_8)));}
    public static void main(String[] args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);
        Object registries=IglooProbeSupport.registryAccess();List<String> names=new ArrayList<>();
        try(JarFile jar=new JarFile(new java.io.File(cls("ent").getProtectionDomain().getCodeSource().getLocation().toURI()))) {
            var entries=jar.entries();while(entries.hasMoreElements()){String name=entries.nextElement().getName();if(name.startsWith("data/minecraft/structure/") && name.endsWith(".nbt"))names.add(name.substring("data/minecraft/structure/".length(),name.length()-4));}
        }
        Object manager=IglooProbeSupport.manager(names.toArray(String[]::new));
        var repository=cls("enu").getDeclaredField("f");repository.setAccessible(true);
        ((Map<Object,Object>)repository.get(manager)).put(cls("akr").getMethod("a",String.class).invoke(null,"minecraft:ancient_city/walls/intact_horizontal_wall_stairs_5"),Optional.empty());
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Object lookup=cls("jo$a").getMethod("b").invoke(registries);var registry=cls("jn$a").getMethod("b",cls("akq"));
        Object structureLookup=registry.invoke(lookup,cls("lu").getField("aR").get(null)),biomeLookup=registry.invoke(lookup,cls("lu").getField("aF").get(null));
        Map<String,Object> structures=new HashMap<>(),biomes=new HashMap<>();
        for(Object h:((Stream<?>)cls("jo").getMethod("b").invoke(structureLookup)).toList())structures.put(cls("akq").getMethod("a").invoke(cls("jm$c").getMethod("h").invoke(h)).toString(),cls("jm").getMethod("a").invoke(h));
        for(Object h:((Stream<?>)cls("jo").getMethod("b").invoke(biomeLookup)).toList())biomes.put(cls("akq").getMethod("a").invoke(cls("jm$c").getMethod("h").invoke(h)).toString(),h);
        Object serialization=cls("ekh").getConstructor(cls("aue"),cls("ka"),cls("enu")).newInstance(null,registries,manager);
        String[][] cases={{"bastion_remnant","nether_wastes"},{"ancient_city","deep_dark"},{"pillager_outpost","plains"},{"trail_ruins","taiga"},{"trial_chambers","plains"},{"village_desert","desert"},{"village_plains","plains"},{"village_savanna","savanna"},{"village_snowy","snowy_plains"},{"village_taiga","taiga"}};
        for(String[] test:cases) {
        if(args.length>0 && !test[0].equals(args[0]))continue;
        boolean nether=test[0].equals("bastion_remnant");Object structure=structures.get("minecraft:"+test[0]);
        Object key=cls("dzd").getField(nether?"f":"c").get(null),settings=cls("jn").getMethod("b",cls("akq")).invoke(registry.invoke(lookup,cls("lu").getField("aO").get(null)),key),heights=cls("dcy").getMethod("e",int.class,int.class).invoke(null,nether?0:-64,nether?256:384);
        for(long seed:new long[]{0,1,-1,1234,Long.MAX_VALUE,9212937586943976949L}) {
            Object noise=cls("dzm").getMethod("a",cls("jn$a"),cls("akq"),long.class).invoke(null,lookup,key,seed);
            for(String name:new String[]{test[1],"end_highlands"}) {
                Object source=cls("deh").getConstructor(cls("jm")).newInstance(biomes.get("minecraft:"+name)),generator=cls("dzb").getConstructor(cls("dea"),cls("jm")).newInstance(source,settings);
                for(int[] p:new int[][]{{0,0},{1,-1},{-1,1},{17,-9},{-41,23},{31,31},{-33,32},{145,-62},{-9,-15},{19,5},{-63,-27},{72,13},{5,6},{-8,-5},{101,-31},{-121,23}}) {
                    Object chunkPos=cls("dcd").getConstructor(int.class,int.class).newInstance(p[0],p[1]);
                    Object context=cls("ejr$a").getConstructor(cls("ka"),cls("duz"),cls("dea"),cls("dzm"),cls("enu"),long.class,cls("dcd"),cls("dcy"),Predicate.class).newInstance(registries,generator,source,noise,manager,seed,chunkPos,heights,(Predicate<Object>)h->!name.equals("end_highlands"));
                    Optional<?> stub=(Optional<?>)cls("ejr").getMethod("b",cls("ejr$a")).invoke(structure,context);
                    int size=0;String digest="-";
                    if(stub.isPresent()) {
                        Object builder=cls("ejr$b").getMethod("a").invoke(stub.get()),container=cls("ekj").getMethod("a").invoke(builder);
                        size=((List<?>)cls("ekg").getMethod("c").invoke(container)).size();
                        Object start=cls("ejz").getConstructor(cls("ejr"),cls("dcd"),int.class,cls("ekg")).newInstance(structure,chunkPos,0,container);
                        Object saved=cls("ejz").getMethod("a",cls("ekh"),cls("dcd")).invoke(start,serialization,chunkPos);digest=hash(saved);
                        if(seed==0 && p[0]==0 && p[1]==0)new PrintStream(new FileOutputStream(FileDescriptor.err),true).println("jigsaw_canonical "+test[0]+" "+IglooProbeSupport.canonicalNBT(saved));
                    }
                    Object random=cls("ejr$a").getMethod("f").invoke(context);
                    System.out.printf("jigsaw_start minecraft:%s %d %d %d minecraft:%s %d %s %s%n",test[0],seed,p[0],p[1],name,size,digest,cls("ayw").getMethod("g").invoke(random));
                }
            }
        }
        }
    }
}
