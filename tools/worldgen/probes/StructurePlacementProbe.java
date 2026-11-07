package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.lang.reflect.Constructor;
import java.lang.reflect.Method;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.TreeMap;
import java.util.stream.Stream;

/** Calls original random-spread placement, all frequency reducers and exclusion
 * zones. Placements are potential starts, not structure block generation.
 */
public final class StructurePlacementProbe {
    static Class<?> cls(String name) throws Exception { return Class.forName(name); }

    public static void main(String[] args) throws Exception {
        Object version = cls("t").getMethod("a").invoke(null);
        cls("ab").getMethod("a", cls("ae")).invoke(null, version);
        cls("akt").getMethod("a").invoke(null);
        Object registries = cls("pd").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out), true));
        Object lookup = cls("jo$a").getMethod("b").invoke(registries);
        Object registry = cls("jn$a").getMethod("b", cls("akq"))
            .invoke(lookup, cls("lu").getField("aT").get(null));
        Map<String,Object> placements = new TreeMap<>();
        for (Object holder : ((Stream<?>)cls("jo").getMethod("b").invoke(registry)).toList()) {
            Object key = ((Optional<?>)cls("jm").getMethod("e").invoke(holder)).orElseThrow();
            String name = cls("akq").getMethod("a").invoke(key).toString();
            Object set = cls("jm").getMethod("a").invoke(holder);
            placements.put(name, cls("ejx").getMethod("b").invoke(set));
        }
        Constructor<?> stateConstructor = cls("dva").getDeclaredConstructor(cls("dzm"), cls("dea"), long.class, long.class, List.class);
        stateConstructor.setAccessible(true);
        Method potential = cls("ekm").getMethod("a", long.class, int.class, int.class);
        Method isStructure = cls("eko").getMethod("b", cls("dva"), int.class, int.class);
        Method frequency = cls("eko").getMethod("a", int.class, int.class, long.class);
        long[] seeds = {0L,1L,-1L,Long.MIN_VALUE,Long.MAX_VALUE,9212937586943976949L};
        int[][] chunks = {{0,0},{1,-1},{-1,1},{17,-9},{-41,23},{-32,-32},{31,31},{-33,32},
                          {-1024,4095},{1234567,-7654321},{Integer.MIN_VALUE,Integer.MAX_VALUE}};
        for (long seed : seeds) {
            Object state = stateConstructor.newInstance(null,null,seed,seed,List.of());
            for (var entry : placements.entrySet()) {
                Object placement = entry.getValue();
                if (!cls("ekm").isInstance(placement)) continue;
                for (int[] chunk : chunks) {
                    Object pos = potential.invoke(placement,seed,chunk[0],chunk[1]);
                    int x = cls("dcd").getField("e").getInt(pos), z = cls("dcd").getField("f").getInt(pos);
                    System.out.printf("potential %d %s %d %d %d %d%n",seed,entry.getKey(),chunk[0],chunk[1],x,z);
                    // Avoid exclusion-zone coordinate wrap at extreme ints;
                    // Minecraft's valid world coordinates never approach it.
                    if (Math.abs((long)x)<10000000 && Math.abs((long)z)<10000000) {
                        System.out.printf("placement %d %s %d %d %d%n",seed,entry.getKey(),x,z,
                            (Boolean)isStructure.invoke(placement,state,x,z)?1:0);
                        System.out.printf("placement %d %s %d %d %d%n",seed,entry.getKey(),x+1,z,
                            (Boolean)isStructure.invoke(placement,state,x+1,z)?1:0);
                    }
                }
                for (int x=-320; x<=320; x+=32) for (int z=-320; z<=320; z+=32) {
                    Object pos=potential.invoke(placement,seed,x,z);
                    int px=cls("dcd").getField("e").getInt(pos),pz=cls("dcd").getField("f").getInt(pos);
                    System.out.printf("placement %d %s %d %d %d%n",seed,entry.getKey(),px,pz,
                        (Boolean)isStructure.invoke(placement,state,px,pz)?1:0);
                }
            }
            String[] methods={"default","legacy_type_1","legacy_type_2","legacy_type_3"};
            String[] enumFields={"a","b","c","d"};
            float[] frequencies={0.0f,0.004f,0.01f,0.2f,0.5f,1.0f};
            for (int index=0; index<methods.length; index++) for (float value:frequencies) {
                Object method=cls("eko$c").getField(enumFields[index]).get(null);
                Object placement=cls("ekm").getConstructor(cls("kh"),cls("eko$c"),float.class,int.class,
                    Optional.class,int.class,int.class,cls("ekn"))
                    .newInstance(cls("kh").getField("g").get(null),method,value,987654321,Optional.empty(),34,8,cls("ekn").getField("a").get(null));
                for (int[] chunk:chunks) {
                    System.out.printf("frequency %d %s 987654321 %s %d %d %d%n",seed,methods[index],Float.toString(value),chunk[0],chunk[1],
                        (Boolean)frequency.invoke(placement,chunk[0],chunk[1],seed)?1:0);
                }
            }
        }
    }
}
