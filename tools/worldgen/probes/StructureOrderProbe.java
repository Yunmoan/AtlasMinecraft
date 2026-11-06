package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.util.HashMap;
import java.util.Map;
import java.util.List;
import java.util.stream.Stream;

/** Original registry encounter order and WorldGenRegion positional random. */
public final class StructureOrderProbe {
    static Class<?> cls(String name)throws Exception{return Class.forName(name);}
    public static void main(String[] args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);
        cls("ab").getMethod("a",cls("ae")).invoke(null,version);
        cls("akt").getMethod("a").invoke(null);
        Object registries=cls("pd").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Object lookup=cls("jo$a").getMethod("b").invoke(registries);
        // VanillaRegistries.createLookup is an unordered data-generation lookup.
        // The running server uses RegistryDataLoader and MappedRegistry.stream.
        var createPack=cls("ats").getDeclaredMethod("b");createPack.setAccessible(true);
        Object pack=createPack.invoke(null);
        Object resources=cls("atx").getConstructor(cls("ass"),List.class).newInstance(cls("ass").getField("b").get(null),List.of(pack));
        Object builtin=cls("ka").getMethod("a",cls("jz")).invoke(null,cls("lt").getField("aA").get(null));
        Object loaded=cls("akm").getMethod("a",cls("aue"),cls("ka"),List.class).invoke(null,resources,builtin,cls("akm").getField("a").get(null));
        Object structures=cls("ka").getMethod("d",cls("akq")).invoke(loaded,cls("lu").getField("aR").get(null));
        Map<Integer,Integer> indices=new HashMap<>();
        for(Object structure:((Stream<?>)cls("jz").getMethod("t").invoke(structures)).toList()) {
            int step=((Enum<?>)cls("ejr").getMethod("c").invoke(structure)).ordinal();
            int index=indices.getOrDefault(step,0);indices.put(step,index+1);
            System.out.printf("order %s %d %d%n",cls("jz").getMethod("b",Object.class).invoke(structures,structure),step,index);
        }
        Object settings=cls("dzd").getField("c").get(null);
        Object salt=cls("akr").getMethod("b",String.class).invoke(null,"worldgen_region_random");
        for(long seed:new long[]{0,1,-1,1234,Long.MAX_VALUE,9212937586943976949L}) {
            Object state=cls("dzm").getMethod("a",cls("jn$a"),cls("akq"),long.class).invoke(null,lookup,settings,seed);
            Object factory=cls("dzm").getMethod("a",cls("akr")).invoke(state,salt);
            for(int[] chunk:new int[][]{{0,0},{1,-1},{-1,1},{17,-9},{-41,23},{31,31},{-33,32},{145,-62}}) {
                Object pos=cls("dcd").getConstructor(int.class,int.class).newInstance(chunk[0],chunk[1]);
                Object origin=cls("dcd").getMethod("l").invoke(pos);
                Object random=cls("dzl").getMethod("a",cls("jd")).invoke(factory,origin);
                System.out.printf("region_random %d %d %d",seed,chunk[0],chunk[1]);
                for(int index=0;index<8;index++)System.out.print(" "+cls("ayw").getMethod("g").invoke(random));
                System.out.println();
            }
        }
    }
}
