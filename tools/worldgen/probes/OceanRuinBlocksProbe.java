package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Proxy;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.TreeSet;
import java.util.function.Supplier;

/** Original complete OceanRuinPiece.postProcess with original capped
 * archaeology and successful drowned markers. Actor-finalize NBT is measured
 * separately; this oracle compares resulting blocks and persistent requests. */
public final class OceanRuinBlocksProbe {
    static Class<?> cls(String name)throws Exception{return Class.forName(name);}
    static Object level(long seed,Object registries)throws Exception {
        Object level=OceanRuinProcessorsProbe.seedLevel(seed),registry=cls("ka").getMethod("d",cls("akq")).invoke(registries,cls("lu").getField("aK").get(null)),dimension=((Optional<?>)cls("jz").getMethod("c",cls("akr")).invoke(registry,cls("akr").getMethod("a",String.class).invoke(null,"minecraft:overworld"))).orElseThrow();
        OceanRuinProcessorsProbe.field(level,"dcw","D",dimension);OceanRuinProcessorsProbe.field(level,"dcw","I",registries);OceanRuinProcessorsProbe.field(level,"dcw","E",(Supplier<Object>)()->{try{return cls("bnc").getField("a").get(null);}catch(Exception error){throw new IllegalStateException(error);}});
        Object data=Proxy.newProxyInstance(cls("erm").getClassLoader(),new Class<?>[]{cls("erm")},(proxy,method,args)->{if(method.getReturnType()==long.class)return 0L;if(method.getReturnType()==boolean.class)return false;if(method.getReturnType()==int.class)return 0;throw new UnsupportedOperationException(method.toString());});OceanRuinProcessorsProbe.field(level,"dcw","A",data);return level;
    }
    public static void main(String[] args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);RuinedPortalProbeSupport.bindTags();Object registries=IglooProbeSupport.registryAccess(),manager=IglooProbeSupport.manager(OceanRuinProbeSupport.names());System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Object difficulty=cls("bqp").getConstructor(cls("bqo"),long.class,long.class,float.class).newInstance(cls("bqo").getField("c").get(null),0L,0L,0f);
        for(String name:OceanRuinProbeSupport.names())for(int rotation=0;rotation<4;rotation++)for(int mode=0;mode<6;mode++)for(int[] anchor:new int[][]{{0,90,0},{-16,90,-16},{15,90,15}})for(int outside=0;outside<3;outside++) {
            long seed=new long[]{0,1,-1,1234}[(rotation+mode+anchor[0])&3];boolean cold=!name.contains("warm"),large=name.contains("big_");float integrity=name.contains("cracked")?0.7f:name.contains("mossy")?0.5f:large?0.9f:0.8f;
            Object origin=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(anchor[0],anchor[1],anchor[2]),key=cls("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name),piece=cls("emd$a").getConstructor(cls("enu"),cls("akr"),cls("jd"),cls("dmm"),float.class,cls("eme$a"),boolean.class).newInstance(manager,key,origin,cls("dmm").getEnumConstants()[rotation],integrity,cls("eme$a").getEnumConstants()[cold?1:0],large);
            int cx=(anchor[0]>>4)+outside,cz=anchor[2]>>4;Object clip=cls("ejj").getConstructor(int.class,int.class,int.class,int.class,int.class,int.class).newInstance(cx*16,-64,cz*16,cx*16+15,319,cz*16+15),chunk=cls("dcd").getConstructor(int.class,int.class).newInstance(cx,cz);
            Map<String,Object> changed=new HashMap<>(),entities=new HashMap<>();TreeSet<String> ticks=new TreeSet<>();List<String> actors=new ArrayList<>();Object delegate=RuinedPortalProbeSupport.world(mode,changed,entities,registries),level=level(seed,registries),worldRandom=cls("dzz").getConstructor(long.class).newInstance(seed^0x123456789abcdefL);
            Object world=Proxy.newProxyInstance(cls("dds").getClassLoader(),new Class<?>[]{cls("dds")},(proxy,method,values)->{
                String methodName=method.getName();if(methodName.equals("E")&&method.getParameterCount()==0)return level;if(methodName.equals("E_"))return worldRandom;if(methodName.equals("d_"))return difficulty;if(methodName.equals("z_"))return 63;
                if(values!=null&&values.length>=3&&cls("jd").isInstance(values[0])&&cls("epd").isInstance(values[1]))ticks.add(cls("jd").getMethod("u").invoke(values[0])+" "+cls("jd").getMethod("v").invoke(values[0])+" "+cls("jd").getMethod("w").invoke(values[0]));
                if(methodName.equals("a_")&&values!=null&&values.length==1&&cls("bsr").isInstance(values[0])){actors.add("minecraft:drowned "+(int)Math.floor((Double)cls("bsr").getMethod("dt").invoke(values[0]))+" "+(int)Math.floor((Double)cls("bsr").getMethod("dv").invoke(values[0]))+" "+(int)Math.floor((Double)cls("bsr").getMethod("dz").invoke(values[0]))+" true");return null;}
                try{return method.invoke(delegate,values);}catch(InvocationTargetException error){throw error.getCause();}
            }),random=cls("dzx").getConstructor(cls("ayw")).newInstance(cls("dzz").getConstructor(long.class).newInstance(seed));
            var post=cls("ejv").getMethod("a",cls("dds"),cls("ddq"),cls("duz"),cls("ayw"),cls("ejj"),cls("dcd"),cls("jd"));
            if(outside<2)post.invoke(piece,world,null,null,random,clip,chunk,origin);else {Object box=cls("ejv").getMethod("f").invoke(piece);int minX=(Integer)cls("ejj").getMethod("h").invoke(box)>>4,maxX=(Integer)cls("ejj").getMethod("k").invoke(box)>>4,minZ=(Integer)cls("ejj").getMethod("j").invoke(box)>>4,maxZ=(Integer)cls("ejj").getMethod("m").invoke(box)>>4;for(int x=minX;x<=maxX;x++)for(int z=minZ;z<=maxZ;z++)post.invoke(piece,world,null,null,random,cls("ejj").getConstructor(int.class,int.class,int.class,int.class,int.class,int.class).newInstance(x*16,-64,z*16,x*16+15,319,z*16+15),cls("dcd").getConstructor(int.class,int.class).newInstance(x,z),origin);}
            List<String> edits=new ArrayList<>(),fullEntities=new ArrayList<>();for(var entry:changed.entrySet())edits.add(entry.getKey()+" "+cls("dfy").getMethod("i",cls("dtc")).invoke(null,entry.getValue()));for(var entry:entities.entrySet())fullEntities.add(entry.getKey()+" "+IglooProbeSupport.canonicalNBT(cls("dqh").getMethod("b",cls("jo$a")).invoke(entry.getValue(),registries)));Object saved=cls("ejv").getMethod("a",cls("ekh")).invoke(piece,(Object)null);
            String debugKey=name+" "+seed+" "+rotation+" "+mode+" "+anchor[0]+" "+anchor[2]+" "+outside;if("1".equals(System.getenv("ATLAS_OCEAN_RUIN_DEBUG"))||debugKey.equals(System.getenv("ATLAS_OCEAN_RUIN_DEBUG"))){edits.sort(String::compareTo);java.nio.file.Files.writeString(java.nio.file.Path.of(System.getenv("ATLAS_OCEAN_RUIN_DEBUG_OUTPUT")),String.join("\n",edits));System.err.println(saved);throw new IllegalStateException("explicit ocean ruin debug stop");}
            System.out.printf("%s %s %d %d %d %d %d %d %d %d %s %d %s %d %s %d %s %d %s%n",outside==2?"oceansequence":"ocean",name,seed,rotation,mode,anchor[0],anchor[2],cx,cz,edits.size(),RuinedPortalBlocksProbe.digest(edits),entities.size(),RuinedPortalBlocksProbe.digest(fullEntities),cls("ayw").getMethod("g").invoke(random),RuinedPortalBlocksProbe.digest(new ArrayList<>(List.of(IglooProbeSupport.canonicalNBT(saved)))),ticks.size(),RuinedPortalBlocksProbe.digest(new ArrayList<>(ticks)),actors.size(),RuinedPortalBlocksProbe.digest(actors));
        }
    }
}
