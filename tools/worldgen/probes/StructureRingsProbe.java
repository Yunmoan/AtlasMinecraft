package atlasprobe;

import com.google.gson.JsonParser;
import com.mojang.datafixers.util.Pair;
import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.InputStream;
import java.io.PrintStream;
import java.lang.reflect.Method;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.TreeSet;
import java.util.concurrent.CompletableFuture;
import java.util.function.Consumer;
import java.util.stream.Stream;

/** Calls the actual ChunkGeneratorStructureState ring generator with original
 * preferred-biome data and MultiNoiseBiomeSource, including biome relocation.
 */
public final class StructureRingsProbe {
    static Class<?> cls(String name) throws Exception { return Class.forName(name); }
    static TreeSet<String> biomeTag(String name) throws Exception {
        try(InputStream input=StructureRingsProbe.class.getClassLoader().getResourceAsStream("data/minecraft/tags/worldgen/biome/"+name+".json")) {
            if(input==null) throw new IllegalStateException("missing biome tag "+name);
            var json=JsonParser.parseString(new String(input.readAllBytes(),StandardCharsets.UTF_8)).getAsJsonObject();
            var result=new TreeSet<String>();
            for(var entry:json.getAsJsonArray("values")) {
                String value=entry.getAsString();
                if(value.startsWith("#minecraft:")) result.addAll(biomeTag(value.substring(11)));
                else result.add(value);
            }
            return result;
        }
    }
    public static void main(String[] args) throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);
        cls("ab").getMethod("a",cls("ae")).invoke(null,version);
        cls("akt").getMethod("a").invoke(null);
        Object registries=cls("pd").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Object lookup=cls("jo$a").getMethod("b").invoke(registries);
        Object biomeLookup=cls("jn$a").getMethod("b",cls("akq")).invoke(lookup,cls("lu").getField("aF").get(null));
        Map<String,Object> holders=new HashMap<>();
        for(Object holder:((Stream<?>)cls("jo").getMethod("b").invoke(biomeLookup)).toList()) {
            Object key=cls("jm$c").getMethod("h").invoke(holder);
            holders.put(cls("akq").getMethod("a").invoke(key).toString(),holder);
        }
        List<Pair<Object,Object>> points=new ArrayList<>();
        Method addPoints=cls("dem").getDeclaredMethod("a",Consumer.class);addPoints.setAccessible(true);
        addPoints.invoke(cls("dem").getConstructor().newInstance(),(Consumer<Pair<Object,Object>>)pair->{
            try {String name=cls("akq").getMethod("a").invoke(pair.getSecond()).toString();points.add(Pair.of(pair.getFirst(),holders.get(name)));}
            catch(Exception e){throw new RuntimeException(e);}
        });
        Object parameters=cls("def$c").getConstructor(List.class).newInstance(points);
        Object naturalBiomeSource=cls("dej").getMethod("a",cls("def$c")).invoke(null,parameters);
        Object fixedBiomeSource=cls("deh").getConstructor(cls("jm")).newInstance(holders.get("minecraft:plains"));
        var preferredHolders=new ArrayList<Object>();
        for(String name:biomeTag("stronghold_biased_to"))preferredHolders.add(holders.get(name));
        Object preferred=cls("jq").getMethod("a",List.class).invoke(null,preferredHolders);
        Object empty=cls("jq").getMethod("a",List.class).invoke(null,List.of());
        Method ringGenerator=cls("dva").getDeclaredMethod("a",cls("jm"),cls("ekl"));ringGenerator.setAccessible(true);
        var stateConstructor=cls("dva").getDeclaredConstructor(cls("dzm"),cls("dea"),long.class,long.class,List.class);stateConstructor.setAccessible(true);
        long[] seeds={0L,1L,-1L,1234L,Long.MAX_VALUE,9212937586943976949L};
        for(long seed:seeds) {
            Object randomState=cls("dzm").getMethod("a",cls("jn$a"),cls("akq"),long.class)
                .invoke(null,lookup,cls("dzd").getField("c").get(null),seed);
            String[] modes={"unrelocated","overworld"};
            for(String mode:modes) {
                Object placement=cls("ekl").getConstructor(int.class,int.class,int.class,cls("jq"))
                    .newInstance(32,3,128,mode.equals("overworld")?preferred:empty);
                Object state=stateConstructor.newInstance(randomState,mode.equals("overworld")?naturalBiomeSource:fixedBiomeSource,seed,seed,List.of());
                @SuppressWarnings("unchecked")
                List<Object> positions=(List<Object>)((CompletableFuture<?>)ringGenerator.invoke(state,null,placement)).join();
                for(int index=0;index<positions.size();index++) {
                    Object pos=positions.get(index);
                    System.out.printf("ring %d %s %d %d %d%n",seed,mode,index,cls("dcd").getField("e").getInt(pos),cls("dcd").getField("f").getInt(pos));
                }
            }
        }
    }
}
