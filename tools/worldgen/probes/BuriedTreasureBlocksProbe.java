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

/** Calls the original BuriedTreasurePiece, including un-clipped neighbor
 * repair, default-liquid identity tests, saved bounding box and loot RNG. */
public final class BuriedTreasureBlocksProbe {
    static Class<?> cls(String n)throws Exception{return Class.forName(n);}
    static String digest(List<String> rows)throws Exception {
        rows.sort(String::compareTo);
        return HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(String.join("\n",rows).getBytes(StandardCharsets.UTF_8)));
    }
    public static void main(String[] args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);
        cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);
        Object registries=IglooProbeSupport.registryAccess();
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Map<String,Object> states=new HashMap<>();Object blockRegistry=cls("lt").getField("e").get(null);
        for(String name:new String[]{"air","stone","sandstone","andesite","granite","diorite","deepslate","sand","water","lava","chest","dirt","bedrock"})
            states.put(name,cls("dfy").getMethod("o").invoke(cls("jz").getMethod("a",cls("akr")).invoke(blockRegistry,cls("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name))));
        Object level=cls("dts").getField("aP").get(null);
        // Flowing fluid must not be treated as the default source state.
        for(String name:new String[]{"water","lava"})states.put(name+"_flowing",cls("dte").getMethod("a",cls("duf"),Comparable.class).invoke(states.get(name),level,1));
        var getX=cls("jd").getMethod("u");var getY=cls("jd").getMethod("v");var getZ=cls("jd").getMethod("w");
        var stateID=cls("dfy").getMethod("i",cls("dtc"));var fluid=cls("dtb$a").getMethod("u");
        for(long seed:new long[]{0,1,-1,1234})for(int scenario=0;scenario<24;scenario++)for(int[] anchor:new int[][]{{9,9},{15,15},{-16,-16}})for(int clipOffset=0;clipOffset<2;clipOffset++) {
            int chunkX=(anchor[0]>>4)+clipOffset,chunkZ=anchor[1]>>4,minY=scenario==23?0:-64;
            Object base=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(anchor[0],90,anchor[1]);
            Object piece=cls("elk$a").getConstructor(cls("jd")).newInstance(base);
            Map<String,Object> changed=new HashMap<>(),entities=new HashMap<>();final int mode=scenario;
            java.util.function.Function<int[],Object> read=p->{
                String key=p[0]+" "+p[1]+" "+p[2];if(changed.containsKey(key))return changed.get(key);
                if(mode==22)return states.get("air");
                int floor=mode==23?0:60;
                if(p[1]<floor)return states.get("bedrock");
                if(p[1]==floor)return states.get(new String[]{"stone","sandstone","andesite","granite","diorite","deepslate"}[mode%6]);
                if(mode==21&&p[0]==anchor[0]&&p[2]==anchor[1]&&p[1]==floor+1)return states.get("chest");
                if(mode>=12&&mode<18&&p[1]==floor+1&&p[0]==anchor[0]+1)return states.get("dirt");
                if(mode>=18&&mode<21&&p[1]==floor+1&&p[0]==anchor[0]&&p[2]==anchor[1])return states.get("sand");
                if(p[1]<=floor+3)return states.get(new String[]{"air","water","lava","water_flowing","lava_flowing","air"}[(mode/6)%4]);
                return states.get("air");
            };
            Object world=Proxy.newProxyInstance(cls("dds").getClassLoader(),new Class<?>[]{cls("dds")},(proxy,method,values)->{
                String n=method.getName();if(n.equals("I_"))return minY;if(n.equals("an"))return 320-minY;if(n.equals("am"))return 320;if(n.equals("H_"))return registries;
                if(values!=null&&values.length==3&&cls("dyy$a").isInstance(values[0]))return mode==22?minY:mode==23?1:mode>=18&&mode<=21?62:61;
                if(values!=null&&values.length>0&&cls("jd").isInstance(values[0])) {
                    int x=(Integer)getX.invoke(values[0]),y=(Integer)getY.invoke(values[0]),z=(Integer)getZ.invoke(values[0]);String key=x+" "+y+" "+z;
                    if(n.equals("a_"))return read.apply(new int[]{x,y,z});if(n.equals("b_"))return fluid.invoke(read.apply(new int[]{x,y,z}));if(n.equals("c_"))return entities.get(key);
                    if(n.equals("a")&&values.length==3&&cls("dtc").isInstance(values[1])) {
                        changed.put(key,values[1]);Object block=cls("dtb$a").getMethod("b").invoke(values[1]);
                        if(cls("diq").isInstance(block))entities.put(key,cls("diq").getMethod("a",cls("jd"),cls("dtc")).invoke(block,values[0],values[1]));else entities.remove(key);return true;
                    }
                }
                if(method.isDefault())return java.lang.reflect.InvocationHandler.invokeDefault(proxy,method,values);
                throw new UnsupportedOperationException(method.toString());
            });
            Object random=cls("dzx").getConstructor(cls("ayw")).newInstance(cls("dzz").getConstructor(long.class).newInstance(seed));
            Object clip=cls("ejj").getConstructor(int.class,int.class,int.class,int.class,int.class,int.class).newInstance(chunkX*16,minY,chunkZ*16,chunkX*16+15,319,chunkZ*16+15);
            Object chunk=cls("dcd").getConstructor(int.class,int.class).newInstance(chunkX,chunkZ);
            cls("ejv").getMethod("a",cls("dds"),cls("ddq"),cls("duz"),cls("ayw"),cls("ejj"),cls("dcd"),cls("jd")).invoke(piece,world,null,null,random,clip,chunk,base);
            List<String> edits=new ArrayList<>(),savedEntities=new ArrayList<>();for(var e:changed.entrySet())edits.add(e.getKey()+" "+stateID.invoke(null,e.getValue()));
            for(var e:entities.entrySet())savedEntities.add(e.getKey()+" "+IglooProbeSupport.canonicalNBT(cls("dqh").getMethod("b",cls("jo$a")).invoke(e.getValue(),registries)));
            Object saved=cls("ejv").getMethod("a",cls("ekh")).invoke(piece,(Object)null);
            System.out.printf("treasure %d %d %d %d %d %d %d %s %d %s %s %d%n",seed,scenario,anchor[0],anchor[1],chunkX,chunkZ,edits.size(),digest(edits),entities.size(),digest(savedEntities),digest(new ArrayList<>(List.of(IglooProbeSupport.canonicalNBT(saved)))),cls("ayw").getMethod("g").invoke(random));
        }
    }
}
