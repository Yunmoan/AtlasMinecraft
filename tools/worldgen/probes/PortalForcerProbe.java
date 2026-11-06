package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.*;

public final class PortalForcerProbe {
    private static final class WorldLoader extends ClassLoader {
        WorldLoader() { super(PortalForcerProbe.class.getClassLoader()); }
        Class<?> define(byte[] bytes) { return defineClass("PortalForcerLevel",bytes,0,bytes.length); }
    }
    public static void main(String[] args) throws Exception {
        Object version=Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a",Class.forName("ae")).invoke(null,version);
        Class.forName("akt").getMethod("a").invoke(null);
        Class<?> level;
        try(InputStream stream=PortalForcerProbe.class.getResourceAsStream("/PortalForcerLevel.class")) {
            level=new WorldLoader().define(stream.readAllBytes());
        }
        Class<?> pos=Class.forName("jd");
        Class<?> unsafeType=Class.forName("sun.misc.Unsafe");
        Field uf=unsafeType.getDeclaredField("theUnsafe");uf.setAccessible(true);
        Object unsafe=uf.get(null);
        Method allocate=unsafeType.getMethod("allocateInstance",Class.class);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        int[][] bounds={{0,256,128},{-64,320,384},{0,80,80}};
        for(String axis:new String[]{"x","z"})for(int b=0;b<bounds.length;b++)for(int mode=0;mode<9;mode++) {
            Object world=allocate.invoke(unsafe,level);
            level.getField("edits").set(world,new HashMap<>());
            List<String> writes=new ArrayList<>();level.getField("writes").set(world,writes);
            level.getField("minY").setInt(world,bounds[b][0]);level.getField("maxY").setInt(world,bounds[b][1]);level.getField("logical").setInt(world,bounds[b][2]);level.getField("mode").setInt(world,mode);
            Object border=Class.forName("dut").getConstructor().newInstance();
            Class.forName("dut").getMethod("a",double.class).invoke(border,mode==7?2:1000);
            level.getField("border").set(world,border);
            Object forcer=Class.forName("eqd").getConstructor(Class.forName("aqu")).newInstance(world);
            Object target=pos.getConstructor(int.class,int.class,int.class).newInstance(0,65,0);
            Object a=Class.forName("ji$a").getField(axis.equals("x")?"a":"c").get(null);
            Optional<?> result=(Optional<?>)Class.forName("eqd").getMethod("a",pos,Class.forName("ji$a")).invoke(forcer,target,a);
            Object origin=result.isPresent()?Class.forName("l$a").getField("a").get(result.get()):target;
            String hash=HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest((String.join("\n",writes)+(writes.isEmpty()?"":"\n")).getBytes(StandardCharsets.UTF_8)));
            System.out.printf("portal_create %s %d %d %s %d %d %d %d %s%n",axis,b,mode,result.isPresent(),pos.getMethod("u").invoke(origin),pos.getMethod("v").invoke(origin),pos.getMethod("w").invoke(origin),writes.size(),hash);
        }
    }
}
