package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.HexFormat;
import java.util.List;
import java.util.function.Predicate;
import java.util.stream.Stream;

/** Original fossil height sampling, base-column descent, biome gate and piece. */
public final class NetherFossilStartsProbe {
    static Class<?> cls(String n)throws Exception{return Class.forName(n);}
    public static void main(String[] args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);
        Object registries=IglooProbeSupport.registryAccess(),lookup=cls("jo$a").getMethod("b").invoke(registries);
        String[] names=new String[14];for(int i=0;i<14;i++)names[i]="nether_fossils/fossil_"+(i+1);Object manager=IglooProbeSupport.manager(names);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        var registry=cls("jn$a").getMethod("b",cls("akq"));Object structure=null,biome=null;
        for(Object h:((Stream<?>)cls("jo").getMethod("b").invoke(registry.invoke(lookup,cls("lu").getField("aR").get(null)))).toList())if(cls("akq").getMethod("a").invoke(cls("jm$c").getMethod("h").invoke(h)).toString().equals("minecraft:nether_fossil"))structure=cls("jm").getMethod("a").invoke(h);
        for(Object h:((Stream<?>)cls("jo").getMethod("b").invoke(registry.invoke(lookup,cls("lu").getField("aF").get(null)))).toList())if(cls("akq").getMethod("a").invoke(cls("jm$c").getMethod("h").invoke(h)).toString().equals("minecraft:soul_sand_valley"))biome=h;
        Object settingsKey=cls("dzd").getField("f").get(null),settingsLookup=registry.invoke(lookup,cls("lu").getField("aO").get(null));
        Object settings=cls("jn").getMethod("b",cls("akq")).invoke(settingsLookup,settingsKey),heights=cls("dcy").getMethod("e",int.class,int.class).invoke(null,0,256);
        Object source=cls("deh").getConstructor(cls("jm")).newInstance(biome),generator=cls("dzb").getConstructor(cls("dea"),cls("jm")).newInstance(source,settings);
        var generate=cls("ejr").getMethod("a",cls("ka"),cls("duz"),cls("dea"),cls("dzm"),cls("enu"),long.class,cls("dcd"),int.class,cls("dcy"),Predicate.class);
        for(long seed:new long[]{0,1,-1,1234,Long.MAX_VALUE,9212937586943976949L}) {
            Object noise=cls("dzm").getMethod("a",cls("jn$a"),cls("akq"),long.class).invoke(null,lookup,settingsKey,seed);
            for(int[] p:new int[][]{{0,0},{1,-1},{-1,1},{17,-9},{-41,23},{31,31},{-33,32},{145,-62},{-9,-15},{19,5},{-63,-27},{72,13},{5,6},{-8,-5},{101,-31},{-121,23}})for(boolean valid:new boolean[]{true,false}) {
                Object start=generate.invoke(structure,registries,generator,source,noise,manager,seed,cls("dcd").getConstructor(int.class,int.class).newInstance(p[0],p[1]),0,heights,(Predicate<Object>)h->valid);
                boolean accepted=(Boolean)cls("ejz").getMethod("b").invoke(start);String digest="-";
                if(accepted) {Object piece=((List<?>)cls("ejz").getMethod("i").invoke(start)).get(0),saved=cls("ejv").getMethod("a",cls("ekh")).invoke(piece,(Object)null);digest=HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(IglooProbeSupport.canonicalNBT(saved).getBytes(StandardCharsets.UTF_8)));}
                System.out.printf("nether_fossil_start %d %d %d %s %d %s%n",seed,p[0],p[1],valid?"minecraft:soul_sand_valley":"minecraft:nether_wastes",accepted?1:0,digest);
            }
        }
    }
}
