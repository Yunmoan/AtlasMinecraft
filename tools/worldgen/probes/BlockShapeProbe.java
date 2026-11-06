package atlasprobe;

import com.google.gson.Gson;
import java.io.*;
import java.lang.reflect.*;
import java.util.*;
import java.util.stream.Stream;

/** Original updateShape for every state, face and defined neighbor scene. */
public final class BlockShapeProbe {
    static Class<?> cls(String n)throws Exception{return Class.forName(n);}
    public static void main(String[]args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);RuinedPortalProbeSupport.bindTags();
        Class<?> level;try(InputStream input=FluidStepProbe.class.getResourceAsStream("/FluidStepLevel.class")){level=new FluidStepProbe.Loader().define(input.readAllBytes());}
        Field unsafeField=cls("sun.misc.Unsafe").getDeclaredField("theUnsafe");unsafeField.setAccessible(true);Object unsafe=unsafeField.get(null);Object world=cls("sun.misc.Unsafe").getMethod("allocateInstance",Class.class).invoke(unsafe,level);
        Object registries=IglooProbeSupport.registryAccess(),lookup=cls("jo$a").getMethod("b").invoke(registries);
        Object dimensionLookup=cls("jn$a").getMethod("b",cls("akq")).invoke(lookup,cls("lu").getField("aK").get(null));
        for(Object holder:((Stream<?>)cls("jo").getMethod("b").invoke(dimensionLookup)).toList()){Object key=((Optional<?>)cls("jm").getMethod("e").invoke(holder)).orElseThrow();if(cls("akq").getMethod("a").invoke(key).toString().equals("minecraft:overworld"))level.getField("dimension").set(world,cls("jm").getMethod("a").invoke(holder));}
        boolean plants=args.length>0 && args[0].equals("plants");
        Method run=level.getMethod(plants?"updatePlantShapeCase":"updateShapeCase",int.class,int.class,int.class),next=cls("ayw").getMethod("g"),isPlant=level.getMethod("plantShapeState",int.class);Field events=level.getField("events"),random=level.getField("random");
        Gson gson=new Gson();System.setOut(new PrintStream(new BufferedOutputStream(new FileOutputStream(FileDescriptor.out),65536),false));System.setErr(new PrintStream(new FileOutputStream(FileDescriptor.err),true));
        for(int state=0;state<26684;state++)if(!plants || (boolean)isPlant.invoke(null,state))for(int face=0;face<6;face++)for(int scene=0;scene<(plants?25:6);scene++) {
            Map<String,Object> row=new LinkedHashMap<>();row.put("state",state);row.put("face",face);row.put("scene",scene);
            try { row.put("result",run.invoke(world,state,face,scene)); }
            catch(InvocationTargetException e) {System.out.flush();throw new IllegalStateException(gson.toJson(row),e.getCause());}
            row.put("events",events.get(world));row.put("next",Long.toUnsignedString((long)next.invoke(random.get(world))));
            if(plants){row.put("inputs",level.getField("shapeInputs").get(world));row.put("brightness",level.getField("shapeBrightness").get(world));}
            System.out.println("block_shape "+gson.toJson(row));
        }
        System.out.flush();
    }
}
