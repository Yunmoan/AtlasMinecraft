package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.lang.reflect.Field;
import java.util.List;

/** Calls the actual Beardifier constructor and compute method on synthetic
 * rigid pieces and jigsaw junctions; does not emulate the kernel in Java.
 */
public final class StructureBeardifierProbe {
    static Class<?> cls(String name) throws Exception { return Class.forName(name); }
    static Object iterator(List<Object> values) throws Exception {
        Object list=cls("it.unimi.dsi.fastutil.objects.ObjectArrayList").getConstructor(java.util.Collection.class).newInstance(values);
        return list.getClass().getMethod("iterator").invoke(list);
    }
    static Object sample(List<Object> pieces,List<Object> junctions) throws Exception {
        return cls("dyl").getConstructor(cls("it.unimi.dsi.fastutil.objects.ObjectListIterator"),cls("it.unimi.dsi.fastutil.objects.ObjectListIterator"))
            .newInstance(iterator(pieces),iterator(junctions));
    }
    static Object piece(int adjustment,int delta) throws Exception {
        Object box=cls("ejj").getConstructor(int.class,int.class,int.class,int.class,int.class,int.class)
            .newInstance(-3,60,2,8,72,13);
        Object terrain=cls("ekc").getField(new String[]{"a","b","c","d","e"}[adjustment]).get(null);
        return cls("dyl$a").getConstructor(cls("ejj"),cls("ekc"),int.class).newInstance(box,terrain,delta);
    }
    public static void main(String[] args) throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);
        cls("ab").getMethod("a",cls("ae")).invoke(null,version);
        cls("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Field kernelField=cls("dyl").getDeclaredField("g");kernelField.setAccessible(true);
        float[] kernel=(float[])kernelField.get(null);
        for(int index=0;index<kernel.length;index++)
            System.out.printf("kernel %d %s%n",index,Integer.toUnsignedString(Float.floatToRawIntBits(kernel[index])));
        Object junction=cls("eku").getConstructor(int.class,int.class,int.class,int.class,cls("elb$a"))
            .newInstance(5,65,-2,7,cls("elb$a").getField("b").get(null));
        String[] adjustments={"none","bury","beard_thin","beard_box","encapsulate"};
        int[] coordinates={-24,-15,-12,-3,0,4,8,13,24,32};
        for(int adjustment=0;adjustment<adjustments.length;adjustment++) for(int delta:new int[]{0,3,-4}) {
            Object beard=sample(List.of(piece(adjustment,delta)),List.of());
            for(int x:coordinates) for(int dy=-18;dy<=25;dy++) for(int z:coordinates) {
                Object context=cls("dyr$e").getConstructor(int.class,int.class,int.class).newInstance(x,60+dy,z);
                double density=(Double)cls("dyl").getMethod("a",cls("dyr$b")).invoke(beard,context);
                System.out.printf("density %s %d %d %d %d %s%n",adjustments[adjustment],delta,x,60+dy,z,
                    Long.toUnsignedString(Double.doubleToRawLongBits(density)));
            }
        }
        Object junctionOnly=sample(List.of(),List.of(junction));
        Object mixed=sample(List.of(piece(1,0),piece(2,3),piece(3,-4),piece(4,0)),List.of(junction));
        for(int x:coordinates) for(int dy=-18;dy<=25;dy++) for(int z:coordinates) {
            Object context=cls("dyr$e").getConstructor(int.class,int.class,int.class).newInstance(x,60+dy,z);
            Object[] instances={junctionOnly,mixed};
            String[] names={"junction","mixed"};
            for(int index=0;index<instances.length;index++) {
                double density=(Double)cls("dyl").getMethod("a",cls("dyr$b")).invoke(instances[index],context);
                System.out.printf("density %s 0 %d %d %d %s%n",names[index],x,60+dy,z,
                    Long.toUnsignedString(Double.doubleToRawLongBits(density)));
            }
        }
    }
}
