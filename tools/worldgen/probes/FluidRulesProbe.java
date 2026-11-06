package atlasprobe;

import com.google.gson.Gson;
import java.io.*;
import java.lang.reflect.*;
import java.util.*;

/** Actual FlowingFluid predicates and collision-face tests from the pinned JAR. */
public final class FluidRulesProbe {
    static Class<?> cls(String n)throws Exception{return Class.forName(n);}
    public static void main(String[]args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);
        cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);
        RuinedPortalProbeSupport.bindTags();
        Object world=cls("dcl").getField("a").get(null);
        Constructor<?> pos=cls("jd").getConstructor(int.class,int.class,int.class);
        Object origin=pos.newInstance(0,0,0);
        Object[] fluids=new Object[5];for(int i=0;i<5;i++)fluids[i]=cls("epf").getField(String.valueOf((char)('a'+i))).get(null);
        Method hold=cls("epc").getDeclaredMethod("a",cls("dcc"),cls("jd"),cls("dtc"),cls("epd"));hold.setAccessible(true);
        Method wall=cls("epc").getDeclaredMethod("a",cls("ji"),cls("dcc"),cls("jd"),cls("dtc"),cls("jd"),cls("dtc"));wall.setAccessible(true);
        Method decode=cls("dfy").getMethod("a",int.class),collision=cls("dtb$a").getMethod("k",cls("dcc"),cls("jd"));
        Method getFluid=cls("dtb$a").getMethod("u"),fluidType=cls("epe").getMethod("a"),amount=cls("epe").getMethod("e");
        Method property=cls("dtc").getMethod("c",cls("duf"));Object falling=cls("epc").getField("a").get(null);
        List<Integer> representatives=new ArrayList<>();Set<String> shapes=new HashSet<>();
        List<double[]> slices=new ArrayList<>();Map<String,Integer> sliceIDs=new HashMap<>();List<Integer> sliceIndices=new ArrayList<>();
        Method coords=cls("exv").getDeclaredMethod("a",cls("ji$a"));coords.setAccessible(true);
        Object[] axes=cls("ji$a").getEnumConstants();
        byte[] rules=new byte[26684*2];Object[] states=new Object[26684];
        Field[] bounds=new Field[6];for(int i=0;i<6;i++)bounds[i]=cls("ewx").getField(String.valueOf((char)('a'+i)));
        Gson gson=new Gson();
        for(int id=0;id<states.length;id++) {
            Object state=decode.invoke(null,id);states[id]=state;
            Object fs=getFluid.invoke(state),type=fluidType.invoke(fs);int kind=-1;
            for(int i=0;i<5;i++)if(type==fluids[i])kind=i;
            if(kind<0)throw new IllegalStateException("unknown fluid");
            boolean falls=kind!=0&&(boolean)property.invoke(fs,falling);
            rules[id*2]=(byte)(kind|((int)amount.invoke(fs)<<3)|(falls?128:0));
            int flags=0;
            for(int i=1;i<5;i++)if((boolean)hold.invoke(fluids[2],world,origin,state,fluids[i]))flags|=1<<(i-1);
            if((boolean)hold.invoke(fluids[2],world,origin,state,fluids[0]))flags|=128;
            if((boolean)cls("dtb$a").getMethod("e").invoke(state))flags|=16;
            if((boolean)cls("dtb$a").getMethod("d").invoke(state))flags|=32;
            if(cls("dkp").isInstance(cls("dtb$a").getMethod("b").invoke(state)))flags|=64;
            rules[id*2+1]=(byte)flags;
            Object shape=collision.invoke(state,world,origin);
            List<double[]> boxes=new ArrayList<>();
            for(Object box:(List<?>)cls("exv").getMethod("e").invoke(shape)) {
                double[] values=new double[6];for(int i=0;i<6;i++)values[i]=bounds[i].getDouble(box);boxes.add(values);
            }
            double[] slice=new double[6];
            for(int axis=0;axis<3;axis++) {
                List<?> values=(List<?>)coords.invoke(shape,axes[axis]);
                if(values.size()>1) {
                    slice[axis]=((Number)values.get(0)).doubleValue()*.5+((Number)values.get(1)).doubleValue()*.5;
                    slice[axis+3]=((Number)values.get(values.size()-2)).doubleValue()*.5+((Number)values.get(values.size()-1)).doubleValue()*.5;
                }
            }
            String sliceKey=gson.toJson(slice);Integer sliceID=sliceIDs.get(sliceKey);
            if(sliceID==null) {sliceID=slices.size();sliceIDs.put(sliceKey,sliceID);slices.add(slice);}
            sliceIndices.add(sliceID);
            if(shapes.add(gson.toJson(boxes)+sliceKey))representatives.add(id);
        }
        int[][] offsets={{0,-1,0},{0,1,0},{0,0,-1},{0,0,1},{-1,0,0},{1,0,0}};
        Object[] directions=cls("ji").getEnumConstants();
        byte[] passages=new byte[representatives.size()*representatives.size()*6];int index=0;
        for(int first:representatives)for(int second:representatives)for(int face=0;face<6;face++) {
            int[] o=offsets[face];Object other=pos.newInstance(o[0],o[1],o[2]);
            passages[index++]=(byte)((boolean)wall.invoke(fluids[2],directions[face],world,origin,states[first],other,states[second])?1:0);
        }
        Map<String,Object> result=new TreeMap<>();
        result.put("rules",Base64.getEncoder().encodeToString(rules));result.put("wall_states",representatives);
        result.put("slice_points",slices);result.put("slice_indices",sliceIndices);
        result.put("walls",Base64.getEncoder().encodeToString(passages));
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        System.out.println("fluid_rules "+gson.toJson(result));
    }
}
