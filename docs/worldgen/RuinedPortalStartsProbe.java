package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.Base64;
import java.util.HashMap;
import java.util.HashSet;
import java.util.HexFormat;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.function.Predicate;
import java.util.stream.Stream;

/** Full original ruined portal start methods against actual NoiseBasedChunkGenerator
 * columns, with valid/cold/rejected biomes and Overworld/Nether settings. */
public final class RuinedPortalStartsProbe {
    static Class<?> cls(String n)throws Exception{return Class.forName(n);}
    static String digest(String text)throws Exception{return HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(text.getBytes(StandardCharsets.UTF_8)));}
    static void bindBiomeTags(Object registries)throws Exception {
        var createPack=cls("ats").getDeclaredMethod("b");createPack.setAccessible(true);
        Object pack=createPack.invoke(null),resources=cls("atx").getConstructor(cls("ass"),List.class).newInstance(cls("ass").getField("b").get(null),List.of(pack));
        Object registryKey=cls("lu").getField("aF").get(null),registry=cls("ka").getMethod("d",cls("akq")).invoke(registries,registryKey);
        java.util.function.Function<Object,java.util.Optional<?>> lookup=key->{try{return (java.util.Optional<?>)cls("jz").getMethod("c",cls("akr")).invoke(registry,key);}catch(Exception error){throw new IllegalStateException(error);}};
        Object loader=cls("awv").getConstructor(java.util.function.Function.class,String.class).newInstance(lookup,"tags/worldgen/biome");
        Map<?,?> loaded=(Map<?,?>)cls("awv").getMethod("b",cls("aue")).invoke(loader,resources);Map<Object,Object> tags=new HashMap<>();
        for(var entry:loaded.entrySet())tags.put(cls("awu").getMethod("a",cls("akq"),cls("akr")).invoke(null,registryKey,entry.getKey()),new java.util.ArrayList<>((java.util.Collection<?>)entry.getValue()));
        cls("jz").getMethod("a",Map.class).invoke(registry,tags);
    }
    public static void main(String[] args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);
        RuinedPortalProbeSupport.bindTags();Object registries=IglooProbeSupport.registryAccess(),manager=IglooProbeSupport.manager(RuinedPortalProbeSupport.names());
        bindBiomeTags(registries);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Object provider=cls("jo$a").getMethod("b").invoke(registries);
        var lookupRegistry=cls("jn$a").getMethod("b",cls("akq"));
        Object biomeLookup=lookupRegistry.invoke(provider,cls("lu").getField("aF").get(null)),structureLookup=lookupRegistry.invoke(provider,cls("lu").getField("aR").get(null)),settingsLookup=lookupRegistry.invoke(provider,cls("lu").getField("aO").get(null));
        Map<String,Object> biomes=new HashMap<>(),structures=new HashMap<>();
        for(Object holder:((Stream<?>)cls("jo").getMethod("b").invoke(biomeLookup)).toList()){Object key=cls("jm$c").getMethod("h").invoke(holder);biomes.put(cls("akq").getMethod("a").invoke(key).toString(),holder);}
        for(Object holder:((Stream<?>)cls("jo").getMethod("b").invoke(structureLookup)).toList()){Object key=cls("jm$c").getMethod("h").invoke(holder);structures.put(cls("akq").getMethod("a").invoke(key).toString(),cls("jm").getMethod("a").invoke(holder));}
        Object serialization=cls("ekh").getConstructor(cls("aue"),cls("ka"),cls("enu")).newInstance(null,registries,manager);
        Set<String> seenColumns=new HashSet<>();
        for(String[] family:new String[][]{{"ruined_portal","plains","snowy_plains","nether_wastes"},{"ruined_portal_desert","desert","plains"},{"ruined_portal_jungle","jungle","plains"},{"ruined_portal_mountain","jagged_peaks","plains"},{"ruined_portal_nether","nether_wastes","plains"},{"ruined_portal_ocean","cold_ocean","plains"},{"ruined_portal_swamp","swamp","plains"}})
        for(long seed:new long[]{0,1,-1,1234,Long.MAX_VALUE,9212937586943976949L})for(int biomeIndex=1;biomeIndex<family.length;biomeIndex++) {
            boolean nether=family[0].endsWith("nether");String dimension=nether?"nether":"overworld",biome="minecraft:"+family[biomeIndex];
            Object settingsKey=cls("dzd").getField(nether?"f":"c").get(null),settings=cls("jn").getMethod("b",cls("akq")).invoke(settingsLookup,settingsKey);
            Object randomState=cls("dzm").getMethod("a",cls("jn$a"),cls("akq"),long.class).invoke(null,provider,settingsKey,seed);
            Object biomeSource=cls("deh").getConstructor(cls("jm")).newInstance(biomes.get(biome)),generator=cls("dzb").getConstructor(cls("dea"),cls("jm")).newInstance(biomeSource,settings);
            int minY=nether?0:-64,height=nether?128:384;Object heights=cls("dcy").getMethod("e",int.class,int.class).invoke(null,minY,nether?256:384);
            Object structure=structures.get("minecraft:"+family[0]),allowed=cls("ejr").getMethod("a").invoke(structure);
            Predicate<Object> predicate=holder->{try{return (Boolean)cls("jq").getMethod("a",cls("jm")).invoke(allowed,holder);}catch(Exception error){throw new IllegalStateException(error);}};
            for(int[] chunk:new int[][]{{0,0},{1,-1},{-1,1},{17,-9},{-41,23},{31,31},{-33,32},{145,-62}}) {
                Object position=cls("dcd").getConstructor(int.class,int.class).newInstance(chunk[0],chunk[1]);
                Object start=cls("ejr").getMethod("a",cls("ka"),cls("duz"),cls("dea"),cls("dzm"),cls("enu"),long.class,cls("dcd"),int.class,cls("dcy"),Predicate.class).invoke(structure,registries,generator,biomeSource,randomState,manager,seed,position,0,heights,predicate);
                List<?> pieces=(List<?>)cls("ejz").getMethod("i").invoke(start);
                for(Object piece:pieces){
                    Object box=cls("ejv").getMethod("f").invoke(piece),center=cls("ejj").getMethod("g").invoke(box);
                    int minX=(Integer)cls("ejj").getMethod("h").invoke(box),maxX=(Integer)cls("ejj").getMethod("k").invoke(box),minZ=(Integer)cls("ejj").getMethod("j").invoke(box),maxZ=(Integer)cls("ejj").getMethod("m").invoke(box);
                    for(int[] column:new int[][]{{(Integer)cls("jd").getMethod("u").invoke(center),(Integer)cls("jd").getMethod("w").invoke(center)},{minX,minZ},{maxX,minZ},{minX,maxZ},{maxX,maxZ}}){
                        String key=dimension+" "+seed+" "+column[0]+" "+column[1];if(!seenColumns.add(key))continue;
                        Object base=cls("duz").getMethod("a",int.class,int.class,cls("dcy"),cls("dzm")).invoke(generator,column[0],column[1],heights,randomState);
                        byte[] materials=new byte[height];
                        for(int index=0;index<height;index++){Object state=cls("ddi").getMethod("a",int.class).invoke(base,minY+index),block=cls("dtb$a").getMethod("b").invoke(state);String name=cls("jz").getMethod("b",Object.class).invoke(cls("lt").getField("e").get(null),block).toString();materials[index]=switch(name){case "minecraft:air"->0;case "minecraft:stone"->1;case "minecraft:water"->2;case "minecraft:lava"->3;case "minecraft:granite"->4;case "minecraft:tuff"->5;case "minecraft:copper_ore"->6;case "minecraft:raw_copper_block"->7;case "minecraft:deepslate_iron_ore"->8;case "minecraft:raw_iron_block"->9;case "minecraft:netherrack"->10;case "minecraft:end_stone"->11;default->throw new IllegalStateException("unexpected base column block "+name);};}
                        System.out.printf("column %s %d %d %d %d %d %s%n",dimension,seed,column[0],column[1],minY,height,Base64.getEncoder().encodeToString(materials));
                    }
                }
                Object saved=cls("ejz").getMethod("a",cls("ekh"),cls("dcd")).invoke(start,serialization,position);
                System.out.printf("start %s %s %d %d %d %d %s%n",family[0],biome,seed,chunk[0],chunk[1],pieces.size(),digest(IglooProbeSupport.canonicalNBT(saved)));
            }
        }
    }
}
