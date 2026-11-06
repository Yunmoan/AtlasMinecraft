package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.function.Predicate;
import java.util.stream.Stream;

/** Calls original DesertPyramidStructure / SwampHutStructure.generate and
 * serializes their original unprocessed pieces. Heights come from the actual
 * NoiseBasedChunkGenerator rather than a hand-written terrain substitute.
 */
public final class StructureStartsProbe {
    static Class<?> cls(String name) throws Exception { return Class.forName(name); }
    public static void main(String[] args) throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);
        cls("ab").getMethod("a",cls("ae")).invoke(null,version);
        cls("akt").getMethod("a").invoke(null);
        Object registries=cls("pd").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Object lookup=cls("jo$a").getMethod("b").invoke(registries);
        var lookupRegistry=cls("jn$a").getMethod("b",cls("akq"));
        Object biomeLookup=lookupRegistry.invoke(lookup,cls("lu").getField("aF").get(null));
        Object structureLookup=lookupRegistry.invoke(lookup,cls("lu").getField("aR").get(null));
        Map<String,Object> biomes=new HashMap<>(),structures=new HashMap<>();
        for(Object holder:((Stream<?>)cls("jo").getMethod("b").invoke(biomeLookup)).toList()) {
            Object key=cls("jm$c").getMethod("h").invoke(holder);
            biomes.put(cls("akq").getMethod("a").invoke(key).toString(),holder);
        }
        for(Object holder:((Stream<?>)cls("jo").getMethod("b").invoke(structureLookup)).toList()) {
            Object key=cls("jm$c").getMethod("h").invoke(holder);
            structures.put(cls("akq").getMethod("a").invoke(key).toString(),cls("jm").getMethod("a").invoke(holder));
        }
        Object settingsKey=cls("dzd").getField("c").get(null);
        Object settingsLookup=lookupRegistry.invoke(lookup,cls("lu").getField("aO").get(null));
        Object settings=cls("jn").getMethod("b",cls("akq")).invoke(settingsLookup,settingsKey);
        Object heights=cls("dcy").getMethod("e",int.class,int.class).invoke(null,-64,384);
        Object worldSurface=cls("dyy$a").getField("a").get(null);
        var getHeight=cls("duz").getMethod("c",int.class,int.class,cls("dyy$a"),cls("dcy"),cls("dzm"));
        var generate=cls("ejr").getMethod("a",cls("ka"),cls("duz"),cls("dea"),cls("dzm"),cls("enu"),long.class,cls("dcd"),int.class,cls("dcy"),Predicate.class);
        long[] seeds={0L,1L,-1L,1234L,Long.MAX_VALUE,9212937586943976949L};
        int[][] chunks={{0,0},{1,-1},{-1,1},{17,-9},{-41,23},{31,31},{-33,32},{145,-62}};
        for(long seed:seeds) {
            Object randomState=cls("dzm").getMethod("a",cls("jn$a"),cls("akq"),long.class).invoke(null,lookup,settingsKey,seed);
            for(String name:new String[]{"minecraft:desert_pyramid","minecraft:swamp_hut","minecraft:jungle_pyramid"}) {
                String validBiome=name.endsWith("desert_pyramid")?"minecraft:desert":name.endsWith("jungle_pyramid")?"minecraft:jungle":"minecraft:swamp";
                Object biomeSource=cls("deh").getConstructor(cls("jm")).newInstance(biomes.get(validBiome));
                Object generator=cls("dzb").getConstructor(cls("dea"),cls("jm")).newInstance(biomeSource,settings);
                for(int[] chunk:chunks) {
                    int bx=chunk[0]*16,bz=chunk[1]*16;
                    int width=name.endsWith("jungle_pyramid")?12:21,length=name.endsWith("jungle_pyramid")?15:21;
                    for(int[] offset:new int[][]{{0,0},{0,length},{width,0},{width,length},{8,8}}) {
                        int y=(Integer)getHeight.invoke(generator,bx+offset[0],bz+offset[1],worldSurface,heights,randomState);
                        System.out.printf("height %d %s %d %d %d %d %d%n",seed,name,chunk[0],chunk[1],bx+offset[0],bz+offset[1],y);
                    }
                    Object pos=cls("dcd").getConstructor(int.class,int.class).newInstance(chunk[0],chunk[1]);
                    for(boolean valid:new boolean[]{true,false}) {
                        Predicate<Object> biomePredicate=holder->valid;
                        Object start=generate.invoke(structures.get(name),null,generator,biomeSource,randomState,null,seed,pos,0,heights,biomePredicate);
                        boolean accepted=(Boolean)cls("ejz").getMethod("b").invoke(start);
                        System.out.printf("start %d %s %d %d %s %d%n",seed,name,chunk[0],chunk[1],valid?validBiome:"minecraft:plains",accepted?1:0);
                        if(!accepted)continue;
                        for(Object piece:(List<?>)cls("ejz").getMethod("i").invoke(start)) {
                            Object tag=cls("ejv").getMethod("a",cls("ekh")).invoke(piece,(Object)null);
                            int[] box=(int[])cls("ub").getMethod("n",String.class).invoke(tag,"BB");
                            String id=(String)cls("ub").getMethod("l",String.class).invoke(tag,"id");
                            int orientation=(Integer)cls("ub").getMethod("h",String.class).invoke(tag,"O");
                            int depth=(Integer)cls("ub").getMethod("h",String.class).invoke(tag,"GD");
                            System.out.printf("piece %d %s %d %d %s %d %d",seed,name,chunk[0],chunk[1],id,orientation,depth);
                            for(int value:box)System.out.print(" "+value);
                            for(String field:new String[]{"Width","Height","Depth","HPos"})System.out.print(" "+cls("ub").getMethod("h",String.class).invoke(tag,field));
                            for(String field:name.endsWith("desert_pyramid")?new String[]{"hasPlacedChest0","hasPlacedChest1","hasPlacedChest2","hasPlacedChest3"}:name.endsWith("jungle_pyramid")?new String[]{"placedMainChest","placedHiddenChest","placedTrap1","placedTrap2"}:new String[]{"Witch","Cat"})
                                System.out.print(" "+((Boolean)cls("ub").getMethod("q",String.class).invoke(tag,field)?1:0));
                            System.out.println();
                        }
                    }
                }
            }
        }
    }
}
