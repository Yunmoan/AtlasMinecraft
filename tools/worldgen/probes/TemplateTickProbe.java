package atlasprobe;

import com.google.gson.Gson;
import java.io.*;
import java.lang.reflect.*;
import java.util.*;
import java.util.stream.Stream;

public final class TemplateTickProbe {
    static Class<?> c(String name)throws Exception{return Class.forName(name);}
    static final class Loader extends ClassLoader {Class<?> define(byte[] data){return defineClass("TemplateTickLevel",data,0,data.length);}}
    public static void main(String[]args)throws Exception {
        Object version=c("t").getMethod("a").invoke(null);c("ab").getMethod("a",c("ae")).invoke(null,version);c("akt").getMethod("a").invoke(null);RuinedPortalProbeSupport.bindTags();
        Class<?> level;try(InputStream in=TemplateTickProbe.class.getResourceAsStream("/TemplateTickLevel.class")){level=new Loader().define(in.readAllBytes());}
        Field f=c("sun.misc.Unsafe").getDeclaredField("theUnsafe");f.setAccessible(true);Object unsafe=f.get(null),world=c("sun.misc.Unsafe").getMethod("allocateInstance",Class.class).invoke(unsafe,level);
        Object registries=IglooProbeSupport.registryAccess(),lookup=c("jo$a").getMethod("b").invoke(registries);
        Object dimensions=c("jn$a").getMethod("b",c("akq")).invoke(lookup,c("lu").getField("aK").get(null));
        for(Object h:((Stream<?>)c("jo").getMethod("b").invoke(dimensions)).toList()){Object key=((Optional<?>)c("jm").getMethod("e").invoke(h)).orElseThrow();if(c("akq").getMethod("a").invoke(key).toString().equals("minecraft:overworld"))level.getField("dimension").set(world,c("jm").getMethod("a").invoke(h));}
        Gson gson=new Gson();Method run=level.getMethod("run",int.class,int.class,int.class,boolean.class);
        System.setOut(new PrintStream(new BufferedOutputStream(new FileOutputStream(FileDescriptor.out),65536),false));
        for(int id:(int[])level.getMethod("states").invoke(null))for(int scene=0;scene<12;scene++)for(int shape=0;shape<3;shape++)for(boolean finish:new boolean[]{false,true}) {
            try {
                Map<String,Object> row=(Map<String,Object>)run.invoke(world,id,scene,shape,finish);
                row.put("packed",TreeFeatureProbe.packedTicks((List<Map<String,Object>>)row.get("requests")));
                System.out.println("template_ticks "+gson.toJson(row));
            }catch(InvocationTargetException e){System.out.flush();throw new IllegalStateException("state="+id+" scene="+scene+" shape="+shape+" finish="+finish,e.getCause());}
        }
        System.out.flush();
    }
}
