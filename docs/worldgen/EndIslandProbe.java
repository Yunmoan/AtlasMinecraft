package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.util.*;

/** Official End island density and its independently seeded 2D simplex noise. */
public final class EndIslandProbe {
    public static void main(String[] args) throws Exception {
        Object version=Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a",Class.forName("ae")).invoke(null,version);
        Class.forName("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Class<?> density=Class.forName("dys$i"), context=Class.forName("dyr$b");
        Constructor<?> constructor=density.getDeclaredConstructor(long.class);constructor.setAccessible(true);
        Method sample=density.getDeclaredMethod("a",context);sample.setAccessible(true);
        Class<?> simplex=Class.forName("eok");
        Method noise=simplex.getMethod("a",double.class,double.class);
        List<int[]> points=new ArrayList<>();
        int[] borders={0,1,-1,7,-7,8,-8,511,-511,512,-512,1023,-1023,1024,-1024,30000000,-30000000,Integer.MIN_VALUE,Integer.MAX_VALUE};
        for(int x:borders)for(int z:borders)points.add(new int[]{x,z});
        Random positions=new Random(767);
        for(int i=0;i<512;i++)points.add(new int[]{positions.nextInt(60000001)-30000000,positions.nextInt(60000001)-30000000});
        for(long seed:new long[]{0,1,-1,1234,9212937586943976949L,Long.MIN_VALUE}) {
            Object islands=constructor.newInstance(seed);
            for(int[] p:points) {
                Object point=Proxy.newProxyInstance(context.getClassLoader(),new Class<?>[]{context},(proxy,method,unused)-> switch(method.getName()) {
                    case "a" -> p[0]; case "b" -> 0; case "c" -> p[1];
                    default -> throw new UnsupportedOperationException(method.getName());
                });
                double value=(Double)sample.invoke(islands,point);
                System.out.printf("end_islands %d %d %d %s%n",seed,p[0],p[1],Long.toUnsignedString(Double.doubleToRawLongBits(value)));
            }
            Object random=Class.forName("dyz").getConstructor(long.class).newInstance(seed);
            Object n=simplex.getConstructor(Class.forName("ayw")).newInstance(random);
            for(double[] p:new double[][]{{0,0},{0.125,-0.75},{-17.25,92.5},{30000000,-30000000},{-2147483648.0,2147483647.0}}) {
                double value=(Double)noise.invoke(n,p[0],p[1]);
                System.out.printf(Locale.ROOT,"simplex2d %d %.17g %.17g %s%n",seed,p[0],p[1],Long.toUnsignedString(Double.doubleToRawLongBits(value)));
            }
        }
    }
}
