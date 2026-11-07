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

/** Complete original IglooPiece block/NBT/shape stage. Template actors are
 * disabled explicitly so the independent block stage needs no ServerLevel. */
public final class IglooBlocksProbe {
    static Class<?> cls(String n)throws Exception{return Class.forName(n);}
    static String digest(List<String> values)throws Exception {values.sort(String::compareTo);return HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(String.join("\n",values).getBytes(StandardCharsets.UTF_8)));}
    public static void main(String[] args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);
        Object registries=IglooProbeSupport.registryAccess(),manager=IglooProbeSupport.manager();
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Map<String,Object> terrain=new HashMap<>();Object blockRegistry=cls("lt").getField("e").get(null);
        for(String name:new String[]{"air","stone","grass_block","water","oak_leaves"})terrain.put(name,cls("dfy").getMethod("o").invoke(cls("jz").getMethod("a",cls("akr")).invoke(blockRegistry,cls("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name))));
        var getX=cls("jd").getMethod("u");var getY=cls("jd").getMethod("v");var getZ=cls("jd").getMethod("w");var fluid=cls("dtb$a").getMethod("u");var isAir=cls("dtb$a").getMethod("i");
        for(long seed:new long[]{0,1,-1,1234})for(int rotation=0;rotation<4;rotation++)for(int scenario=0;scenario<4;scenario++)for(String name:new String[]{"top","middle","bottom"})for(int[] anchor:new int[][]{{0,0},{15,15},{-16,-16}})for(int dx=0;dx<=1;dx++)for(int dz=0;dz<=1;dz++) {
            int chunkX=(anchor[0]>>4)+dx,chunkZ=(anchor[1]>>4)+dz;
            Object key=cls("akr").getMethod("a",String.class).invoke(null,"minecraft:igloo/"+name),base=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(anchor[0],90,anchor[1]);
            Object piece=cls("elq$a").getConstructor(cls("enu"),cls("akr"),cls("jd"),cls("dmm"),int.class).newInstance(manager,key,base,cls("dmm").getEnumConstants()[rotation],name.equals("bottom")?15:0);
            var placement=cls("ekb").getDeclaredField("c");placement.setAccessible(true);cls("enp").getMethod("a",boolean.class).invoke(placement.get(piece),true);
            Map<String,Object> changed=new HashMap<>(),entities=new HashMap<>();final int mode=scenario;
            java.util.function.Function<int[],Object> read=p->{String pos=p[0]+" "+p[1]+" "+p[2];if(changed.containsKey(pos))return changed.get(pos);int floor=mode==1?60+Math.floorMod(p[0]+2*p[2],5):mode==2||mode==3?60:63;if(p[1]<=floor)return terrain.get(p[1]==floor?"grass_block":"stone");if(mode==2&&p[1]<=66)return terrain.get("water");if(mode==3&&p[1]==72)return terrain.get("oak_leaves");return terrain.get("air");};
            Object world=Proxy.newProxyInstance(cls("dds").getClassLoader(),new Class<?>[]{cls("dds")},(proxy,method,values)->{
                String methodName=method.getName();if(methodName.equals("I_"))return -64;if(methodName.equals("an"))return 384;if(methodName.equals("am"))return 320;if(methodName.equals("H_"))return registries;if(methodName.equals("C"))return seed;
                if(values!=null&&values.length==3&&cls("dyy$a").isInstance(values[0])){int x=(Integer)values[1],z=(Integer)values[2];for(int y=319;y>=-64;y--)if(!(Boolean)isAir.invoke(read.apply(new int[]{x,y,z})))return y+1;return -64;}
                if(values!=null&&values.length>0&&cls("jd").isInstance(values[0])){
                    int x=(Integer)getX.invoke(values[0]),y=(Integer)getY.invoke(values[0]),z=(Integer)getZ.invoke(values[0]);String pos=x+" "+y+" "+z;
                    if(methodName.equals("a_"))return read.apply(new int[]{x,y,z});if(methodName.equals("b_"))return fluid.invoke(read.apply(new int[]{x,y,z}));if(methodName.equals("c_"))return entities.get(pos);
                    if(methodName.equals("a")&&values.length==3&&cls("dtc").isInstance(values[1])){
                        Object old=read.apply(new int[]{x,y,z}),oldBlock=cls("dtb$a").getMethod("b").invoke(old),block=cls("dtb$a").getMethod("b").invoke(values[1]);changed.put(pos,values[1]);
                        if(cls("diq").isInstance(block)){if(oldBlock!=block||!entities.containsKey(pos))entities.put(pos,cls("diq").getMethod("a",cls("jd"),cls("dtc")).invoke(block,values[0],values[1]));}else entities.remove(pos);return true;
                    }
                }
                if(method.getReturnType()==void.class)return null;
                if(method.isDefault())return java.lang.reflect.InvocationHandler.invokeDefault(proxy,method,values);
                throw new UnsupportedOperationException(method.toString());
            });
            Object clip=cls("ejj").getConstructor(int.class,int.class,int.class,int.class,int.class,int.class).newInstance(chunkX*16,-64,chunkZ*16,chunkX*16+15,319,chunkZ*16+15);
            Object random=cls("dzx").getConstructor(cls("ayw")).newInstance(cls("dzz").getConstructor(long.class).newInstance(seed));
            Object chunk=cls("dcd").getConstructor(int.class,int.class).newInstance(chunkX,chunkZ);
            cls("ejv").getMethod("a",cls("dds"),cls("ddq"),cls("duz"),cls("ayw"),cls("ejj"),cls("dcd"),cls("jd")).invoke(piece,world,null,null,random,clip,chunk,base);
            List<String> edits=new ArrayList<>(),entityRows=new ArrayList<>(),fullEntities=new ArrayList<>();for(var entry:changed.entrySet())edits.add(entry.getKey()+" "+cls("dfy").getMethod("i",cls("dtc")).invoke(null,entry.getValue()));
            for(var entry:entities.entrySet()){
                Object tag=cls("dqh").getMethod("b",cls("jo$a")).invoke(entry.getValue(),registries);
                if("1".equals(System.getenv("ATLAS_IGLOO_NBT_DEBUG"))||name.equals(System.getenv("ATLAS_IGLOO_NBT_DEBUG")))System.err.println(entry.getKey()+" "+tag);
                entityRows.add(entry.getKey()+" "+cls("ub").getMethod("l",String.class).invoke(tag,"id")+" "+cls("ub").getMethod("l",String.class).invoke(tag,"LootTable")+" "+cls("ub").getMethod("i",String.class).invoke(tag,"LootTableSeed"));
                fullEntities.add(entry.getKey()+" "+IglooProbeSupport.canonicalNBT(tag));
            }
            if("1".equals(System.getenv("ATLAS_IGLOO_NBT_DEBUG"))||name.equals(System.getenv("ATLAS_IGLOO_NBT_DEBUG")))throw new IllegalStateException("explicit NBT debug stop");
            Object saved=cls("ejv").getMethod("a",cls("ekh")).invoke(piece,(Object)null);int[] box=(int[])cls("ub").getMethod("n",String.class).invoke(saved,"BB");
            System.out.printf("igloo %d %d %d %s %d %d %d %d %d %s %d %s %d %d %s%n",seed,rotation,scenario,name,anchor[0],anchor[1],chunkX,chunkZ,edits.size(),digest(edits),entities.size(),digest(entityRows),box[1],cls("ayw").getMethod("g").invoke(random),digest(fullEntities));
        }
    }
}
