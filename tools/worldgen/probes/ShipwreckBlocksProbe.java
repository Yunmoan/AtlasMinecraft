package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.TreeSet;
import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Proxy;

/** Complete original ShipwreckPiece.postProcess: rectangle ground sampling,
 * beached depth, transforms, liquids, shape, clipped loot markers and NBT/RNG. */
public final class ShipwreckBlocksProbe {
    static Class<?> cls(String name)throws Exception{return Class.forName(name);}
    public static void main(String[] args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);RuinedPortalProbeSupport.bindTags();
        Object registries=IglooProbeSupport.registryAccess(),manager=IglooProbeSupport.manager(ShipwreckProbeSupport.names());System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        for(String name:ShipwreckProbeSupport.names())for(int rotation=0;rotation<4;rotation++)for(boolean beached:new boolean[]{false,true})for(int mode=0;mode<6;mode++)for(int[] anchor:new int[][]{{0,90,0},{-16,90,-16},{15,90,15}})for(int outside=0;outside<3;outside++) {
            long seed=new long[]{0,1,-1,1234}[(rotation+mode+anchor[0])&3];Object origin=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(anchor[0],anchor[1],anchor[2]),key=cls("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name),piece=cls("emh$a").getConstructor(cls("enu"),cls("akr"),cls("jd"),cls("dmm"),boolean.class).newInstance(manager,key,origin,cls("dmm").getEnumConstants()[rotation],beached);
            int cx=(anchor[0]>>4)+outside,cz=anchor[2]>>4;Object clip=cls("ejj").getConstructor(int.class,int.class,int.class,int.class,int.class,int.class).newInstance(cx*16,-64,cz*16,cx*16+15,319,cz*16+15),chunk=cls("dcd").getConstructor(int.class,int.class).newInstance(cx,cz);
            Map<String,Object> changed=new HashMap<>(),entities=new HashMap<>();java.util.Set<String> ticks=new TreeSet<>();Object delegate=RuinedPortalProbeSupport.world(mode,changed,entities,registries);
            Object world=Proxy.newProxyInstance(cls("dds").getClassLoader(),new Class<?>[]{cls("dds")},(proxy,method,arguments)->{
                if(arguments!=null&&arguments.length>=3&&cls("jd").isInstance(arguments[0])&&cls("epd").isInstance(arguments[1]))ticks.add(cls("jd").getMethod("u").invoke(arguments[0])+" "+cls("jd").getMethod("v").invoke(arguments[0])+" "+cls("jd").getMethod("w").invoke(arguments[0]));
                try{return method.invoke(delegate,arguments);}catch(InvocationTargetException error){throw error.getCause();}
            }),random=cls("dzx").getConstructor(cls("ayw")).newInstance(cls("dzz").getConstructor(long.class).newInstance(seed));
            var postprocess=cls("ejv").getMethod("a",cls("dds"),cls("ddq"),cls("duz"),cls("ayw"),cls("ejj"),cls("dcd"),cls("jd"));
            if(outside<2)postprocess.invoke(piece,world,null,null,random,clip,chunk,origin);
            else {
                Object box=cls("ejv").getMethod("f").invoke(piece);int minX=(Integer)cls("ejj").getMethod("h").invoke(box)>>4,maxX=(Integer)cls("ejj").getMethod("k").invoke(box)>>4,minZ=(Integer)cls("ejj").getMethod("j").invoke(box)>>4,maxZ=(Integer)cls("ejj").getMethod("m").invoke(box)>>4;
                for(int x=minX;x<=maxX;x++)for(int z=minZ;z<=maxZ;z++)postprocess.invoke(piece,world,null,null,random,cls("ejj").getConstructor(int.class,int.class,int.class,int.class,int.class,int.class).newInstance(x*16,-64,z*16,x*16+15,319,z*16+15),cls("dcd").getConstructor(int.class,int.class).newInstance(x,z),origin);
            }
            List<String> edits=new ArrayList<>(),fullEntities=new ArrayList<>();for(var entry:changed.entrySet())edits.add(entry.getKey()+" "+cls("dfy").getMethod("i",cls("dtc")).invoke(null,entry.getValue()));for(var entry:entities.entrySet()){Object tag=cls("dqh").getMethod("b",cls("jo$a")).invoke(entry.getValue(),registries);fullEntities.add(entry.getKey()+" "+IglooProbeSupport.canonicalNBT(tag));}
            Object saved=cls("ejv").getMethod("a",cls("ekh")).invoke(piece,(Object)null);
            String debugKey=name+" "+seed+" "+rotation+" "+beached+" "+mode+" "+anchor[0]+" "+anchor[2]+" "+outside;if("1".equals(System.getenv("ATLAS_SHIPWRECK_DEBUG"))||debugKey.equals(System.getenv("ATLAS_SHIPWRECK_DEBUG"))){edits.sort(String::compareTo);java.nio.file.Files.writeString(java.nio.file.Path.of(System.getenv("ATLAS_SHIPWRECK_DEBUG_OUTPUT")),String.join("\n",edits));System.err.println(saved);throw new IllegalStateException("explicit shipwreck debug stop");}
            System.out.printf("%s %s %d %d %d %d %d %d %d %d %d %s %d %s %d %s %d %s%n",outside==2?"shipsequence":"ship",name,seed,rotation,beached?1:0,mode,anchor[0],anchor[2],cx,cz,edits.size(),RuinedPortalBlocksProbe.digest(edits),entities.size(),RuinedPortalBlocksProbe.digest(fullEntities),cls("ayw").getMethod("g").invoke(random),RuinedPortalBlocksProbe.digest(new ArrayList<>(List.of(IglooProbeSupport.canonicalNBT(saved)))),ticks.size(),RuinedPortalBlocksProbe.digest(new ArrayList<>(ticks)));
        }
    }
}
