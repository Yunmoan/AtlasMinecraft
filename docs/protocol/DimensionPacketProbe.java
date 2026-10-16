package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.nio.file.*;
import java.util.*;
import io.netty.buffer.*;
import com.mojang.serialization.Codec;
import com.mojang.serialization.DynamicOps;

/** Uses the actual Atlas configuration registry order with the 1.21.1 codec. */
public final class DimensionPacketProbe {
    public static void main(String[] args)throws Exception {
        Object version=Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a",Class.forName("ae")).invoke(null,version);
        Class.forName("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        System.setErr(new PrintStream(new FileOutputStream(FileDescriptor.err),true));
        Class<?> key=Class.forName("akq"),id=Class.forName("akr");
        Object dimKey=Class.forName("lu").getField("aK").get(null);
        Object dimensions=Class.forName("ju").getConstructor(key,com.mojang.serialization.Lifecycle.class).newInstance(dimKey,com.mojang.serialization.Lifecycle.stable());
        Method decode=Class.forName("yx").getMethod("decode",Object.class),encode=Class.forName("yx").getMethod("encode",Object.class,Object.class);
        byte[] bytes=Files.readAllBytes(Path.of("internal/registry/registries.bin"));
        ByteBuf raw=Unpooled.wrappedBuffer(bytes);Object container=Class.forName("vw").getConstructor(ByteBuf.class).newInstance(raw);
        int packets=(int)Class.forName("vw").getMethod("l").invoke(container);
        for(int i=0;i<packets;i++) {
            int length=(int)Class.forName("vw").getMethod("l").invoke(container);byte[] packet=new byte[length];raw.readBytes(packet);
            ByteBuf packetBytes=Unpooled.wrappedBuffer(packet);Object buffer=Class.forName("vw").getConstructor(ByteBuf.class).newInstance(packetBytes);
            Object decoded=decode.invoke(Class.forName("abe").getField("a").get(null),buffer);
            if(packetBytes.isReadable())throw new IllegalStateException("registry payload trailing bytes");
            if(Class.forName("abe").getMethod("b").invoke(decoded).equals(dimKey)) {
                int index=0;
                for(Object entry:(List<?>)Class.forName("abe").getMethod("e").invoke(decoded)) {
                    Object resource=Class.forName("kd$a").getMethod("a").invoke(entry);
                    Object tag=((Optional<?>)Class.forName("kd$a").getMethod("b").invoke(entry)).orElseThrow();
                    Object dimension=((Codec<Object>)Class.forName("dwy").getField("h").get(null)).parse((DynamicOps<Object>)Class.forName("up").getField("a").get(null),tag).getOrThrow();
                    Object registryKey=key.getMethod("a",key,id).invoke(null,dimKey,resource);
                    Class.forName("jz").getMethod("a",Class.forName("jz"),key,Object.class).invoke(null,dimensions,registryKey,dimension);
                    System.out.printf("dimension_registry %s %d %s %s %s%n",resource,index++,Class.forName("dwy").getMethod("n").invoke(dimension),Class.forName("dwy").getMethod("o").invoke(dimension),Class.forName("dwy").getMethod("g").invoke(dimension));
                }
            }
            packetBytes.release();
        }
        raw.release();
        Class.forName("jz").getMethod("m").invoke(dimensions);
        Object access=Class.forName("ka$c").getConstructor(List.class).newInstance(List.of(dimensions));
        Object creative=Class.forName("dct").getField("b").get(null);
        for(String dimension:new String[]{"minecraft:overworld","minecraft:the_end","minecraft:the_nether"})for(long seed:new long[]{0,1,-1,1234})for(boolean flat:new boolean[]{false,true})for(int previous:new int[]{-1,1}) {
            Object resource=id.getMethod("a",String.class).invoke(null,dimension);
            Object holder=((Optional<?>)Class.forName("jz").getMethod("c",id).invoke(dimensions,resource)).orElseThrow();
            Object dimensionKey=key.getMethod("a",key,id).invoke(null,Class.forName("lu").getField("ba").get(null),resource);
            long hash=(long)Class.forName("ddy").getMethod("a",long.class).invoke(null,seed);
            Object spawn=Class.forName("agd").getConstructor(Class.forName("jm"),key,long.class,Class.forName("dct"),Class.forName("dct"),boolean.class,boolean.class,Optional.class,int.class).newInstance(holder,dimensionKey,hash,creative,previous<0?null:creative,false,flat,Optional.empty(),0);
            Object respawn=Class.forName("aei").getConstructor(Class.forName("agd"),byte.class).newInstance(spawn,(byte)3);
            var out=Unpooled.buffer();Object buffer=Class.forName("wk").getConstructor(ByteBuf.class,Class.forName("ka")).newInstance(out,access);
            encode.invoke(Class.forName("aei").getField("a").get(null),buffer,respawn);
            byte[] encoded=new byte[out.readableBytes()];out.getBytes(0,encoded);
            Object decoded=decode.invoke(Class.forName("aei").getField("a").get(null),buffer);
            if(!decoded.equals(respawn)||out.isReadable())throw new IllegalStateException("respawn round trip failed");out.release();
            System.out.printf("dimension_respawn %s %d %s %d %d %s%n",dimension,seed,flat,previous,hash,HexFormat.of().formatHex(encoded));
        }
    }
}
