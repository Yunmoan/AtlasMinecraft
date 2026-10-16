package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.lang.reflect.Proxy;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.HexFormat;
import java.util.List;
import java.util.Map;

/** Original NetherFossilPiece across every template/rotation, expanded clip,
 * ignored air, surrounding terrain and caller RNG continuation. */
public final class NetherFossilBlocksProbe {
    static Class<?> cls(String n)throws Exception{return Class.forName(n);}
    static String digest(List<String> rows)throws Exception {
        rows.sort(String::compareTo);
        return HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(String.join("\n",rows).getBytes(StandardCharsets.UTF_8)));
    }
    public static void main(String[] args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);
        String[] names=new String[14];for(int i=0;i<14;i++)names[i]="nether_fossils/fossil_"+(i+1);
        Object manager=IglooProbeSupport.manager(names),registries=IglooProbeSupport.registryAccess();
        Object dimensionRegistry=cls("ka").getMethod("d",cls("akq")).invoke(registries,cls("lu").getField("aK").get(null));
        Object dimensionType=cls("jz").getMethod("f",cls("akq")).invoke(dimensionRegistry,cls("dww").getField("b").get(null));
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Map<String,Object> states=new HashMap<>();Object registry=cls("lt").getField("e").get(null);
        for(String name:new String[]{"air","netherrack","water","lava"})states.put(name,cls("dfy").getMethod("o").invoke(cls("jz").getMethod("a",cls("akr")).invoke(registry,cls("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name))));
        var getX=cls("jd").getMethod("u");var getY=cls("jd").getMethod("v");var getZ=cls("jd").getMethod("w");var fluid=cls("dtb$a").getMethod("u");
        for(long seed:new long[]{0,1,-1})for(int i=1;i<=14;i++)for(int rotation=0;rotation<4;rotation++)for(int scenario=0;scenario<3;scenario++)for(int[] anchor:new int[][]{{0,0},{15,15},{-16,-16}})for(int offset=0;offset<2;offset++) {
            int chunkX=(anchor[0]>>4)+offset,chunkZ=anchor[1]>>4;
            Object origin=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(anchor[0],60,anchor[1]);
            Object key=cls("akr").getMethod("a",String.class).invoke(null,"minecraft:nether_fossils/fossil_"+i);
            Object piece=cls("elz$a").getConstructor(cls("enu"),cls("akr"),cls("jd"),cls("dmm")).newInstance(manager,key,origin,cls("dmm").getEnumConstants()[rotation]);
            Map<String,Object> changes=new HashMap<>();final int mode=scenario;
            java.util.function.Function<int[],Object> read=p->{String pos=p[0]+" "+p[1]+" "+p[2];if(changes.containsKey(pos))return changes.get(pos);if(p[1]<=59)return states.get("netherrack");if(p[1]<=65)return states.get(mode==1?"water":mode==2?"lava":"air");return states.get("air");};
            Object world=Proxy.newProxyInstance(cls("dds").getClassLoader(),new Class<?>[]{cls("dds")},(proxy,method,values)->{
                String n=method.getName();if(n.equals("I_"))return 0;if(n.equals("an"))return 256;if(n.equals("am"))return 256;if(n.equals("H_"))return registries;if(n.equals("D_"))return dimensionType;
                if(values!=null&&values.length>0&&cls("jd").isInstance(values[0])) {
                    int x=(Integer)getX.invoke(values[0]),y=(Integer)getY.invoke(values[0]),z=(Integer)getZ.invoke(values[0]);String pos=x+" "+y+" "+z;
                    if(n.equals("a_"))return read.apply(new int[]{x,y,z});if(n.equals("b_"))return fluid.invoke(read.apply(new int[]{x,y,z}));if(n.equals("c_"))return null;
                    if(n.equals("a")&&values.length==3&&cls("dtc").isInstance(values[1])) {changes.put(pos,values[1]);return true;}
                }
                if(method.getReturnType()==void.class)return null;
                if(method.isDefault())return java.lang.reflect.InvocationHandler.invokeDefault(proxy,method,values);
                throw new UnsupportedOperationException(method.toString());
            });
            Object random=cls("dzx").getConstructor(cls("ayw")).newInstance(cls("dzz").getConstructor(long.class).newInstance(seed));
            Object clip=cls("ejj").getConstructor(int.class,int.class,int.class,int.class,int.class,int.class).newInstance(chunkX*16,0,chunkZ*16,chunkX*16+15,255,chunkZ*16+15);
            cls("ejv").getMethod("a",cls("dds"),cls("ddq"),cls("duz"),cls("ayw"),cls("ejj"),cls("dcd"),cls("jd")).invoke(piece,world,null,null,random,clip,cls("dcd").getConstructor(int.class,int.class).newInstance(chunkX,chunkZ),origin);
            List<String> rows=new ArrayList<>();for(var e:changes.entrySet())rows.add(e.getKey()+" "+cls("dfy").getMethod("i",cls("dtc")).invoke(null,e.getValue()));
            Object saved=cls("ejv").getMethod("a",cls("ekh")).invoke(piece,(Object)null);
            System.out.printf("nether_fossil %d %d %d %d %d %d %d %d %d %s %s %d%n",seed,i,rotation,scenario,anchor[0],anchor[1],chunkX,chunkZ,changes.size(),digest(rows),digest(new ArrayList<>(List.of(IglooProbeSupport.canonicalNBT(saved)))),cls("ayw").getMethod("g").invoke(random));
        }
    }
}
