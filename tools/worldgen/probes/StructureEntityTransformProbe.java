package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.lang.reflect.Field;

/** Calls the original precise template transform and Entity.rotate/mirror.
 * Entity allocation avoids a ServerLevel; these methods only read yaw. */
public final class StructureEntityTransformProbe {
    static Class<?> cls(String name)throws Exception{return Class.forName(name);}
    public static void main(String[] args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);
        cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Object[] mirrors=cls("dkv").getEnumConstants(),rotations=cls("dmm").getEnumConstants();
        for(double[] p:new double[][]{{0,0,0},{0.5,2.75,3.25},{-7.3,-4,-3.125},{23.75,10,-15.5},{-2.25,9,19.5},{3.2,-2,5.8},{-0.0,0.0,-0.0},{1e100,-1e100,1e-100}})
        for(int[] pivot:new int[][]{{0,0,0},{3,5,5},{1,3,1},{3,6,7},{Integer.MAX_VALUE,0,Integer.MIN_VALUE}})
        for(int m=0;m<3;m++)for(int r=0;r<4;r++) {
            Object input=cls("exc").getConstructor(double.class,double.class,double.class).newInstance(p[0],p[1],p[2]);
            Object q=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(pivot[0],pivot[1],pivot[2]);
            Object out=cls("ent").getMethod("a",cls("exc"),cls("dkv"),cls("dmm"),cls("jd")).invoke(null,input,mirrors[m],rotations[r],q);
            System.out.printf("position %016x %016x %016x %d %d %d %d %d %016x %016x %016x%n",Double.doubleToRawLongBits(p[0]),Double.doubleToRawLongBits(p[1]),Double.doubleToRawLongBits(p[2]),pivot[0],pivot[1],pivot[2],m,r,Double.doubleToRawLongBits((Double)cls("exc").getField("c").get(out)),Double.doubleToRawLongBits((Double)cls("exc").getField("d").get(out)),Double.doubleToRawLongBits((Double)cls("exc").getField("e").get(out)));
        }
        Class<?> unsafeClass=cls("sun.misc.Unsafe");Field field=unsafeClass.getDeclaredField("theUnsafe");field.setAccessible(true);
        Object unsafe=field.get(null),actor=unsafeClass.getMethod("allocateInstance",Class.class).invoke(unsafe,cls("cmk"));
        for(float yaw:new float[]{0,-0.0f,90,-90,180,-180,270,-270,359.75f,-360.25f,721.125f,-1080.5f})
        for(int m=0;m<3;m++)for(int r=0;r<4;r++) {
            cls("bsr").getMethod("t",float.class).invoke(actor,yaw);
            float rotated=(Float)cls("bsr").getMethod("a",cls("dmm")).invoke(actor,rotations[r]);
            rotated+=(Float)cls("bsr").getMethod("a",cls("dkv")).invoke(actor,mirrors[m])-yaw;
            System.out.printf("yaw %08x %d %d %08x%n",Float.floatToRawIntBits(yaw),m,r,Float.floatToRawIntBits(rotated));
        }
    }
}
