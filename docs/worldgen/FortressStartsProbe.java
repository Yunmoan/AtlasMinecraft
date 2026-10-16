package atlasprobe;

import java.io.*;
import java.util.*;

/** Invokes the original fortress piece assembler and serializes all saved pieces. */
public final class FortressStartsProbe {
    static Class<?> cls(String name)throws Exception{return Class.forName(name);}
    public static void main(String[] args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);
        cls("ab").getMethod("a",cls("ae")).invoke(null,version);
        cls("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        for(long seed:new long[]{0,1,-1,1234,Long.MIN_VALUE,Long.MAX_VALUE,9212937586943976949L})
        for(int[] chunk:new int[][]{{0,0},{1,-1},{-1,1},{17,-9},{-41,23},{31,31},{-33,32},{145,-62}}) {
            Object random=cls("dyz").getConstructor(long.class).newInstance(seed);
            long mx=(Long)cls("ayw").getMethod("g").invoke(random),mz=(Long)cls("ayw").getMethod("g").invoke(random);
            cls("ayw").getMethod("b",long.class).invoke(random,(long)chunk[0]*mx^(long)chunk[1]*mz^seed);
            Object builder=cls("ekj").getConstructor().newInstance();
            Object start=cls("elx$q").getConstructor(cls("ayw"),int.class,int.class).newInstance(random,chunk[0]*16+2,chunk[1]*16+2);
            var add=cls("ejw").getMethod("a",cls("ejv"));add.invoke(builder,start);
            var children=cls("ejv").getMethod("a",cls("ejv"),cls("ejw"),cls("ayw"));children.invoke(start,start,builder,random);
            List<?> pending=(List<?>)cls("elx$q").getField("d").get(start);
            while(!pending.isEmpty()) {
                int index=(Integer)cls("ayw").getMethod("a",int.class).invoke(random,pending.size());
                Object piece=pending.remove(index);children.invoke(piece,start,builder,random);
            }
            cls("ekj").getMethod("a",cls("ayw"),int.class,int.class).invoke(builder,random,48,70);
            Object container=cls("ekj").getMethod("a").invoke(builder);
            List<?> pieces=(List<?>)cls("ekg").getMethod("c").invoke(container);
            System.out.printf("start %d %d %d %d%n",seed,chunk[0],chunk[1],pieces.size());
            for(Object piece:pieces) {
                Object tag=cls("ejv").getMethod("a",cls("ekh")).invoke(piece,(Object)null);
                System.out.printf("piece %s %d %d",cls("ub").getMethod("l",String.class).invoke(tag,"id"),cls("ub").getMethod("h",String.class).invoke(tag,"O"),cls("ub").getMethod("h",String.class).invoke(tag,"GD"));
                for(int value:(int[])cls("ub").getMethod("n",String.class).invoke(tag,"BB"))System.out.print(" "+value);
                for(String name:new String[]{"Seed","Chest","Mob"}) {
                    boolean exists=(Boolean)cls("ub").getMethod("e",String.class).invoke(tag,name);
                    System.out.print(" "+(exists?name.equals("Seed")?cls("ub").getMethod("h",String.class).invoke(tag,name):((Boolean)cls("ub").getMethod("q",String.class).invoke(tag,name)?1:0):"-"));
                }
                System.out.println();
            }
            System.out.printf("next %d%n",cls("ayw").getMethod("g").invoke(random));
        }
    }
}
