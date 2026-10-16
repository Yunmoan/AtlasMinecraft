package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.util.*;
import io.netty.buffer.Unpooled;

/** Registry IDs and actual chunk packet entity entries, including update tags. */
public final class BlockEntityPacketProbe {
    public static void main(String[] args) throws Exception {
        Object version=Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a",Class.forName("ae")).invoke(null,version);
        Class.forName("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Object types=Class.forName("lt").getField("j").get(null);
        List<String> rows=new ArrayList<>();
        for(Object type:(Iterable<?>)types)rows.add("entity_type "+Class.forName("jz").getMethod("b",Object.class).invoke(types,type)+" "+Class.forName("jz").getMethod("a",Object.class).invoke(types,type));
        rows.sort(String::compareTo);for(String row:rows)System.out.println(row);
        Class<?> pos=Class.forName("jd"),state=Class.forName("dtc"),entity=Class.forName("dqh"),info=Class.forName("adf$a");
        Object blocks=Class.forName("lt").getField("e").get(null);
        Constructor<?> entry=info.getDeclaredConstructor(int.class,int.class,Class.forName("dqj"),Class.forName("ub"));entry.setAccessible(true);
        Object registries=Class.forName("pd").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Method write=info.getDeclaredMethod("a",Class.forName("wk"));write.setAccessible(true);
        for(String[] input:new String[][]{{"brushable_block","suspicious_sand","dql"},{"chest","chest","dqo"},{"dispenser","dispenser","dra"},{"brewing_stand","brewing_stand","dqk"},{"sign","oak_wall_sign","drs"},{"mob_spawner","spawner","drw"},{"beehive","bee_nest","dqe"},{"sculk_catalyst","sculk_catalyst","dro"},{"sculk_sensor","sculk_sensor","drp"},{"sculk_shrieker","sculk_shrieker","drq"},{"end_portal","end_portal","drz"},{"end_gateway","end_gateway","dry"}}) {
            Object block=Class.forName("jz").getMethod("a",Class.forName("akr")).invoke(blocks,Class.forName("akr").getMethod("a",String.class).invoke(null,"minecraft:"+input[1]));
            Object blockState=Class.forName("dfy").getMethod("o").invoke(block);
            Object p=pos.getConstructor(int.class,int.class,int.class).newInstance(-1,-60,-16);
            Object e=Class.forName(input[2]).getConstructor(pos,state).newInstance(p,blockState);
            if(input[0].equals("beehive")) {
                Class<?> occupant=Class.forName("dqe$c");
                Method create=occupant.getDeclaredMethod("a",int.class);create.setAccessible(true);
                Class.forName("dqe").getMethod("a",occupant).invoke(e,create.invoke(null,123));
            }
            if(input[0].equals("mob_spawner"))Class.forName("drw").getMethod("a",Class.forName("bsx"),Class.forName("ayw")).invoke(e,Class.forName("bsx").getField("aN").get(null),Class.forName("dzz").getConstructor(long.class).newInstance(0L));
            var bytes=Unpooled.buffer();Object buffer=Class.forName("wk").getConstructor(io.netty.buffer.ByteBuf.class,Class.forName("ka")).newInstance(bytes,Class.forName("ka$c").getConstructor(java.util.List.class).newInstance(java.util.List.of(types)));
            Object tag=entity.getMethod("a",Class.forName("jo$a")).invoke(e,registries);
            if((boolean)Class.forName("ub").getMethod("g").invoke(tag))tag=null;
            write.invoke(entry.newInstance(240,-60,entity.getMethod("r").invoke(e),tag),buffer);
            byte[] result=new byte[bytes.readableBytes()];bytes.readBytes(result);bytes.release();
            System.out.println("entity_packet minecraft:"+input[0]+" "+HexFormat.of().formatHex(result));
            Object saved=entity.getMethod("b",Class.forName("jo$a")).invoke(e,registries);
            var saveBytes=Unpooled.buffer();
            Object saveBuffer=Class.forName("vw").getConstructor(io.netty.buffer.ByteBuf.class).newInstance(saveBytes);
            Class.forName("vw").getMethod("a",Class.forName("uy")).invoke(saveBuffer,saved);
            byte[] saveResult=new byte[saveBytes.readableBytes()];saveBytes.readBytes(saveResult);saveBytes.release();
            System.out.println("entity_default minecraft:"+input[0]+" "+HexFormat.of().formatHex(saveResult));
        }
    }
}
