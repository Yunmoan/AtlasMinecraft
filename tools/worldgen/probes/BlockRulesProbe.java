package atlasprobe;

import com.google.gson.Gson;
import java.io.*;
import java.lang.reflect.*;
import java.util.*;

/** Extracts state-dependent collision boxes and base block rules from the actual JAR. */
public final class BlockRulesProbe {
    public static void main(String[] args) throws Exception {
        Object version=Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a",Class.forName("ae")).invoke(null,version);
        Class.forName("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Object world=Class.forName("dcl").getField("a").get(null);
        Object origin=Class.forName("jd").getConstructor(int.class,int.class,int.class).newInstance(0,0,0);
        Method decode=Class.forName("dfy").getMethod("a",int.class);
        Method collision=Class.forName("dtb$a").getMethod("k",Class.forName("dcc"),Class.forName("jd"));
        Method replaceable=Class.forName("dtb$a").getMethod("r");
        Method free=Class.forName("diu").getMethod("m",Class.forName("dtc"));
        Method support=Class.forName("dtb$a").getMethod("d",Class.forName("dcc"),Class.forName("jd"),Class.forName("ji"));
        Method centerSupport=Class.forName("dtb$a").getMethod("a",Class.forName("dcc"),Class.forName("jd"),Class.forName("ji"),Class.forName("doa"));
        Method connectionException=Class.forName("dfy").getMethod("j",Class.forName("dtc"));
        Method offset=Class.forName("dtb$a").getMethod("q");
        Method solid=Class.forName("dtb$a").getMethod("e");
        Method solidRender=Class.forName("dtb$a").getMethod("i",Class.forName("dcc"),Class.forName("jd"));
        List<List<double[]>> shapes=new ArrayList<>();
        Map<String,Integer> shapeIDs=new HashMap<>();
        List<int[]> states=new ArrayList<>();
        Field[] axes=new Field[6]; for(int i=0;i<6;i++) axes[i]=Class.forName("ewx").getField("abcdef".substring(i,i+1));
        for(int id=0;id<26684;id++) {
            Object state=decode.invoke(null,id);
            List<double[]> boxes=new ArrayList<>();
            Object shape=collision.invoke(state,world,origin);
            for(Object box:(List<?>)Class.forName("exv").getMethod("e").invoke(shape)) {
                double[] bounds=new double[6]; for(int i=0;i<6;i++) bounds[i]=axes[i].getDouble(box);boxes.add(bounds);
            }
            String key=new Gson().toJson(boxes); Integer shapeID=shapeIDs.get(key);
            if(shapeID==null) { shapeID=shapes.size();shapeIDs.put(key,shapeID);shapes.add(boxes); }
            int flags=((boolean)replaceable.invoke(state)?1:0)|((boolean)free.invoke(null,state)?2:0)
                |((boolean)support.invoke(state,world,origin,Class.forName("ji").getField("b").get(null))?4:0)
                |((boolean)offset.invoke(state)?8:0)|((boolean)solid.invoke(state)?16:0)|((boolean)solidRender.invoke(state,world,origin)?32:0);
            for(int face=0;face<6;face++) {
                if((boolean)support.invoke(state,world,origin,Class.forName("ji").getField("abcdef".substring(face,face+1)).get(null))) flags|=1<<(6+face);
            }
            if((boolean)centerSupport.invoke(state,world,origin,Class.forName("ji").getField("b").get(null),Class.forName("doa").getField("b").get(null))) flags|=1<<12;
            if((boolean)centerSupport.invoke(state,world,origin,Class.forName("ji").getField("a").get(null),Class.forName("doa").getField("b").get(null))) flags|=1<<13;
            if((boolean)connectionException.invoke(null,state)) flags|=1<<14;
            states.add(new int[]{shapeID,flags});
        }
        Map<String,Object> output=new TreeMap<>();output.put("shapes",shapes);output.put("states",states);
        System.out.println(new Gson().toJson(output));
    }
}
