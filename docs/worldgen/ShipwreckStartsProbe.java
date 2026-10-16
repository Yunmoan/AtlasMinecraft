package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.function.Predicate;
import java.util.stream.Stream;

/** Original ocean/beached ShipwreckStructure.generate with real base-height
 * queries and official biome tags, comparing complete saved starts. */
public final class ShipwreckStartsProbe {
    static Class<?> cls(String name)throws Exception{return Class.forName(name);}
    public static void main(String[] args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);
        Object registries=IglooProbeSupport.registryAccess(),manager=IglooProbeSupport.manager(ShipwreckProbeSupport.names());RuinedPortalStartsProbe.bindBiomeTags(registries);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));Object provider=cls("jo$a").getMethod("b").invoke(registries);var lookup=cls("jn$a").getMethod("b",cls("akq"));
        Object biomeLookup=lookup.invoke(provider,cls("lu").getField("aF").get(null)),structureLookup=lookup.invoke(provider,cls("lu").getField("aR").get(null)),settingsLookup=lookup.invoke(provider,cls("lu").getField("aO").get(null));
        Map<String,Object> biomes=new HashMap<>(),structures=new HashMap<>();for(Object holder:((Stream<?>)cls("jo").getMethod("b").invoke(biomeLookup)).toList()){Object key=cls("jm$c").getMethod("h").invoke(holder);biomes.put(cls("akq").getMethod("a").invoke(key).toString(),holder);}for(Object holder:((Stream<?>)cls("jo").getMethod("b").invoke(structureLookup)).toList()){Object key=cls("jm$c").getMethod("h").invoke(holder);structures.put(cls("akq").getMethod("a").invoke(key).toString(),cls("jm").getMethod("a").invoke(holder));}
        Object serialization=cls("ekh").getConstructor(cls("aue"),cls("ka"),cls("enu")).newInstance(null,registries,manager),settingsKey=cls("dzd").getField("c").get(null),settings=cls("jn").getMethod("b",cls("akq")).invoke(settingsLookup,settingsKey),heights=cls("dcy").getMethod("e",int.class,int.class).invoke(null,-64,384);
        for(String[] family:new String[][]{{"shipwreck","ocean","deep_cold_ocean","plains"},{"shipwreck_beached","beach","snowy_beach","plains"}})for(long seed:new long[]{0,1,-1,1234,Long.MAX_VALUE,9212937586943976949L})for(int biomeIndex=1;biomeIndex<family.length;biomeIndex++) {
            Object randomState=cls("dzm").getMethod("a",cls("jn$a"),cls("akq"),long.class).invoke(null,provider,settingsKey,seed),biomeSource=cls("deh").getConstructor(cls("jm")).newInstance(biomes.get("minecraft:"+family[biomeIndex])),generator=cls("dzb").getConstructor(cls("dea"),cls("jm")).newInstance(biomeSource,settings),structure=structures.get("minecraft:"+family[0]),allowed=cls("ejr").getMethod("a").invoke(structure),heightmap=cls("dyy$a").getField(family[0].endsWith("beached")?"a":"c").get(null);
            Predicate<Object> predicate=holder->{try{return (Boolean)cls("jq").getMethod("a",cls("jm")).invoke(allowed,holder);}catch(Exception error){throw new IllegalStateException(error);}};
            for(int[] chunk:new int[][]{{0,0},{1,-1},{-1,1},{17,-9},{-41,23},{31,31},{-33,32},{145,-62}}) {
                int height=(Integer)cls("duz").getMethod("c",int.class,int.class,cls("dyy$a"),cls("dcy"),cls("dzm")).invoke(generator,chunk[0]*16+8,chunk[1]*16+8,heightmap,heights,randomState);
                Object position=cls("dcd").getConstructor(int.class,int.class).newInstance(chunk[0],chunk[1]),start=cls("ejr").getMethod("a",cls("ka"),cls("duz"),cls("dea"),cls("dzm"),cls("enu"),long.class,cls("dcd"),int.class,cls("dcy"),Predicate.class).invoke(structure,registries,generator,biomeSource,randomState,manager,seed,position,0,heights,predicate),saved=cls("ejz").getMethod("a",cls("ekh"),cls("dcd")).invoke(start,serialization,position);List<?> pieces=(List<?>)cls("ejz").getMethod("i").invoke(start);
                System.out.printf("shipstart %s minecraft:%s %d %d %d %d %d %s%n",family[0],family[biomeIndex],seed,chunk[0],chunk[1],height,pieces.size(),RuinedPortalStartsProbe.digest(IglooProbeSupport.canonicalNBT(saved)));
            }
        }
    }
}
