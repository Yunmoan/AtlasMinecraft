package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.function.Predicate;
import java.util.stream.Stream;

/** Original IglooStructure starts, template sorting and saved piece NBT. */
public final class IglooStartsProbe {
    static Class<?> cls(String n)throws Exception{return Class.forName(n);}
    public static void main(String[] args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);
        cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);
        Object registries=cls("pd").getMethod("a").invoke(null),manager=IglooProbeSupport.manager();
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        for(String name:new String[]{"top","middle","bottom"}) {
            Object key=cls("akr").getMethod("a",String.class).invoke(null,"minecraft:igloo/"+name);
            Object template=cls("enu").getMethod("a",cls("akr")).invoke(manager,key);
            Object size=cls("ent").getMethod("a").invoke(template);
            var palettes=cls("ent").getDeclaredField("l");palettes.setAccessible(true);
            var entities=cls("ent").getDeclaredField("m");entities.setAccessible(true);
            List<?> blocks=(List<?>)cls("ent$a").getMethod("a").invoke(((List<?>)palettes.get(template)).get(0));
            System.out.printf("template %s %d %d %d %d %d%n",name,cls("kh").getMethod("u").invoke(size),cls("kh").getMethod("v").invoke(size),cls("kh").getMethod("w").invoke(size),blocks.size(),((List<?>)entities.get(template)).size());
            int index=0;
            for(Object block:blocks) {
                Object p=cls("ent$c").getMethod("a").invoke(block),state=cls("ent$c").getMethod("b").invoke(block);
                Object tag=cls("ent$c").getMethod("c").invoke(block);
                System.out.printf("template_block %s %d %d %d %d %d %s%n",name,index++,cls("jd").getMethod("u").invoke(p),cls("jd").getMethod("v").invoke(p),cls("jd").getMethod("w").invoke(p),cls("dfy").getMethod("i",cls("dtc")).invoke(null,state),tag==null?"none":cls("ub").getMethod("l",String.class).invoke(tag,"id"));
            }
        }
        Object lookup=cls("jo$a").getMethod("b").invoke(registries);
        var lookupRegistry=cls("jn$a").getMethod("b",cls("akq"));
        Object biomeLookup=lookupRegistry.invoke(lookup,cls("lu").getField("aF").get(null));
        Object structureLookup=lookupRegistry.invoke(lookup,cls("lu").getField("aR").get(null));
        Map<String,Object> biomes=new HashMap<>(),structures=new HashMap<>();
        for(Object holder:((Stream<?>)cls("jo").getMethod("b").invoke(biomeLookup)).toList()) {
            Object key=cls("jm$c").getMethod("h").invoke(holder);biomes.put(cls("akq").getMethod("a").invoke(key).toString(),holder);
        }
        for(Object holder:((Stream<?>)cls("jo").getMethod("b").invoke(structureLookup)).toList()) {
            Object key=cls("jm$c").getMethod("h").invoke(holder);structures.put(cls("akq").getMethod("a").invoke(key).toString(),cls("jm").getMethod("a").invoke(holder));
        }
        Object settingsKey=cls("dzd").getField("c").get(null);
        Object settingsLookup=lookupRegistry.invoke(lookup,cls("lu").getField("aO").get(null));
        Object settings=cls("jn").getMethod("b",cls("akq")).invoke(settingsLookup,settingsKey);
        Object heights=cls("dcy").getMethod("e",int.class,int.class).invoke(null,-64,384);
        var getHeight=cls("duz").getMethod("c",int.class,int.class,cls("dyy$a"),cls("dcy"),cls("dzm"));
        var generate=cls("ejr").getMethod("a",cls("ka"),cls("duz"),cls("dea"),cls("dzm"),cls("enu"),long.class,cls("dcd"),int.class,cls("dcy"),Predicate.class);
        for(long seed:new long[]{0,1,-1,1234,Long.MAX_VALUE,9212937586943976949L}) {
            Object randomState=cls("dzm").getMethod("a",cls("jn$a"),cls("akq"),long.class).invoke(null,lookup,settingsKey,seed);
            Object biomeSource=cls("deh").getConstructor(cls("jm")).newInstance(biomes.get("minecraft:snowy_plains"));
            Object generator=cls("dzb").getConstructor(cls("dea"),cls("jm")).newInstance(biomeSource,settings);
            for(int[] chunk:new int[][]{{0,0},{1,-1},{-1,1},{17,-9},{-41,23},{31,31},{-33,32},{145,-62}}) {
                int height=(Integer)getHeight.invoke(generator,chunk[0]*16+8,chunk[1]*16+8,cls("dyy$a").getField("a").get(null),heights,randomState);
                System.out.printf("height %d %d %d %d%n",seed,chunk[0],chunk[1],height);
                Object pos=cls("dcd").getConstructor(int.class,int.class).newInstance(chunk[0],chunk[1]);
                Object start=generate.invoke(structures.get("minecraft:igloo"),null,generator,biomeSource,randomState,manager,seed,pos,0,heights,(Predicate<Object>)holder->true);
                List<?> pieces=(List<?>)cls("ejz").getMethod("i").invoke(start);
                System.out.printf("start %d %d %d %d%n",seed,chunk[0],chunk[1],pieces.size());
                int index=0;
                for(Object piece:pieces) {
                    Object tag=cls("ejv").getMethod("a",cls("ekh")).invoke(piece,(Object)null);
                    System.out.printf("piece %d %d %d %d %s %s %s",seed,chunk[0],chunk[1],index++,cls("ub").getMethod("l",String.class).invoke(tag,"id"),cls("ub").getMethod("l",String.class).invoke(tag,"Template"),cls("ub").getMethod("l",String.class).invoke(tag,"Rot"));
                    for(String field:new String[]{"O","GD","TPX","TPY","TPZ"})System.out.print(" "+cls("ub").getMethod("h",String.class).invoke(tag,field));
                    for(int value:(int[])cls("ub").getMethod("n",String.class).invoke(tag,"BB"))System.out.print(" "+value);
                    System.out.println();
                }
            }
        }
    }
}
