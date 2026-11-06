package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.util.*;
import java.util.function.Supplier;
import java.util.stream.Stream;

/** Actual MonumentBuilding.postProcess on explicit synthetic terrain. The
 * original child postprocessors and elder-guardian creation run unchanged. */
public final class MonumentBlocksProbe {
    static Class<?> cls(String name)throws Exception{return Class.forName(name);}
    static void field(Object value,String owner,String name,Object replacement)throws Exception {
        var f=cls(owner).getDeclaredField(name);f.setAccessible(true);f.set(value,replacement);
    }
    static String digest(Collection<String> lines)throws Exception {
        List<String> sorted=new ArrayList<>(lines);Collections.sort(sorted);
        return MonumentLayoutProbe.hash(String.join("\n",sorted));
    }
    static Object level(Object registries)throws Exception {
        var f=cls("sun.misc.Unsafe").getDeclaredField("theUnsafe");f.setAccessible(true);Object unsafe=f.get(null);
        Object level=cls("sun.misc.Unsafe").getMethod("allocateInstance",Class.class).invoke(unsafe,cls("aqu"));
        Object server=cls("sun.misc.Unsafe").getMethod("allocateInstance",Class.class).invoke(unsafe,cls("apn"));
        Object worldData=Proxy.newProxyInstance(cls("erl").getClassLoader(),new Class<?>[]{cls("erl")},(p,m,v)->{
            if(m.getName().equals("K"))return cls("cpn").getField("g").get(null);
            throw new UnsupportedOperationException(m.toString());
        });
        field(server,"net.minecraft.server.MinecraftServer","j",worldData);
        field(level,"aqu","J",server);
        Object lookup=cls("jo$a").getMethod("b").invoke(registries),dimension=null;
        Object dimensions=cls("jn$a").getMethod("b",cls("akq")).invoke(lookup,cls("lu").getField("aK").get(null));
        for(Object holder:((Stream<?>)cls("jo").getMethod("b").invoke(dimensions)).toList()) {
            Object key=((Optional<?>)cls("jm").getMethod("e").invoke(holder)).orElseThrow();
            if(cls("akq").getMethod("a").invoke(key).toString().equals("minecraft:overworld"))dimension=holder;
        }
        field(level,"dcw","D",dimension);
        field(level,"dcw","I",registries);
        field(level,"dcw","E",(Supplier<Object>)()->{try{return cls("bnc").getField("a").get(null);}catch(Exception e){throw new IllegalStateException(e);}});
        Object data=Proxy.newProxyInstance(cls("erm").getClassLoader(),new Class<?>[]{cls("erm")},(p,m,v)->{
            if(m.getReturnType()==long.class)return 0L;
            if(m.getReturnType()==boolean.class)return false;
            if(m.getReturnType()==int.class)return 0;
            if(m.getReturnType()==cls("bqo"))return cls("bqo").getField("c").get(null);
            throw new UnsupportedOperationException(m.toString());
        });
        field(level,"dcw","A",data);
        return level;
    }
    public static void main(String[] args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);
        Object registries=IglooProbeSupport.registryAccess(),level=level(registries);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        System.setErr(new PrintStream(new FileOutputStream(FileDescriptor.err),true));
        Object difficulty=cls("bqp").getConstructor(cls("bqo"),long.class,long.class,float.class).newInstance(cls("bqo").getField("c").get(null),0L,0L,0.0f);
        Map<String,Object> states=new HashMap<>();
        Object registry=cls("lt").getField("e").get(null);
        for(String name:new String[]{"air","stone","water","ice","packed_ice","blue_ice","seagrass"}) {
            Object block=cls("jz").getMethod("a",cls("akr")).invoke(registry,cls("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name));
            states.put(name,cls("dfy").getMethod("o").invoke(block));
        }
        var getX=cls("jd").getMethod("u");var getY=cls("jd").getMethod("v");var getZ=cls("jd").getMethod("w");
        var stateID=cls("dfy").getMethod("i",cls("dtc"));var fluid=cls("dtb$a").getMethod("u");
        int[][] clips={{-40,-4,40,80},{-32,0,-17,15},{-16,16,-1,31},{16,48,31,63}};
        int cases=0;
        for(long seed:new long[]{0,1,1234})for(String direction:new String[]{"north","east","south","west"})for(int scenario=0;scenario<4;scenario++)for(int clipIndex=0;clipIndex<clips.length;clipIndex++) {
            final int mode=scenario,sea=mode==1?70:63;
            Object layoutRandom=cls("dyz").getConstructor(long.class).newInstance(seed);
            Object facing=cls("ji").getMethod("a",String.class).invoke(null,direction);
            Object building=cls("emb$h").getConstructor(cls("ayw"),int.class,int.class,cls("ji")).newInstance(layoutRandom,-29,7,facing);
            Object random=cls("dyz").getConstructor(long.class).newInstance(seed^0x123456789abcdefL);
            Object regionRandom=cls("dyz").getConstructor(long.class).newInstance(seed^0x7654321L);
            Map<String,Object> changed=new HashMap<>();Set<String> ticks=new TreeSet<>();List<String> actors=new ArrayList<>();
            java.util.function.Function<int[],Object> read=p->{
                String key=p[0]+" "+p[1]+" "+p[2];if(changed.containsKey(key))return changed.get(key);
                int floor=mode==1?50+Math.floorMod(p[0]*3+p[2],7):mode==3?80:30;
                if(p[1]<=floor)return states.get("stone");
                if(p[1]>=sea)return states.get("air");
                if(mode==2) {
                    int kind=Math.floorMod(p[0]+p[1]+p[2],11);
                    if(kind<3)return states.get(new String[]{"ice","packed_ice","blue_ice"}[kind]);
                    if(kind==3)return states.get("seagrass");
                }
                return states.get("water");
            };
            Object world=Proxy.newProxyInstance(cls("dds").getClassLoader(),new Class<?>[]{cls("dds")},(proxy,method,values)->{
                String name=method.getName();
                if(name.equals("I_"))return -64;if(name.equals("J_")||name.equals("am"))return 384;if(name.equals("an"))return 320;
                if(name.equals("z_"))return sea;if(name.equals("E"))return level;
                if(name.equals("E_"))return regionRandom;
                if(name.equals("d_"))return difficulty;
                if(name.equals("a_")&&values.length==1&&cls("bsr").isInstance(values[0])) {
                    Object p=cls("bsr").getMethod("do").invoke(values[0]);
                    actors.add(getX.invoke(p)+" "+getY.invoke(p)+" "+getZ.invoke(p));return null;
                }
                if(values!=null && values.length>0 && cls("jd").isInstance(values[0])) {
                    int x=(Integer)getX.invoke(values[0]),y=(Integer)getY.invoke(values[0]),z=(Integer)getZ.invoke(values[0]);
                    if(name.equals("a_"))return read.apply(new int[]{x,y,z});
                    if(name.equals("b_"))return fluid.invoke(read.apply(new int[]{x,y,z}));
                    if(name.equals("a")&&values.length==3&&cls("dtc").isInstance(values[1])) {changed.put(x+" "+y+" "+z,values[1]);return true;}
                    if(name.equals("a")&&values.length==3&&cls("epd").isInstance(values[1])) {ticks.add(x+" "+y+" "+z);return null;}
                }
                if(method.isDefault())return InvocationHandler.invokeDefault(proxy,method,values);
                throw new UnsupportedOperationException(method.toString());
            });
            int[] box=clips[clipIndex];
            Object clip=cls("ejj").getConstructor(int.class,int.class,int.class,int.class,int.class,int.class).newInstance(box[0],-64,box[1],box[2],319,box[3]);
            Object chunk=cls("dcd").getConstructor(int.class,int.class).newInstance(box[0]>>4,box[1]>>4);
            Object origin=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(0,0,0);
            cls("ejv").getMethod("a",cls("dds"),cls("ddq"),cls("duz"),cls("ayw"),cls("ejj"),cls("dcd"),cls("jd")).invoke(building,world,null,null,random,clip,chunk,origin);
            List<String> blocks=new ArrayList<>();for(var entry:changed.entrySet())blocks.add(entry.getKey()+" "+stateID.invoke(null,entry.getValue()));
            System.out.printf("monument_blocks %d %s %d %d %d %s %d %s %d %s %d%n",seed,direction,scenario,clipIndex,blocks.size(),digest(blocks),ticks.size(),digest(ticks),actors.size(),digest(actors),(Long)cls("ayw").getMethod("g").invoke(random));
            if(++cases%16==0)System.err.println("completed "+cases+" monument block cases");
        }
    }
}
