package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.util.*;

/** Calls the stock WorldBorder clamp and EndPlatformFeature runtime helper. */
public final class PortalGeometryProbe {
    public static void main(String[] args) throws Exception {
        Object version=Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a",Class.forName("ae")).invoke(null,version);
        Class.forName("akt").getMethod("a").invoke(null);
        Class<?> pos=Class.forName("jd"),borderType=Class.forName("dut");
        Method px=pos.getMethod("u"),py=pos.getMethod("v"),pz=pos.getMethod("w");
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        double[][] borders={{0,0,59999968},{-17.25,32.75,64.5},{29999980,-29999980,64},{0.5,-0.5,1}};
        double[][] points={{0,64,0},{-1.5,63.9,-0.01},{128.9,-63.01,-128.9},{29999999,255.5,-29999999},{-29999999,400,29999999},{-17.25,70,32.75},{-17.251,70,64},{16,70,-16}};
        double[][] scales={{1,8},{8,1},{1,1}};
        for(int b=0;b<borders.length;b++) {
            Object border=borderType.getConstructor().newInstance();
            borderType.getMethod("c",double.class,double.class).invoke(border,borders[b][0],borders[b][1]);
            borderType.getMethod("a",double.class).invoke(border,borders[b][2]);
            for(int s=0;s<scales.length;s++)for(int p=0;p<points.length;p++) {
                double scale=scales[s][0]/scales[s][1];
                Object target=borderType.getMethod("b",double.class,double.class,double.class).invoke(border,points[p][0]*scale,points[p][1],points[p][2]*scale);
                System.out.printf(Locale.ROOT,"portal_target %d %d %d %d %d %d%n",b,s,p,px.invoke(target),py.invoke(target),pz.invoke(target));
            }
        }
        Object spawn=Class.forName("aqu").getField("a").get(null);
        Object center=pos.getMethod("c").invoke(spawn);
        Object arrival=Class.forName("exc").getMethod("a",double.class,double.class,double.class).invoke(center,0,1,0);
        Object west=Class.forName("ji").getField("e").get(null);
        System.out.printf(Locale.ROOT,"end_arrival %d %d %d %.1f %.1f %.1f %.1f%n",px.invoke(spawn),py.invoke(spawn),pz.invoke(spawn),Class.forName("exc").getField("c").get(arrival),Class.forName("exc").getField("d").get(arrival),Class.forName("exc").getField("e").get(arrival),Class.forName("ji").getMethod("p").invoke(west));
        Object registry=Class.forName("lt").getField("e").get(null);
        Map<String,Object> states=new HashMap<>();
        for(String name:new String[]{"air","obsidian","stone"}) {
            Object block=Class.forName("jz").getMethod("a",Class.forName("akr")).invoke(registry,Class.forName("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name));
            states.put(name,Class.forName("dfy").getMethod("o").invoke(block));
        }
        for(String base:new String[]{"air","stone","obsidian"}) {
            Object world=Proxy.newProxyInstance(Class.forName("ddl").getClassLoader(),new Class<?>[]{Class.forName("ddl")},(proxy,method,values)->{
                if(method.getName().equals("a_")&&values.length==1)return states.get(base);
                if(method.getName().equals("a")&&values.length==3&&values[1]!=null&&values[1].getClass().getName().equals("dtc")) {
                    String name=values[1]==states.get("obsidian")?"obsidian":"air";
                    System.out.printf("end_platform %s %d %d %d %s%n",base,px.invoke(values[0]),py.invoke(values[0]),pz.invoke(values[0]),name);
                    return true;
                }
                if(method.getName().equals("a")&&values.length==3&&values[1] instanceof Boolean)return true;
                if(method.isDefault())return InvocationHandler.invokeDefault(proxy,method,values);
                throw new UnsupportedOperationException(method.toString());
            });
            Object origin=pos.getMethod("e").invoke(spawn);
            Class.forName("ecc").getMethod("a",Class.forName("ddl"),pos,boolean.class).invoke(null,world,origin,true);
        }
    }
}
