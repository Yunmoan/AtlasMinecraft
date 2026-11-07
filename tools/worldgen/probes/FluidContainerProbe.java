package atlasprobe;

import com.google.gson.Gson;
import java.io.*;
import java.lang.reflect.*;
import java.util.*;
import java.util.stream.Stream;

/** Every registered container block state, through the original placeLiquid. */
public final class FluidContainerProbe {
    static Class<?> cls(String n)throws Exception{return Class.forName(n);}
    public static void main(String[]args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);RuinedPortalProbeSupport.bindTags();
        Class<?> level;try(InputStream input=FluidStepProbe.class.getResourceAsStream("/FluidStepLevel.class")){level=new FluidStepProbe.Loader().define(input.readAllBytes());}
        Field unsafeField=cls("sun.misc.Unsafe").getDeclaredField("theUnsafe");unsafeField.setAccessible(true);Object unsafe=unsafeField.get(null);Method allocate=cls("sun.misc.Unsafe").getMethod("allocateInstance",Class.class);
        Object registries=IglooProbeSupport.registryAccess(),lookup=cls("jo$a").getMethod("b").invoke(registries);
        Object dimensionLookup=cls("jn$a").getMethod("b",cls("akq")).invoke(lookup,cls("lu").getField("aK").get(null));Object dimension=null;
        for(Object holder:((Stream<?>)cls("jo").getMethod("b").invoke(dimensionLookup)).toList()){Object key=((Optional<?>)cls("jm").getMethod("e").invoke(holder)).orElseThrow();if(cls("akq").getMethod("a").invoke(key).toString().equals("minecraft:overworld"))dimension=cls("jm").getMethod("a").invoke(holder);}
        Gson gson=new Gson();System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        System.out.println("container_sounds "+gson.toJson(level.getMethod("containerSounds").invoke(null)));
        for(Object packet:(List<?>)level.getMethod("containerSoundPackets").invoke(null))System.out.println("container_sound_packet "+packet);
        Method decode=cls("dfy").getMethod("a",int.class),block=cls("dtb$a").getMethod("b");
        for(int id=0;id<26684;id++) {
            if(!cls("dkp").isInstance(block.invoke(decode.invoke(null,id))))continue;
            for(int kind=0;kind<5;kind++) {
                Object world=allocate.invoke(unsafe,level);level.getField("dimension").set(world,dimension);
                long seed=id*31L+kind;level.getMethod("configureContainer",int.class,long.class).invoke(world,id,seed);
                Map<String,Object> row=new LinkedHashMap<>();row.put("state",id);row.put("kind",kind);row.put("seed",seed);
                try{row.put("accepted",level.getMethod("placeContainer",int.class).invoke(world,kind));}
                catch(InvocationTargetException e){throw new IllegalStateException(gson.toJson(row),e.getCause());}
                row.put("events",level.getField("events").get(world));Object random=level.getField("random").get(world);row.put("next",Long.toUnsignedString((long)cls("ayw").getMethod("g").invoke(random)));
                System.out.println("fluid_container "+gson.toJson(row));
            }
        }
    }
}
