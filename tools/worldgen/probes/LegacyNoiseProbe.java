package atlasprobe;

import java.lang.reflect.*;
import java.util.*;

public final class LegacyNoiseProbe {
    public static void main(String[] args) throws Exception {
        Class<?> random=Class.forName("dyz"),source=Class.forName("ayw"),parameters=Class.forName("eoh$a"),normal=Class.forName("eoh");
        for(long seed:new long[]{0,1,-1,1234,9212937586943976949L,Long.MIN_VALUE}) {
            Object r=random.getConstructor(long.class).newInstance(seed);
            Object p=parameters.getConstructor(int.class,double.class,double[].class).newInstance(-7,1.0,new double[]{1});
            Object n=normal.getMethod("a",source,parameters).invoke(null,r,p);
            for(double[] q:new double[][]{{0,0,0},{-7500000,0,-7500000},{-257,0,-1},{123.25,48.75,-987.125},{30000000,-64,-30000000}}) {
                double v=(Double)normal.getMethod("a",double.class,double.class,double.class).invoke(n,q[0],q[1],q[2]);
                System.out.printf(Locale.ROOT,"legacy_normal %d %.17g %.17g %.17g %s%n",seed,q[0],q[1],q[2],Long.toUnsignedString(Double.doubleToRawLongBits(v)));
            }
            System.out.printf("legacy_normal_next %d %s%n",seed,Long.toUnsignedString((Long)random.getMethod("g").invoke(r)));
        }
    }
}
