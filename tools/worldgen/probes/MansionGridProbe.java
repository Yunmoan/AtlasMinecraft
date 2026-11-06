package atlasprobe;

import java.io.*;
import java.lang.reflect.*;

/** Invokes the original three-floor partitioner; no replacement algorithm. */
public final class MansionGridProbe {
    static Class<?> cls(String name) throws Exception { return Class.forName(name); }
    static Object field(Object value,String name) throws Exception {
        Field f=value.getClass().getDeclaredField(name);f.setAccessible(true);return f.get(value);
    }
    public static void main(String[] args) throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);
        cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Constructor<?> ctor=cls("emn$c").getDeclaredConstructor(cls("ayw"));ctor.setAccessible(true);
        Method get=cls("emn$g").getDeclaredMethod("a",int.class,int.class);get.setAccessible(true);
        for(int i=0;i<260;i++) {
            long seed=i<256 ? i : new long[]{-1,1234,Long.MIN_VALUE,Long.MAX_VALUE}[i-256];
            Object random=cls("dyz").getConstructor(long.class).newInstance(seed);
            Object plan=ctor.newInstance(random);Object[] rooms=(Object[])field(plan,"t");
            StringBuilder values=new StringBuilder();
            for(Object grid:new Object[]{field(plan,"r"),field(plan,"s"),rooms[0],rooms[1],rooms[2]})
                for(int z=0;z<11;z++)for(int x=0;x<11;x++)values.append(' ').append(get.invoke(grid,x,z));
            System.out.printf("mansion_grid %d %d%s%n",seed,cls("ayw").getMethod("g").invoke(random),values);
        }
    }
}
