package atlasprobe;

import com.google.gson.Gson;
import java.io.*;
import java.lang.reflect.*;
import java.util.*;

/** Extracts state opacity/emission and exact face-union light occlusion. */
public final class LightRulesProbe {
    public static void main(String[] args) throws Exception {
        Object version=Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a",Class.forName("ae")).invoke(null,version);
        Class.forName("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Object world=Class.forName("dcl").getField("a").get(null);
        Object origin=Class.forName("jd").getConstructor(int.class,int.class,int.class).newInstance(0,0,0);
        Method decode=Class.forName("dfy").getMethod("a",int.class);
        Method opacity=Class.forName("dtb$a").getMethod("b",Class.forName("dcc"),Class.forName("jd"));
        Method emission=Class.forName("dtb$a").getMethod("h");
        Method shape=Class.forName("eov").getMethod("a",Class.forName("dcc"),Class.forName("jd"),Class.forName("dtc"),Class.forName("ji"));
        Method occludes=Class.forName("exs").getMethod("b",Class.forName("exv"),Class.forName("exv"));
        Method boxes=Class.forName("exv").getMethod("e");
        Field[] axes=new Field[6]; for(int i=0;i<6;i++) axes[i]=Class.forName("ewx").getField("abcdef".substring(i,i+1));
        Object[] directions=(Object[])Class.forName("ji").getMethod("values").invoke(null);
        List<Object> faces=new ArrayList<>(); Map<String,Integer> ids=new HashMap<>();
        Object empty=Class.forName("exs").getMethod("a").invoke(null); faces.add(empty); ids.put("[]",0);
        List<int[]> states=new ArrayList<>();
        for(int id=0;id<26684;id++) {
            Object state=decode.invoke(null,id); int[] row=new int[8];
            row[0]=(int)emission.invoke(state); row[1]=(int)opacity.invoke(state,world,origin);
            for(int i=0;i<6;i++) {
                Object face=shape.invoke(null,world,origin,state,directions[i]);
                List<double[]> bounds=new ArrayList<>();
                for(Object box:(List<?>)boxes.invoke(face)) { double[] rowBounds=new double[6]; for(int j=0;j<6;j++) rowBounds[j]=axes[j].getDouble(box); bounds.add(rowBounds); }
                String key=new Gson().toJson(bounds); Integer faceID=ids.get(key);
                if(faceID==null) { faceID=faces.size();ids.put(key,faceID);faces.add(face); }
                row[i+2]=faceID;
            }
            states.add(row);
        }
        boolean[][] blocked=new boolean[faces.size()][faces.size()];
        for(int i=0;i<faces.size();i++) for(int j=0;j<faces.size();j++) blocked[i][j]=(boolean)occludes.invoke(null,faces.get(i),faces.get(j));
        Map<String,Object> output=new TreeMap<>();output.put("states",states);output.put("blocked",blocked);
        System.out.println(new Gson().toJson(output));
    }
}
