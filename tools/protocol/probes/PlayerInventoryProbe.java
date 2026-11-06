package atlasprobe;

import java.io.*;
import java.nio.file.*;
import java.lang.reflect.*;
import java.util.*;
import com.google.gson.JsonParser;
import io.netty.buffer.*;

/** Read Atlas player items with the actual Java 1.21.1 ItemStack codec. */
public final class PlayerInventoryProbe {
    public static void main(String[] args) throws Exception {
        Object version=Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a",Class.forName("ae")).invoke(null,version);
        Class.forName("akt").getMethod("a").invoke(null);
        Object registries=Class.forName("pd").getMethod("a").invoke(null);
        Object lookup=registries;
        Object access=Class.forName("ka").getMethod("a",Class.forName("jz")).invoke(null,Class.forName("lt").getField("aA").get(null));
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Object accounter=Class.forName("uk").getMethod("a").invoke(null);
        Object root=Class.forName("uo").getMethod("a",Path.class,Class.forName("uk")).invoke(null,Path.of(args[0]),accounter);
        Class<?> compound=Class.forName("ub"), list=Class.forName("uh"), stack=Class.forName("cuq");
        Object inventory=compound.getMethod("c",String.class,int.class).invoke(root,"Inventory",10);
        int size=(int)list.getMethod("size").invoke(inventory);
        for(int i=0;i<size;i++) {
            Object entry=list.getMethod("a",int.class).invoke(inventory,i);
            Object item=stack.getMethod("a",Class.forName("jo$a"),compound).invoke(null,lookup,entry);
            if((boolean)stack.getMethod("e").invoke(item)) throw new IllegalStateException("vanilla rejected inventory item "+entry);
            Object serialized=stack.getMethod("a",Class.forName("jo$a")).invoke(item,lookup);
            byte[] network=(byte[])compound.getMethod("m",String.class).invoke(entry,"AtlasNetworkItem");
            if(network.length!=0) {
                ByteBuf raw=Unpooled.wrappedBuffer(network);
                try {
                    Object buffer=Class.forName("wk").getConstructor(ByteBuf.class,Class.forName("ka")).newInstance(raw,access);
                    Object expected=Class.forName("yy").getMethod("decode",Object.class).invoke(stack.getField("h").get(null),buffer);
                    if(raw.isReadable() || !(boolean)stack.getMethod("a",stack,stack).invoke(null,expected,item))
                        throw new IllegalStateException("standard NBT changed item count/components from network stack: "+entry);
                } finally { raw.release(); }
            }
            Object originalComponents=compound.getMethod("p",String.class).invoke(entry,"components");
            Object savedComponents=compound.getMethod("p",String.class).invoke(serialized,"components");
            Set<?> expectedKeys=(Set<?>)compound.getMethod("e").invoke(originalComponents);
            Set<?> savedKeys=(Set<?>)compound.getMethod("e").invoke(savedComponents);
            if(!savedKeys.containsAll(expectedKeys)) throw new IllegalStateException("vanilla discarded components "+expectedKeys+" -> "+savedKeys);
            if(expectedKeys.contains("minecraft:damage")) {
                int expected=(int)compound.getMethod("h",String.class).invoke(originalComponents,"minecraft:damage");
                if((int)stack.getMethod("n").invoke(item)!=expected) throw new IllegalStateException("damage changed");
            }
            if(expectedKeys.contains("minecraft:custom_name")) {
                String json=(String)compound.getMethod("l",String.class).invoke(originalComponents,"minecraft:custom_name");
                var expected=JsonParser.parseString(json);
                if(expected.isJsonPrimitive()) {
                    Object name=stack.getMethod("w").invoke(item);
                    String actual=(String)Class.forName("wz").getMethod("getString").invoke(name);
                    if(!actual.equals(expected.getAsString())) throw new IllegalStateException("custom name changed");
                }
            }
            System.out.println("item "+i+" "+serialized);
        }
        System.out.println("verified "+size+" nonempty inventory entries; standard NBT and network item count/components agree with vanilla ItemStack codecs");
    }
}
