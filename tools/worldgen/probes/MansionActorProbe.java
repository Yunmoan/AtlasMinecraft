package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.util.*;

/** Actual mansion marker actors and STRUCTURE finalization in a fresh NORMAL world. */
public final class MansionActorProbe {
    static Class<?> cls(String name)throws Exception{return Class.forName(name);}
    @SuppressWarnings("unchecked") static String saved(Object entity)throws Exception {
        Object tag=cls("ub").getConstructor().newInstance();
        cls("bsr").getMethod("e",cls("ub")).invoke(entity,tag);
        cls("ub").getMethod("r",String.class).invoke(tag,"UUID");
        List<Object> attributes=(List<Object>)cls("ub").getMethod("c",String.class,int.class).invoke(tag,"attributes",10);
        attributes.sort(Comparator.comparing(a->{try{return (String)cls("ub").getMethod("l",String.class).invoke(a,"id");}catch(Exception e){throw new IllegalStateException(e);}}));
        ByteArrayOutputStream bytes=new ByteArrayOutputStream();
        cls("uo").getMethod("a",cls("uy"),DataOutput.class).invoke(null,tag,new DataOutputStream(bytes));
        return HexFormat.of().formatHex(bytes.toByteArray());
    }
    static Object actor(Object level, String id)throws Exception {
        Object registry=cls("lt").getField("f").get(null);
        Object type=cls("jz").getMethod("a",cls("akr")).invoke(registry,cls("akr").getMethod("a",String.class).invoke(null,id));
        Object entity=cls("bsx").getMethod("a",cls("dcw")).invoke(type,level);
        cls("bsr").getMethod("b",double.class,double.class,double.class,float.class,float.class).invoke(entity,0.5,42.0,0.5,0.0f,0.0f);
        cls("btp").getMethod("fR").invoke(entity);
        return entity;
    }
    public static void main(String[]args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);
        Object registries=IglooProbeSupport.registryAccess(),level=MonumentBlocksProbe.level(registries);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        System.setErr(new PrintStream(new FileOutputStream(FileDescriptor.err),true));
        Object difficulty=cls("bqp").getConstructor(cls("bqo"),long.class,long.class,float.class).newInstance(cls("bqo").getField("c").get(null),0L,0L,0f);
        String[] ids={"minecraft:evoker","minecraft:vindicator","minecraft:allay"};
        for(String id:ids) System.out.println("mansion_actor_default "+id+" "+saved(actor(level,id)));
        for(String id:ids) for(long seed:new long[]{0,1,-1,1234,Long.MIN_VALUE,Long.MAX_VALUE}) {
            Object random=cls("dzz").getConstructor(long.class).newInstance(seed);
            Object world=Proxy.newProxyInstance(cls("ddl").getClassLoader(),new Class<?>[]{cls("ddl")},(p,m,v)->{
                if(m.getName().equals("E_"))return random;if(m.getName().equals("E"))return level;
                if(m.getName().equals("d_"))return difficulty;
                if(m.isDefault())return InvocationHandler.invokeDefault(p,m,v);
                throw new UnsupportedOperationException(m.toString());
            });
            for(int index=0;index<3;index++) {
                Object entity=actor(level,id);
                cls("btp").getMethod("a",cls("ddl"),cls("bqp"),cls("btr"),cls("buh")).invoke(entity,world,difficulty,cls("btr").getField("d").get(null),null);
                // Read-only nextLong oracle: copy Xoroshiro state, so measuring
                // continuation does not advance the shared three-actor stream.
                Object engine=MonumentLayoutProbe.field(random,"dzz","e");
                long lo=(Long)MonumentLayoutProbe.field(engine,"dzy","b"),hi=(Long)MonumentLayoutProbe.field(engine,"dzy","c");
                Object copy=cls("dzz").getConstructor(long.class,long.class).newInstance(lo,hi);
                long next=(Long)cls("ayw").getMethod("g").invoke(copy);
                System.out.printf("mansion_actor %s %d %d %s %d%n",id,seed,index,saved(entity),next);
            }
        }
    }
}
