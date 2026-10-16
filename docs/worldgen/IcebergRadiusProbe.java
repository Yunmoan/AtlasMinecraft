package atlasprobe;
import java.io.*;
public final class IcebergRadiusProbe {
    public static void main(String[] args) throws Exception {
        Object version=Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a",Class.forName("ae")).invoke(null,version);
        Class.forName("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Object feature=Class.forName("ecr").getConstructor(com.mojang.serialization.Codec.class).newInstance(Class.forName("eeb").getField("a").get(null));
        for(long seed:new long[]{0,1,-1,42,987654321})for(int height:new int[]{6,11,17,24,36})for(int y=0;y<height;y++)for(int radius:new int[]{2,7,11})for(String method:new String[]{"a","b"}) {
            Object random=Class.forName("dzx").getConstructor(Class.forName("ayw")).newInstance(Class.forName("dzz").getConstructor(long.class).newInstance(seed));
            var m=Class.forName("ecr").getDeclaredMethod(method,Class.forName("ayw"),int.class,int.class,int.class);m.setAccessible(true);
            System.out.printf("iceberg_radius %s %d %d %d %d %d %d%n",method,seed,y,height,radius,m.invoke(feature,random,y,height,radius),Class.forName("dzx").getMethod("g").invoke(random));
        }
    }
}
