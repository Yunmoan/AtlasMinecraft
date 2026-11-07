package atlasprobe;

import com.google.gson.Gson;
import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;

/** Exhaustive original BlockState mirror/rotation and template transforms. */
public final class StructureTransformProbe {
    static Class<?> cls(String name)throws Exception{return Class.forName(name);}
    public static void main(String[] args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);
        cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Object emptyWorld=cls("dcl").getField("a").get(null);
        Object origin=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(0,0,0);
        var decode=cls("dfy").getMethod("a",int.class);
        var encode=cls("dfy").getMethod("i",cls("dtc"));
        var mirror=cls("dtc").getMethod("a",cls("dkv"));
        var rotate=cls("dtc").getMethod("a",cls("dmm"));
        var fullShape=cls("dtb$a").getMethod("r",cls("dcc"),cls("jd"));
        Object[] mirrors=cls("dkv").getEnumConstants(),rotations=cls("dmm").getEnumConstants();
        List<int[]> states=new ArrayList<>();
        for(int id=0;id<26684;id++) {
            Object state=decode.invoke(null,id),block=cls("dtb$a").getMethod("b").invoke(state);
            int[] row=new int[13];
            row[0]=(!(Boolean)cls("dfy").getMethod("p").invoke(block)&&(Boolean)fullShape.invoke(state,emptyWorld,origin))?1:0;
            for(int m=0;m<3;m++)for(int r=0;r<4;r++)row[1+m*4+r]=(Integer)encode.invoke(null,rotate.invoke(mirror.invoke(state,mirrors[m]),rotations[r]));
            states.add(row);
        }
        List<int[]> positions=new ArrayList<>();
        for(int[] input:new int[][]{{0,0,0},{1,2,3},{-7,-4,-3},{23,10,-15},{-2,9,19},{3,-2,5},{Integer.MAX_VALUE,0,Integer.MIN_VALUE}})
        for(int[] pivot:new int[][]{{0,0,0},{3,5,5},{1,3,1},{3,6,7}})for(int m=0;m<3;m++)for(int r=0;r<4;r++) {
            Object p=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(input[0],input[1],input[2]);
            Object q=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(pivot[0],pivot[1],pivot[2]);
            Object transformed=cls("ent").getMethod("a",cls("jd"),cls("dkv"),cls("dmm"),cls("jd")).invoke(null,p,mirrors[m],rotations[r],q);
            positions.add(new int[]{input[0],input[1],input[2],pivot[0],pivot[1],pivot[2],m,r,
                (Integer)cls("jd").getMethod("u").invoke(transformed),(Integer)cls("jd").getMethod("v").invoke(transformed),(Integer)cls("jd").getMethod("w").invoke(transformed)});
        }
        System.out.println(new Gson().toJson(Map.of("version",3955,"states",states,"positions",positions)));
    }
}
