package atlasprobe;

import com.google.gson.Gson;
import java.io.*;
import java.lang.reflect.*;
import java.util.*;

public final class BlockTickProbe {
    static Class<?> cls(String name)throws Exception{return Class.forName(name);}
    static final class Loader extends ClassLoader {Class<?> define(byte[] data){return defineClass("BlockTickLevel",data,0,data.length);}}
    public static void main(String[]args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);RuinedPortalProbeSupport.bindTags();
        Class<?> level;try(InputStream input=BlockTickProbe.class.getResourceAsStream("/BlockTickLevel.class")){level=new Loader().define(input.readAllBytes());}
        Field f=cls("sun.misc.Unsafe").getDeclaredField("theUnsafe");f.setAccessible(true);Object unsafe=f.get(null),world=cls("sun.misc.Unsafe").getMethod("allocateInstance",Class.class).invoke(unsafe,level);
        Method handler=level.getMethod("handler",int.class),distance=level.getMethod("leafDistance",int.class),name=level.getMethod("blockName",int.class),configure=level.getMethod("configure",int.class,int.class,long.class),tick=level.getMethod("tick");
        Field events=level.getField("events"),inputs=level.getField("inputs");Gson gson=new Gson();
        System.setOut(new PrintStream(new BufferedOutputStream(new FileOutputStream(FileDescriptor.out),65536),false));System.setErr(new PrintStream(new FileOutputStream(FileDescriptor.err),true));
        if(args.length>0&&args[0].equals("postprocess")) {
            var method=level.getMethod("postProcessStates",int.class);
            for(int neighbor:new int[]{0,1,80,86,90,5790,26683}) {
                Map<String,Object> row=new LinkedHashMap<>();row.put("neighbor",neighbor);row.put("states",method.invoke(world,neighbor));
                System.out.println("postprocess_states "+gson.toJson(row));
            }
            System.out.flush();return;
        }
        if(args.length>0&&args[0].equals("bubbles")) {
            var setup=level.getMethod("configureBubble",int.class,int.class,long.class);
            for(int id:(int[])level.getMethod("bubbleStates").invoke(null))for(int scene=0;scene<13;scene++)for(long seed:new long[]{0,1234,-1})for(String action:new String[]{"tick","place"}) {
                setup.invoke(world,id,scene,seed);Map<String,Object> row=new LinkedHashMap<>();row.put("state",id);row.put("scene",scene);row.put("seed",seed);row.put("action",action);row.put("inputs",inputs.get(world));
                try{row.put("next",Long.toUnsignedString((long)level.getMethod(action).invoke(world)));}catch(InvocationTargetException e){System.out.flush();throw new IllegalStateException(gson.toJson(row),e.getCause());}
                row.put("events",events.get(world));System.out.println("bubble_tick "+gson.toJson(row));
            }
            System.out.flush();return;
        }
        Map<String,String> handlers=new TreeMap<>();List<Integer> distances=new ArrayList<>(),leaves=new ArrayList<>();
        for(int id=0;id<26684;id++){String owner=(String)handler.invoke(null,id);handlers.put((String)name.invoke(null,id),owner);distances.add((int)distance.invoke(null,id));if(owner.equals("dki"))leaves.add(id);}
        System.out.println("block_tick_handlers "+gson.toJson(handlers));System.out.println("leaf_distances "+gson.toJson(distances));
        for(int id:leaves)for(int scene=0;scene<17;scene++)for(long seed:new long[]{0,1234,-1}) {
            configure.invoke(world,id,scene,seed);Map<String,Object> row=new LinkedHashMap<>();row.put("state",id);row.put("scene",scene);row.put("seed",seed);row.put("inputs",inputs.get(world));
            try{row.put("next",Long.toUnsignedString((long)tick.invoke(world)));}catch(InvocationTargetException e){System.out.flush();throw new IllegalStateException(gson.toJson(row),e.getCause());}
            row.put("events",events.get(world));System.out.println("leaf_tick "+gson.toJson(row));
        }
        System.out.flush();
    }
}
