package atlasprobe;

import java.io.*;
import java.util.*;
import java.util.function.Predicate;
import java.util.stream.Stream;

/** Original mineshaft assembler and full start gate, with actual base terrain. */
public final class MineshaftStartsProbe {
    static Class<?> cls(String name)throws Exception{return Class.forName(name);}
    public static void main(String[]args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);
        Object registries=IglooProbeSupport.registryAccess();RuinedPortalStartsProbe.bindBiomeTags(registries);
        Object provider=cls("jo$a").getMethod("b").invoke(registries),biomeLookup=cls("jn$a").getMethod("b",cls("akq")).invoke(provider,cls("lu").getField("aF").get(null)),structureLookup=cls("jn$a").getMethod("b",cls("akq")).invoke(provider,cls("lu").getField("aR").get(null)),settingsLookup=cls("jn$a").getMethod("b",cls("akq")).invoke(provider,cls("lu").getField("aO").get(null));
        Map<String,Object> biomes=new HashMap<>(),structures=new HashMap<>();
        for(Object holder:((Stream<?>)cls("jo").getMethod("b").invoke(biomeLookup)).toList()){Object key=cls("jm$c").getMethod("h").invoke(holder);biomes.put(cls("akq").getMethod("a").invoke(key).toString(),holder);}
        for(Object holder:((Stream<?>)cls("jo").getMethod("b").invoke(structureLookup)).toList()){Object key=cls("jm$c").getMethod("h").invoke(holder);structures.put(cls("akq").getMethod("a").invoke(key).toString(),cls("jm").getMethod("a").invoke(holder));}
        Object settingsKey=cls("dzd").getField("c").get(null),settings=cls("jn").getMethod("b",cls("akq")).invoke(settingsLookup,settingsKey),heights=cls("dcy").getMethod("e",int.class,int.class).invoke(null,-64,384),manager=IglooProbeSupport.manager(new String[]{}),serialization=cls("ekh").getConstructor(cls("aue"),cls("ka"),cls("enu")).newInstance(null,registries,manager);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        for(long seed:new long[]{0,1,-1,1234,Long.MIN_VALUE,Long.MAX_VALUE,9212937586943976949L}) {
            Object randomState=cls("dzm").getMethod("a",cls("jn$a"),cls("akq"),long.class).invoke(null,provider,settingsKey,seed);
            for(int mesa=0;mesa<2;mesa++)for(int[]chunk:new int[][]{{0,0},{1,-1},{-1,1},{17,-9},{-41,23},{31,31},{-33,32},{145,-62}}) {
                String family=mesa==0?"mineshaft":"mineshaft_mesa",validBiome=mesa==0?"minecraft:plains":"minecraft:badlands";
                Object structure=structures.get("minecraft:"+family),biomeSource=cls("deh").getConstructor(cls("jm")).newInstance(biomes.get(validBiome)),generator=cls("dzb").getConstructor(cls("dea"),cls("jm")).newInstance(biomeSource,settings),random=cls("dyz").getConstructor(long.class).newInstance(seed);
                long xm=(Long)cls("ayw").getMethod("g").invoke(random),zm=(Long)cls("ayw").getMethod("g").invoke(random);cls("ayw").getMethod("b",long.class).invoke(random,(long)chunk[0]*xm^(long)chunk[1]*zm^seed);cls("ayw").getMethod("j").invoke(random);
                Object builder=cls("ekj").getConstructor().newInstance(),room=cls("elv$d").getConstructor(int.class,cls("ayw"),int.class,int.class,cls("elw$a")).newInstance(0,random,chunk[0]*16+2,chunk[1]*16+2,cls("elw$a").getEnumConstants()[mesa]);
                cls("ejw").getMethod("a",cls("ejv")).invoke(builder,room);cls("ejv").getMethod("a",cls("ejv"),cls("ejw"),cls("ayw")).invoke(room,room,builder,random);
                int surfaceX=0,surfaceZ=0,surfaceHeight=0,shift;
                if(mesa==0)shift=(Integer)cls("ekj").getMethod("a",int.class,int.class,cls("ayw"),int.class).invoke(builder,63,-64,random,10);
                else {
                    Object box=cls("ekj").getMethod("d").invoke(builder),center=cls("ejj").getMethod("g").invoke(box);surfaceX=(Integer)cls("jd").getMethod("u").invoke(center);surfaceZ=(Integer)cls("jd").getMethod("w").invoke(center);
                    surfaceHeight=(Integer)cls("duz").getMethod("a",int.class,int.class,cls("dyy$a"),cls("dcy"),cls("dzm")).invoke(generator,surfaceX,surfaceZ,cls("dyy$a").getField("a").get(null),heights,randomState);
                    int target=surfaceHeight<=63?63:(Integer)cls("ayo").getMethod("b",cls("ayw"),int.class,int.class).invoke(null,random,63,surfaceHeight);shift=target-(Integer)cls("jd").getMethod("v").invoke(center);cls("ekj").getMethod("a",int.class).invoke(builder,shift);
                }
                Object container=cls("ekj").getMethod("a").invoke(builder);List<?>pieces=(List<?>)cls("ekg").getMethod("c").invoke(container);StringBuilder pieceNBT=new StringBuilder();
                for(Object piece:pieces)pieceNBT.append(IglooProbeSupport.canonicalNBT(cls("ejv").getMethod("a",cls("ekh")).invoke(piece,serialization))).append('\n');
                System.out.printf("mineslayout %d %d %d %d %d %d %d %d %d %s %d%n",seed,chunk[0],chunk[1],mesa,surfaceX,surfaceZ,surfaceHeight,shift,pieces.size(),RuinedPortalStartsProbe.digest(pieceNBT.toString()),cls("ayw").getMethod("g").invoke(random));
                for(String biome:new String[]{validBiome,"minecraft:the_void"}) {
                    Object source=cls("deh").getConstructor(cls("jm")).newInstance(biomes.get(biome)),terrain=cls("dzb").getConstructor(cls("dea"),cls("jm")).newInstance(source,settings),allowed=cls("ejr").getMethod("a").invoke(structure);
                    Predicate<Object>predicate=holder->{try{return(Boolean)cls("jq").getMethod("a",cls("jm")).invoke(allowed,holder);}catch(Exception error){throw new IllegalStateException(error);}};
                    Object position=cls("dcd").getConstructor(int.class,int.class).newInstance(chunk[0],chunk[1]),start=cls("ejr").getMethod("a",cls("ka"),cls("duz"),cls("dea"),cls("dzm"),cls("enu"),long.class,cls("dcd"),int.class,cls("dcy"),Predicate.class).invoke(structure,registries,terrain,source,randomState,manager,seed,position,0,heights,predicate),saved=cls("ejz").getMethod("a",cls("ekh"),cls("dcd")).invoke(start,serialization,position);
                    List<?>fullPieces=(List<?>)cls("ejz").getMethod("i").invoke(start);System.out.printf("minesstart %s %s %d %d %d %d %s%n",family,biome,seed,chunk[0],chunk[1],fullPieces.size(),RuinedPortalStartsProbe.digest(IglooProbeSupport.canonicalNBT(saved)));
                }
            }
        }
    }
}
