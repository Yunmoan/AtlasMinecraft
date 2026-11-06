package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.util.*;
import io.netty.buffer.*;

/** Actual 1.21.1 spawn and EndCrystal metadata packet codecs. */
public final class GeneratedEntityPacketProbe {
    static Class<?> cls(String name)throws Exception{return Class.forName(name);}
    static String encode(String packetClass,Object packet,Object access)throws Exception {
        ByteBuf bytes=Unpooled.buffer();
        Object buffer=cls("wk").getConstructor(ByteBuf.class,cls("ka")).newInstance(bytes,access);
        cls("yx").getMethod("encode",Object.class,Object.class).invoke(cls(packetClass).getField("a").get(null),buffer,packet);
        byte[] result=new byte[bytes.readableBytes()];bytes.readBytes(result);bytes.release();
        return HexFormat.of().formatHex(result);
    }
    public static void main(String[] args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);
        cls("ab").getMethod("a",cls("ae")).invoke(null,version);
        cls("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Object types=cls("lt").getField("f").get(null);
        Object access=cls("ka$c").getConstructor(List.class).newInstance(List.of(types));
        UUID uuid=UUID.fromString("00112233-4455-6677-8899-aabbccddeeff");
        Object zeroVelocity=cls("exc").getConstructor(double.class,double.class,double.class).newInstance(0d,0d,0d);
        var constructor=cls("abv").getConstructor(int.class,UUID.class,double.class,double.class,double.class,float.class,float.class,cls("bsx"),int.class,cls("exc"),double.class);
        for(String name:new String[]{"end_crystal","witch","cat","villager","zombie_villager"}) {
            Object resource=cls("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name);
            Object type=cls("jz").getMethod("a",cls("akr")).invoke(types,resource);
            System.out.printf("generated_type minecraft:%s %s%n",name,cls("jz").getMethod("a",Object.class).invoke(types,type));
            for(float yaw:new float[]{0,90,-90,359.99f,-359.99f,1.4f,-1.4f,180.5f})for(float pitch:new float[]{0,25,-25,1.4f,-1.4f}) {
                Object packet=constructor.newInstance(77,uuid,-17.5d,68d,32.5d,pitch,yaw,type,0,zeroVelocity,(double)yaw);
                System.out.printf("generated_spawn minecraft:%s %s %s %s%n",name,yaw,pitch,encode("abv",packet,access));
            }
        }
        Field unsafeField=sun.misc.Unsafe.class.getDeclaredField("theUnsafe");unsafeField.setAccessible(true);
        var unsafe=(sun.misc.Unsafe)unsafeField.get(null);
        Object server=unsafe.allocateInstance(cls("apn")),level=unsafe.allocateInstance(cls("aqu"));
        Object worldData=Proxy.newProxyInstance(cls("erl").getClassLoader(),new Class<?>[]{cls("erl")},(proxy,method,values)->{
            if(method.getName().equals("K"))return cls("cpn").getField("g").get(null);
            throw new UnsupportedOperationException(method.toString());
        });
        Field serverData=cls("net.minecraft.server.MinecraftServer").getDeclaredField("j");serverData.setAccessible(true);serverData.set(server,worldData);
        Field serverField=cls("aqu").getDeclaredField("J");serverField.setAccessible(true);serverField.set(level,server);
        Object crystalType=cls("jz").getMethod("a",cls("akr")).invoke(types,cls("akr").getMethod("a",String.class).invoke(null,"minecraft:end_crystal"));
        for(boolean bottom:new boolean[]{true,false})for(int beam=0;beam<3;beam++) {
            Object crystal=cls("chz").getConstructor(cls("bsx"),cls("dcw")).newInstance(crystalType,level);
            cls("chz").getMethod("a",boolean.class).invoke(crystal,bottom);
            if(beam!=0)cls("chz").getMethod("a",cls("jd")).invoke(crystal,cls("jd").getConstructor(int.class,int.class,int.class).newInstance(beam==1?-100:30000000,beam==1?50:-64,beam==1?123:-30000000));
            Object data=cls("bsr").getMethod("ar").invoke(crystal);
            List<?> values=(List<?>)cls("aka").getMethod("c").invoke(data);
            if(values==null)values=List.of();
            Object packet=cls("aez").getConstructor(int.class,List.class).newInstance(77,values);
            System.out.printf("generated_crystal_metadata %s %d %s%n",bottom,beam,encode("aez",packet,access));
        }
    }
}
