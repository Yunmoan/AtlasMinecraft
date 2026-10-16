package atlasprobe;
import java.io.*;
import java.lang.reflect.*;
import java.util.*;
public final class BlockOffsetProbe {
    public static void main(String[] args) throws Exception {
        Object version=Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a",Class.forName("ae")).invoke(null,version);
        Class.forName("akt").getMethod("a").invoke(null);
        Object world=Class.forName("dcl").getField("a").get(null),registry=Class.forName("lt").getField("e").get(null);
        Class<?> pos=Class.forName("jd");Object origin=pos.getConstructor(int.class,int.class,int.class).newInstance(0,0,0);
        Method decode=Class.forName("dfy").getMethod("a",int.class),has=Class.forName("dtb$a").getMethod("q"),blockOf=Class.forName("dtc").getMethod("b"),nameOf=Class.forName("jz").getMethod("b",Object.class),offset=Class.forName("dtb$a").getMethod("n",Class.forName("dcc"),pos);
        Method horizontal=Class.forName("dtb").getDeclaredMethod("av_"),vertical=Class.forName("dtb").getDeclaredMethod("ax_");horizontal.setAccessible(true);vertical.setAccessible(true);
        TreeMap<String,Object> states=new TreeMap<>();
        for(int id=0;id<26684;id++){Object state=decode.invoke(null,id);if((boolean)has.invoke(state))states.putIfAbsent(nameOf.invoke(registry,blockOf.invoke(state)).toString(),state);}
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        for(var entry:states.entrySet()) {
            Object state=entry.getValue(),block=blockOf.invoke(state),zero=offset.invoke(state,world,origin);
            System.out.printf(Locale.ROOT,"block_offset_rule %s %s %.9g %.9g%n",entry.getKey(),Class.forName("exc").getField("d").getDouble(zero)!=0,horizontal.invoke(block),vertical.invoke(block));
            for(int[] coordinates:new int[][]{{0,0,0},{1,64,2},{-31,-64,17},{123456,255,-654321}}) {
                Object p=pos.getConstructor(int.class,int.class,int.class).newInstance(coordinates[0],coordinates[1],coordinates[2]),value=offset.invoke(state,world,p);
                System.out.printf(Locale.ROOT,"block_offset %s %d %d %d %.17g %.17g %.17g%n",entry.getKey(),coordinates[0],coordinates[1],coordinates[2],Class.forName("exc").getField("c").getDouble(value),Class.forName("exc").getField("d").getDouble(value),Class.forName("exc").getField("e").getDouble(value));
            }
        }
    }
}
