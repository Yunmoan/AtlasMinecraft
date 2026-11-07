package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.util.*;
import java.util.stream.Stream;

/** Original PortalShape validation, including fire, damage and filled frames. */
public final class PortalShapeProbe {
    public static void main(String[] args) throws Exception {
        Object version=Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a",Class.forName("ae")).invoke(null,version);
        Class.forName("akt").getMethod("a").invoke(null);
        Object registry=Class.forName("lt").getField("e").get(null);
        Map<String,Object> states=new HashMap<>(),blocks=new HashMap<>();
        for(String name:new String[]{"air","obsidian","stone","fire","soul_fire","nether_portal"}) {
            Object block=Class.forName("jz").getMethod("a",Class.forName("akr")).invoke(registry,Class.forName("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name));
            blocks.put(name,block);states.put(name,Class.forName("dfy").getMethod("o").invoke(block));
        }
        Map<Object,List<Object>> tags=new HashMap<>();
        for(Object tag:((Stream<?>)Class.forName("jz").getMethod("k").invoke(registry)).toList()) {
            List<Object> holders=new ArrayList<>();
            Object bound=Class.forName("jz").getMethod("b",Class.forName("awu")).invoke(registry,tag);
            if(bound instanceof Optional<?> optional && optional.isPresent())for(Object holder:(Iterable<?>)optional.get())holders.add(holder);
            tags.put(tag,holders);
        }
        List<Object> fire=new ArrayList<>();
        for(String name:new String[]{"fire","soul_fire"})fire.add(Class.forName("jz").getMethod("e",Object.class).invoke(registry,blocks.get(name)));
        tags.put(Class.forName("awe").getField("aL").get(null),fire);
        Class.forName("jz").getMethod("a",Map.class).invoke(registry,tags);
        Class<?> pos=Class.forName("jd");
        Method px=pos.getMethod("u"),py=pos.getMethod("v"),pz=pos.getMethod("w");
        Field origin=Class.forName("eqe").getDeclaredField("l"),width=Class.forName("eqe").getDeclaredField("n"),height=Class.forName("eqe").getDeclaredField("m"),count=Class.forName("eqe").getDeclaredField("k");
        for(Field field:new Field[]{origin,width,height,count})field.setAccessible(true);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        for(String axis:new String[]{"x","z"})for(int w:new int[]{1,2,3,21,22})for(int h:new int[]{2,3,4,21,22})for(int scenario=0;scenario<8;scenario++) {
            int bx=-31,by=20,bz=17;
            Map<List<Integer>,Object> edits=new HashMap<>();
            for(int x=-1;x<=w;x++)for(int y=-1;y<=h;y++) {
                if(x>=0&&x<w&&y>=0&&y<h)continue;
                if((x==-1||x==w)&&(y==-1||y==h))continue;
                edits.put(List.of(bx+(axis.equals("x")?x:0),by+y,bz+(axis.equals("z")?x:0)),states.get("obsidian"));
            }
            for(int x=0;x<w;x++)for(int y=0;y<h;y++) {
                String name=scenario==6?"nether_portal":"air";
                if(x==0&&y==0) name=switch(scenario) {case 1->"fire";case 2->"soul_fire";case 5->"nether_portal";case 7->"stone";default->name;};
                edits.put(List.of(bx+(axis.equals("x")?x:0),by+y,bz+(axis.equals("z")?x:0)),states.get(name));
            }
            if(scenario==3)edits.put(List.of(bx+(axis.equals("x")?-1:0),by+2,bz+(axis.equals("z")?-1:0)),states.get("air"));
            if(scenario==4)edits.put(List.of(bx+(axis.equals("x")?w/2:0),by+h,bz+(axis.equals("z")?w/2:0)),states.get("air"));
            Object world=Proxy.newProxyInstance(Class.forName("dcx").getClassLoader(),new Class<?>[]{Class.forName("dcx")},(proxy,method,values)->{
                if(method.getName().equals("I_"))return 0;
                if(method.getName().equals("a_")&&values.length==1)return edits.getOrDefault(List.of((int)px.invoke(values[0]),(int)py.invoke(values[0]),(int)pz.invoke(values[0])),states.get("air"));
                if(method.isDefault())return InvocationHandler.invokeDefault(proxy,method,values);
                throw new UnsupportedOperationException(method.toString());
            });
            Object p=pos.getConstructor(int.class,int.class,int.class).newInstance(bx,by,bz);
            Object axisValue=Class.forName("ji$a").getField(axis.equals("x")?"a":"c").get(null);
            Object shape=Class.forName("eqe").getConstructor(Class.forName("dcx"),pos,Class.forName("ji$a")).newInstance(world,p,axisValue);
            boolean valid=(boolean)Class.forName("eqe").getMethod("a").invoke(shape),complete=(boolean)Class.forName("eqe").getMethod("c").invoke(shape);
            Object bottom=origin.get(shape);
            System.out.printf("portal_shape %s %d %d %d %s %s %d %d %d %d %d %d%n",axis,w,h,scenario,valid,complete,valid?(int)px.invoke(bottom):0,valid?(int)py.invoke(bottom):0,valid?(int)pz.invoke(bottom):0,valid?width.getInt(shape):0,valid?height.getInt(shape):0,valid?count.getInt(shape):0);
        }
    }
}
