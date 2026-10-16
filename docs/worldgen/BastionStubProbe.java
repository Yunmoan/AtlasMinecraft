package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.util.HashMap;
import java.util.Map;
import java.util.Optional;
import java.util.function.Predicate;
import java.util.stream.Stream;

/** Calls the original Bastion findGenerationPoint and biome-filtered stub;
 * does not execute the deferred Jigsaw assembly consumer. */
public final class BastionStubProbe {
    static Class<?> cls(String name)throws Exception{return Class.forName(name);}
    static int coord(Object p,String name)throws Exception{return (Integer)cls("kh").getMethod(name).invoke(p);}
    public static void main(String[] args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);
        Object registries=IglooProbeSupport.registryAccess(),manager=IglooProbeSupport.manager(new String[]{"bastion/units/air_base","bastion/hoglin_stable/air_base","bastion/treasure/big_air_full","bastion/bridge/starting_pieces/entrance_base"});
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Object lookup=cls("jo$a").getMethod("b").invoke(registries);var registry=cls("jn$a").getMethod("b",cls("akq"));
        Object structureLookup=registry.invoke(lookup,cls("lu").getField("aR").get(null)),biomeLookup=registry.invoke(lookup,cls("lu").getField("aF").get(null));
        Object structure=null;Map<String,Object> biomes=new HashMap<>();
        for(Object h:((Stream<?>)cls("jo").getMethod("b").invoke(structureLookup)).toList())if(cls("akq").getMethod("a").invoke(cls("jm$c").getMethod("h").invoke(h)).toString().equals("minecraft:bastion_remnant"))structure=cls("jm").getMethod("a").invoke(h);
        for(Object h:((Stream<?>)cls("jo").getMethod("b").invoke(biomeLookup)).toList())biomes.put(cls("akq").getMethod("a").invoke(cls("jm$c").getMethod("h").invoke(h)).toString(),h);
        Object key=cls("dzd").getField("f").get(null),settings=cls("jn").getMethod("b",cls("akq")).invoke(registry.invoke(lookup,cls("lu").getField("aO").get(null)),key),heights=cls("dcy").getMethod("e",int.class,int.class).invoke(null,0,256);
        for(long seed:new long[]{0,1,-1,1234,Long.MAX_VALUE,9212937586943976949L}) {
            Object noise=cls("dzm").getMethod("a",cls("jn$a"),cls("akq"),long.class).invoke(null,lookup,key,seed);
            for(String name:new String[]{"nether_wastes","soul_sand_valley","crimson_forest","warped_forest","basalt_deltas"}) {
                Object holder=biomes.get("minecraft:"+name),source=cls("deh").getConstructor(cls("jm")).newInstance(holder),generator=cls("dzb").getConstructor(cls("dea"),cls("jm")).newInstance(source,settings);
                for(int[] p:new int[][]{{0,0},{1,-1},{-1,1},{17,-9},{-41,23},{31,31},{-33,32},{145,-62},{-9,-15},{19,5},{-63,-27},{72,13},{5,6},{-8,-5},{101,-31},{-121,23}}) {
                    Object context=cls("ejr$a").getConstructor(cls("ka"),cls("duz"),cls("dea"),cls("dzm"),cls("enu"),long.class,cls("dcd"),cls("dcy"),Predicate.class).newInstance(registries,generator,source,noise,manager,seed,cls("dcd").getConstructor(int.class,int.class).newInstance(p[0],p[1]),heights,(Predicate<Object>)h->!name.equals("basalt_deltas"));
                    Object stub=((Optional<?>)cls("els").getMethod("a",cls("ejr$a")).invoke(structure,context)).orElseThrow(),position=cls("ejr$b").getMethod("b").invoke(stub);
                    Object consumer=((Optional<?>)cls("com.mojang.datafixers.util.Either").getMethod("left").invoke(cls("ejr$b").getMethod("c").invoke(stub))).orElseThrow(),piece=null;
                    for(var field:consumer.getClass().getDeclaredFields())if(field.getType()==cls("ejn")){field.setAccessible(true);piece=field.get(consumer);}
                    if(piece==null)throw new IllegalStateException("missing captured original start piece");
                    Object element=cls("ejn").getMethod("b").invoke(piece),origin=cls("ejn").getMethod("c").invoke(piece),rotation=cls("ejn").getMethod("a").invoke(piece),box=cls("ejv").getMethod("f").invoke(piece);
                    var location=cls("eky").getDeclaredField("c");location.setAccessible(true);Object either=location.get(element);
                    String template=((Optional<?>)cls("com.mojang.datafixers.util.Either").getMethod("left").invoke(either)).orElseThrow().toString();
                    System.out.printf("bastion_stub %d %d %d minecraft:%s %d %s %d %d %d %d",seed,p[0],p[1],name,name.equals("basalt_deltas")?0:1,template,((Enum<?>)rotation).ordinal(),coord(position,"u"),coord(position,"v"),coord(position,"w"));
                    System.out.printf(" %d %d %d",coord(origin,"u"),coord(origin,"v"),coord(origin,"w"));
                    for(String getter:new String[]{"h","i","j","k","l","m"})System.out.print(" "+cls("ejj").getMethod(getter).invoke(box));
                    Object random=cls("ejr$a").getMethod("f").invoke(context);System.out.println(" "+cls("ayw").getMethod("g").invoke(random));
                    Object checkedContext=cls("ejr$a").getConstructor(cls("ka"),cls("duz"),cls("dea"),cls("dzm"),cls("enu"),long.class,cls("dcd"),cls("dcy"),Predicate.class).newInstance(registries,generator,source,noise,manager,seed,cls("dcd").getConstructor(int.class,int.class).newInstance(p[0],p[1]),heights,(Predicate<Object>)h->!name.equals("basalt_deltas"));
                    boolean accepted=((Optional<?>)cls("ejr").getMethod("b",cls("ejr$a")).invoke(structure,checkedContext)).isPresent();
                    if(accepted==name.equals("basalt_deltas"))throw new IllegalStateException("original filtered stub disagrees");
                }
            }
        }
    }
}
