package atlasprobe;

import java.io.*;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.*;
import java.util.function.Predicate;
import java.util.stream.Stream;

/** Original StrongholdPieces assembly and actual StrongholdStructure starts. */
public final class StrongholdStartsProbe {
    static Class<?> cls(String name)throws Exception{return Class.forName(name);}
    static String hash(String text)throws Exception{return HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(text.getBytes(StandardCharsets.UTF_8)));}
    static String savedPieces(List<?> pieces)throws Exception {
        StringBuilder result=new StringBuilder();
        for(Object piece:pieces)result.append(IglooProbeSupport.canonicalNBT(cls("ejv").getMethod("a",cls("ekh")).invoke(piece,(Object)null))).append('\n');
        return result.toString();
    }
    public static void main(String[]args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);
        Object registries=IglooProbeSupport.registryAccess();RuinedPortalStartsProbe.bindBiomeTags(registries);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        System.setErr(new PrintStream(new FileOutputStream(FileDescriptor.err),true));
        long[] seeds={0,1,-1,1234,Long.MIN_VALUE,Long.MAX_VALUE,9212937586943976949L};
        int[][] positions={{0,0},{1,-1},{-1,1},{17,-9},{-41,23},{31,31},{-33,32},{145,-62},{-100,-125}};
        for(long seed:seeds)for(int[]chunk:positions)for(int[]heights:new int[][]{{-64,63},{0,63}}) {
            Object builder=cls("ekj").getConstructor().newInstance(),start=null,random=null;
            int attempts=0;
            do {
                cls("ekj").getMethod("b").invoke(builder);
                random=cls("dyz").getConstructor(long.class).newInstance(seed+(long)attempts++);
                long mx=(Long)cls("ayw").getMethod("g").invoke(random),mz=(Long)cls("ayw").getMethod("g").invoke(random);
                cls("ayw").getMethod("b",long.class).invoke(random,(long)chunk[0]*mx^(long)chunk[1]*mz^(seed+(long)attempts-1));
                cls("emj").getMethod("a").invoke(null);
                start=cls("emj$m").getConstructor(cls("ayw"),int.class,int.class).newInstance(random,chunk[0]*16+2,chunk[1]*16+2);
                cls("ejw").getMethod("a",cls("ejv")).invoke(builder,start);
                var children=cls("ejv").getMethod("a",cls("ejv"),cls("ejw"),cls("ayw"));children.invoke(start,start,builder,random);
                List<?> pending=(List<?>)cls("emj$m").getField("c").get(start);
                while(!pending.isEmpty()) {
                    int index=(Integer)cls("ayw").getMethod("a",int.class).invoke(random,pending.size());
                    children.invoke(pending.remove(index),start,builder,random);
                }
                cls("ekj").getMethod("a",int.class,int.class,cls("ayw"),int.class).invoke(builder,heights[1],heights[0],random,10);
            }while((Boolean)cls("ekj").getMethod("c").invoke(builder)||cls("emj$m").getField("b").get(start)==null);
            List<?> pieces=(List<?>)cls("ekg").getMethod("c").invoke(cls("ekj").getMethod("a").invoke(builder));
            String saved=savedPieces(pieces);
            System.out.printf("stronghold_layout %d %d %d %d %d %d %d %s %d%n",seed,chunk[0],chunk[1],heights[0],heights[1],attempts,pieces.size(),hash(saved),(Long)cls("ayw").getMethod("g").invoke(random));
            if(seed==0&&chunk[0]==0&&heights[0]==-64)System.err.println("stronghold_pieces\n"+saved);
        }
        Object lookup=cls("jo$a").getMethod("b").invoke(registries);var registry=cls("jn$a").getMethod("b",cls("akq"));
        Object structures=registry.invoke(lookup,cls("lu").getField("aR").get(null)),biomes=registry.invoke(lookup,cls("lu").getField("aF").get(null)),settingsLookup=registry.invoke(lookup,cls("lu").getField("aO").get(null));
        Object structure=null;Map<String,Object> biomeHolders=new HashMap<>();
        for(Object holder:((Stream<?>)cls("jo").getMethod("b").invoke(structures)).toList())if(cls("akq").getMethod("a").invoke(cls("jm$c").getMethod("h").invoke(holder)).toString().equals("minecraft:stronghold"))structure=cls("jm").getMethod("a").invoke(holder);
        for(Object holder:((Stream<?>)cls("jo").getMethod("b").invoke(biomes)).toList())biomeHolders.put(cls("akq").getMethod("a").invoke(cls("jm$c").getMethod("h").invoke(holder)).toString(),holder);
        Object key=cls("dzd").getField("c").get(null),settings=cls("jn").getMethod("b",cls("akq")).invoke(settingsLookup,key),heights=cls("dcy").getMethod("e",int.class,int.class).invoke(null,-64,384);
        Object serialization=cls("ekh").getConstructor(cls("aue"),cls("ka"),cls("enu")).newInstance(null,registries,null);
        for(long seed:seeds) {
            Object noise=cls("dzm").getMethod("a",cls("jn$a"),cls("akq"),long.class).invoke(null,lookup,key,seed);
            for(String name:new String[]{"minecraft:plains","minecraft:the_void"}) {
                Object source=cls("deh").getConstructor(cls("jm")).newInstance(biomeHolders.get(name)),generator=cls("dzb").getConstructor(cls("dea"),cls("jm")).newInstance(source,settings);
                for(int[]chunk:positions) {
                    Object chunkPos=cls("dcd").getConstructor(int.class,int.class).newInstance(chunk[0],chunk[1]),allowed=cls("ejr").getMethod("a").invoke(structure);
                    Predicate<Object> predicate=holder->{try{return(Boolean)cls("jq").getMethod("a",cls("jm")).invoke(allowed,holder);}catch(Exception e){throw new IllegalStateException(e);}};
                    Object context=cls("ejr$a").getConstructor(cls("ka"),cls("duz"),cls("dea"),cls("dzm"),cls("enu"),long.class,cls("dcd"),cls("dcy"),Predicate.class).newInstance(registries,generator,source,noise,null,seed,chunkPos,heights,predicate);
                    Optional<?> stub=(Optional<?>)cls("ejr").getMethod("b",cls("ejr$a")).invoke(structure,context);int count=0;String digest="-";
                    if(stub.isPresent()) {
                        Object builder=cls("ejr$b").getMethod("a").invoke(stub.get()),container=cls("ekj").getMethod("a").invoke(builder);count=((List<?>)cls("ekg").getMethod("c").invoke(container)).size();
                        Object start=cls("ejz").getConstructor(cls("ejr"),cls("dcd"),int.class,cls("ekg")).newInstance(structure,chunkPos,0,container),saved=cls("ejz").getMethod("a",cls("ekh"),cls("dcd")).invoke(start,serialization,chunkPos);digest=hash(IglooProbeSupport.canonicalNBT(saved));
                    }
                    Object random=cls("ejr$a").getMethod("f").invoke(context);
                    System.out.printf("stronghold_start %d %d %d %s %d %s %d%n",seed,chunk[0],chunk[1],name,count,digest,(Long)cls("ayw").getMethod("g").invoke(random));
                }
            }
        }
    }
}
