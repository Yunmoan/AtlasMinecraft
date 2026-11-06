package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Stream;

/** Calls original ChunkGenerator.createStructures itself, with real original
 * portal starts, noise columns and biome acceptance. Controlled all-chunk
 * placement isolates weighted selection/removal from placement coverage. */
public final class StructureSetSelectionProbe {
    static Class<?> cls(String name)throws Exception{return Class.forName(name);}
    public static void main(String[] args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);
        RuinedPortalProbeSupport.bindTags();Object registries=IglooProbeSupport.registryAccess(),manager=IglooProbeSupport.manager(RuinedPortalProbeSupport.names());RuinedPortalStartsProbe.bindBiomeTags(registries);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Object provider=cls("jo$a").getMethod("b").invoke(registries);var lookup=cls("jn$a").getMethod("b",cls("akq"));
        Object biomeLookup=lookup.invoke(provider,cls("lu").getField("aF").get(null)),structureLookup=lookup.invoke(provider,cls("lu").getField("aR").get(null)),settingsLookup=lookup.invoke(provider,cls("lu").getField("aO").get(null));
        Map<String,Object> biomes=new HashMap<>(),structures=new HashMap<>();
        for(Object holder:((Stream<?>)cls("jo").getMethod("b").invoke(biomeLookup)).toList()){Object key=cls("jm$c").getMethod("h").invoke(holder);biomes.put(cls("akq").getMethod("a").invoke(key).toString(),holder);}
        for(Object holder:((Stream<?>)cls("jo").getMethod("b").invoke(structureLookup)).toList()){Object key=cls("jm$c").getMethod("h").invoke(holder);structures.put(cls("akq").getMethod("a").invoke(key).toString(),holder);}
        Object serialization=cls("ekh").getConstructor(cls("aue"),cls("ka"),cls("enu")).newInstance(null,registries,manager);
        Object structureManager=cls("ddq").getConstructors()[0].newInstance(null,null,null);
        Object placement=cls("ekm").getConstructor(int.class,int.class,cls("ekn"),int.class).newInstance(1,0,cls("ekn").getField("a").get(null),34222645);
        var stateConstructor=cls("dva").getDeclaredConstructor(cls("dzm"),cls("dea"),long.class,long.class,List.class);stateConstructor.setAccessible(true);
        String[] names={"ruined_portal","ruined_portal_desert","ruined_portal_jungle","ruined_portal_swamp","ruined_portal_mountain","ruined_portal_ocean","ruined_portal_nether"};
        for(long seed:new long[]{0,1,-1,1234,Long.MAX_VALUE,9212937586943976949L})
        for(String biomeName:new String[]{"plains","snowy_plains","desert","jungle","jagged_peaks","nether_wastes","cold_ocean","swamp","badlands","forest","mushroom_fields","end_highlands"}) {
            boolean nether=biomeName.equals("nether_wastes");String dimension=nether?"nether":"overworld",biome="minecraft:"+biomeName;
            Object settingsKey=cls("dzd").getField(nether?"f":"c").get(null),settings=cls("jn").getMethod("b",cls("akq")).invoke(settingsLookup,settingsKey);
            Object randomState=cls("dzm").getMethod("a",cls("jn$a"),cls("akq"),long.class).invoke(null,provider,settingsKey,seed);
            Object biomeSource=cls("deh").getConstructor(cls("jm")).newInstance(biomes.get(biome)),generator=cls("dzb").getConstructor(cls("dea"),cls("jm")).newInstance(biomeSource,settings);
            Object heights=cls("dcy").getMethod("e",int.class,int.class).invoke(null,nether?0:-64,nether?256:384),biomeRegistry=cls("ka").getMethod("d",cls("akq")).invoke(registries,cls("lu").getField("aF").get(null));
            for(int variant=0;variant<3;variant++) {
                List<Object> entries=new ArrayList<>();for(int index=0;index<(variant==2?1:names.length);index++)entries.add(cls("ejx$a").getConstructor(cls("jm"),int.class).newInstance(structures.get("minecraft:"+names[index]),variant==1?index+1:1));
                Object set=cls("ejx").getConstructor(List.class,cls("eko")).newInstance(entries,placement),holder=cls("jm").getMethod("a",Object.class).invoke(null,set),state=stateConstructor.newInstance(randomState,biomeSource,seed,seed,List.of(holder));
                for(int[] position:new int[][]{{0,0},{1,-1},{-1,1},{17,-9},{-41,23},{31,31},{-33,32},{145,-62}}) {
                    Object chunkPos=cls("dcd").getConstructor(int.class,int.class).newInstance(position[0],position[1]),chunk=cls("dvs").getConstructor(cls("dcd"),cls("dvv"),cls("dcy"),cls("jz"),cls("eab")).newInstance(chunkPos,cls("dvv").getField("a").get(null),heights,biomeRegistry,null);
                    cls("duz").getMethod("a",cls("ka"),cls("dva"),cls("ddq"),cls("duy"),cls("enu")).invoke(generator,registries,state,structureManager,chunk,manager);
                    Map<?,?> starts=(Map<?,?>)cls("duy").getMethod("g").invoke(chunk);Object start=null;
                    for(Object candidate:starts.values())if((Boolean)cls("ejz").getMethod("b").invoke(candidate)){if(start!=null)throw new IllegalStateException("multiple starts in original set");start=candidate;}
                    String selected="INVALID",hash="none";
                    if(start!=null){Object saved=cls("ejz").getMethod("a",cls("ekh"),cls("dcd")).invoke(start,serialization,chunkPos);selected=(String)cls("ub").getMethod("l",String.class).invoke(saved,"id");hash=RuinedPortalStartsProbe.digest(IglooProbeSupport.canonicalNBT(saved));}
                    System.out.printf("set %d %s %s %d %d %d %s %s%n",variant,dimension,biome,seed,position[0],position[1],selected,hash);
                }
            }
        }
    }
}
