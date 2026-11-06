package atlasprobe;

import com.google.gson.Gson;
import java.io.*;
import java.lang.reflect.*;
import java.util.*;
import java.util.stream.Stream;

/** Invokes original liquid on-place, neighbor and shape callbacks. */
public final class LiquidBlockCallbackProbe {
    static Class<?> cls(String n)throws Exception{return Class.forName(n);}
    public static void main(String[]args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);RuinedPortalProbeSupport.bindTags();
        Class<?> level;try(InputStream input=FluidStepProbe.class.getResourceAsStream("/FluidStepLevel.class")){level=new FluidStepProbe.Loader().define(input.readAllBytes());}
        Field unsafeField=cls("sun.misc.Unsafe").getDeclaredField("theUnsafe");unsafeField.setAccessible(true);Object unsafe=unsafeField.get(null);Method allocate=cls("sun.misc.Unsafe").getMethod("allocateInstance",Class.class);
        Object registries=IglooProbeSupport.registryAccess(),lookup=cls("jo$a").getMethod("b").invoke(registries);
        Object dimensionLookup=cls("jn$a").getMethod("b",cls("akq")).invoke(lookup,cls("lu").getField("aK").get(null));Map<String,Object> dimensions=new HashMap<>();
        for(Object holder:((Stream<?>)cls("jo").getMethod("b").invoke(dimensionLookup)).toList()){Object key=((Optional<?>)cls("jm").getMethod("e").invoke(holder)).orElseThrow();dimensions.put(cls("akq").getMethod("a").invoke(key).toString(),cls("jm").getMethod("a").invoke(holder));}
        Gson gson=new Gson();System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        for(boolean warm:new boolean[]{false,true})for(String fluid:new String[]{"water","lava"})for(int amount=0;amount<16;amount++)for(int scene=0;scene<26;scene++)for(int hook=0;hook<3;hook++){
            Object world=allocate.invoke(unsafe,level);level.getField("dimension").set(world,dimensions.get(warm?"minecraft:the_nether":"minecraft:overworld"));int x=-39,y=-53,z=712,face=scene%6;
            level.getMethod("configureLiquidHook",int.class,String.class,int.class,int.class,int.class,int.class).invoke(world,scene,fluid,amount,x,y,z);
            Map<String,Object> row=new LinkedHashMap<>();row.put("warm",warm);row.put("fluid",fluid);row.put("level",amount);row.put("scene",scene);row.put("hook",hook);row.put("face",face);row.put("floor",level.getField("floor").get(world));row.put("origin",new int[]{x,y,z});row.put("inputs",level.getMethod("inputs").invoke(world));
            try{level.getMethod("liquidHook",int.class,int.class,int.class,int.class,int.class).invoke(world,hook,face,x,y,z);}catch(InvocationTargetException e){throw new IllegalStateException(gson.toJson(row),e.getCause());}
            row.put("events",level.getField("events").get(world));System.out.println("liquid_callback "+gson.toJson(row));
        }
    }
}
