package atlasprobe;

import com.google.gson.Gson;
import java.io.*;
import java.lang.reflect.*;
import java.util.*;
import java.util.stream.Stream;

public final class FluidStepProbe {
    static Class<?> cls(String n)throws Exception{return Class.forName(n);}
    static class Loader extends ClassLoader {
        Loader(){super(FluidStepProbe.class.getClassLoader());}
        Class<?> define(byte[] bytes){return defineClass("FluidStepLevel",bytes,0,bytes.length);}
    }
    public static void main(String[]args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);
        RuinedPortalProbeSupport.bindTags();
        Class<?> level;try(InputStream input=FluidStepProbe.class.getResourceAsStream("/FluidStepLevel.class")){level=new Loader().define(input.readAllBytes());}
        Field unsafeField=cls("sun.misc.Unsafe").getDeclaredField("theUnsafe");unsafeField.setAccessible(true);Object unsafe=unsafeField.get(null);
        Method allocate=cls("sun.misc.Unsafe").getMethod("allocateInstance",Class.class);
        Object registries=IglooProbeSupport.registryAccess(),lookup=cls("jo$a").getMethod("b").invoke(registries);
        Object dimensionLookup=cls("jn$a").getMethod("b",cls("akq")).invoke(lookup,cls("lu").getField("aK").get(null));
        Map<String,Object> dimensions=new HashMap<>();
        for(Object holder:((Stream<?>)cls("jo").getMethod("b").invoke(dimensionLookup)).toList()) {
            Object key=((Optional<?>)cls("jm").getMethod("e").invoke(holder)).orElseThrow();
            dimensions.put(cls("akq").getMethod("a").invoke(key).toString(),cls("jm").getMethod("a").invoke(holder));
        }
        Gson gson=new Gson();System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        for(boolean warm:new boolean[]{false,true})for(boolean convert:new boolean[]{false,true})for(long seed:new long[]{0,1234,4096})for(String fluid:new String[]{"water","lava"})for(int amount=0;amount<16;amount++)for(int scene=0;scene<28;scene++) {
            Object world=allocate.invoke(unsafe,level);level.getField("dimension").set(world,dimensions.get(warm?"minecraft:the_nether":"minecraft:overworld"));
            int x=-39,y=-53,z=712;
            level.getMethod("configure",long.class,boolean.class,int.class,String.class,int.class,int.class,int.class,int.class).invoke(world,seed,convert,scene,fluid,amount,x,y,z);
            Map<String,Object> row=new TreeMap<>();row.put("warm",warm);row.put("convert",convert);row.put("seed",seed);row.put("scene",scene);row.put("fluid",fluid);row.put("level",amount);row.put("origin",new int[]{x,y,z});row.put("floor",level.getField("floor").get(world));row.put("inputs",level.getMethod("inputs").invoke(world));
            try {row.put("next",Long.toUnsignedString((long)level.getMethod("step",int.class,int.class,int.class).invoke(world,x,y,z)));}
            catch(InvocationTargetException e){throw new IllegalStateException(gson.toJson(row),e.getCause());}
            row.put("events",level.getField("events").get(world));System.out.println("fluid_step "+gson.toJson(row));
        }
    }
}
