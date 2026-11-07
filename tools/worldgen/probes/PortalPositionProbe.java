package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.util.*;
import java.util.function.Predicate;

public final class PortalPositionProbe {
    private static final class WorldLoader extends ClassLoader {
        WorldLoader() { super(PortalPositionProbe.class.getClassLoader()); }
        Class<?> define(byte[] bytes) { return defineClass("PortalForcerLevel",bytes,0,bytes.length); }
        Class<?> entity(byte[] bytes) { return defineClass("PortalForcerEntity",bytes,0,bytes.length); }
    }
    public static void main(String[] args) throws Exception {
        Object version=Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a",Class.forName("ae")).invoke(null,version);
        Class.forName("akt").getMethod("a").invoke(null);
        Class<?> pos=Class.forName("jd"),vec=Class.forName("exc"),axisType=Class.forName("ji$a"),rect=Class.forName("l$a"),size=Class.forName("bsu");
        Method px=pos.getMethod("u"),py=pos.getMethod("v"),pz=pos.getMethod("w");
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        for(String axis:new String[]{"x","z"})for(int width:new int[]{2,7,21})for(int scenario=0;scenario<5;scenario++) {
            int height=width==21?21:5,bx=-7,by=30,bz=12;
            int mode=scenario;
            Predicate<Object> predicate=p->{try {
                int x=(int)(axis.equals("x")?px.invoke(p):pz.invoke(p))-(axis.equals("x")?bx:bz),y=(int)py.invoke(p)-by;
                if(x<0||x>=width||y<0||y>=height)return false;
                return switch(mode) {case 1->x!=width-1||y<3;case 2->x!=0||y>1;case 3->x!=width/2||y==2;case 4->x!=width/2||y<3;default->true;};
            }catch(Exception e){throw new RuntimeException(e);}};
            Object origin=pos.getConstructor(int.class,int.class,int.class).newInstance(bx+(axis.equals("x")?width/2:0),by+2,bz+(axis.equals("z")?width/2:0));
            Object ax=axisType.getField(axis.equals("x")?"a":"c").get(null),ay=axisType.getField("b").get(null);
            Object result=Class.forName("l").getMethod("a",pos,axisType,int.class,axisType,int.class,Predicate.class).invoke(null,origin,ax,21,ay,21,predicate);
            Object bottom=rect.getField("a").get(result);
            System.out.printf("portal_rectangle %s %d %d %d %d %d %d %d%n",axis,width,scenario,px.invoke(bottom),py.invoke(bottom),pz.invoke(bottom),rect.getField("b").get(result),rect.getField("c").get(result));
        }
        float[][] sizes={{.6f,1.8f},{1.4f,.7f},{5,6}};
        double[][] points={{0.5,1,0.5},{0.1,0.2,0.9},{0,0,0},{1.5,2,0.3},{-1,3,-1}};
        for(String axis:new String[]{"x","z"})for(int w:new int[]{1,2,21})for(int s=0;s<sizes.length;s++)for(int p=0;p<points.length;p++) {
            Object rectangle=rect.getConstructor(pos,int.class,int.class).newInstance(pos.getConstructor(int.class,int.class,int.class).newInstance(0,0,0),w,w==21?21:3);
            Object dimensions=size.getMethod("b",float.class,float.class).invoke(null,sizes[s][0],sizes[s][1]);
            Object vector=vec.getConstructor(double.class,double.class,double.class).newInstance(points[p][0],points[p][1],points[p][2]);
            Object a=axisType.getField(axis.equals("x")?"a":"c").get(null);
            Object result=Class.forName("eqe").getMethod("a",rect,axisType,vec,size).invoke(null,rectangle,a,vector,dimensions);
            System.out.printf(Locale.ROOT,"portal_relative %s %d %d %d %.17g %.17g %.17g%n",axis,w,s,p,vec.getField("c").get(result),vec.getField("d").get(result),vec.getField("e").get(result));
        }
        Class<?> level,entityClass;WorldLoader loader=new WorldLoader();
        try(InputStream stream=PortalPositionProbe.class.getResourceAsStream("/PortalForcerLevel.class")) {level=loader.define(stream.readAllBytes());}
        try(InputStream stream=PortalPositionProbe.class.getResourceAsStream("/PortalForcerEntity.class")) {entityClass=loader.entity(stream.readAllBytes());}
        Class<?> ut=Class.forName("sun.misc.Unsafe");Field uf=ut.getDeclaredField("theUnsafe");uf.setAccessible(true);Object unsafe=uf.get(null);
        Method transition=Class.forName("dlc").getDeclaredMethod("a",Class.forName("aqu"),rect,axisType,vec,Class.forName("bsr"),vec,float.class,float.class,Class.forName("eqc$a"));transition.setAccessible(true);
        double[][] relatives={{.5,0,0},{.1,.7,.25},{1,1,-.4}};
        for(String targetAxis:new String[]{"x","z"})for(String sourceAxis:new String[]{"x","z"})for(int s=0;s<sizes.length;s++)for(int r=0;r<relatives.length;r++) {
            Object world=ut.getMethod("allocateInstance",Class.class).invoke(unsafe,level);
            Object origin=pos.getConstructor(int.class,int.class,int.class).newInstance(-12,70,31);
            Object ax=axisType.getField(targetAxis.equals("x")?"a":"c").get(null);
            Object portalState=Class.forName("dte").getMethod("a",Class.forName("duf"),Comparable.class).invoke(Class.forName("dfy").getMethod("o").invoke(Class.forName("dga").getField("ed").get(null)),Class.forName("dts").getField("H").get(null),ax);
            Map<Object,Object> edits=new HashMap<>();edits.put(origin,portalState);level.getField("edits").set(world,edits);
            level.getField("collisions").set(world,new ArrayList<>());level.getField("border").set(world,Class.forName("dut").getConstructor().newInstance());
            Object dimensions=size.getMethod("b",float.class,float.class).invoke(null,sizes[s][0],sizes[s][1]);
            Object entity=ut.getMethod("allocateInstance",Class.class).invoke(unsafe,entityClass);entityClass.getField("dimensions").set(entity,dimensions);
            Object rectangle=rect.getConstructor(pos,int.class,int.class).newInstance(origin,2,3),relative=vec.getConstructor(double.class,double.class,double.class).newInstance(relatives[r][0],relatives[r][1],relatives[r][2]),velocity=vec.getConstructor(double.class,double.class,double.class).newInstance(1,2,3);
            Object result=transition.invoke(null,world,rectangle,axisType.getField(sourceAxis.equals("x")?"a":"c").get(null),relative,entity,velocity,23f,15f,Class.forName("eqc").getField("a").get(null));
            Object exit=Class.forName("eqc").getMethod("b").invoke(result),speed=Class.forName("eqc").getMethod("c").invoke(result);
            System.out.printf(Locale.ROOT,"portal_exit %s %s %d %d %.17g %.17g %.17g %.9g %.17g %.17g %.17g%n",targetAxis,sourceAxis,s,r,vec.getField("c").get(exit),vec.getField("d").get(exit),vec.getField("e").get(exit),Class.forName("eqc").getMethod("d").invoke(result),vec.getField("c").get(speed),vec.getField("d").get(speed),vec.getField("e").get(speed));
        }
        double[][][] boxes={ {},{{0,0,0,1,1,1}},{{0,0,-2,1,3,2}},{{0,0,0,1,.5,1}},{{-10,-10,-10,10,10,10}},{{-.8,0,-2,-.2,3,2},{.8,0,-2,1.2,3,2}},{{0,0,0,1,.5,1},{0,.5,.5,1,1,1}},{{-.2,0,-2,.1999999,3,2},{.8,0,-2,1.2,3,2}} };
        for(int s=0;s<sizes.length;s++)for(int p=0;p<points.length;p++)for(int b=0;b<boxes.length;b++) {
            Object world=ut.getMethod("allocateInstance",Class.class).invoke(unsafe,level);
            Object border=Class.forName("dut").getConstructor().newInstance();level.getField("border").set(world,border);
            List<Object> collision=new ArrayList<>();
            for(double[] box:boxes[b])collision.add(Class.forName("exs").getMethod("b",double.class,double.class,double.class,double.class,double.class,double.class).invoke(null,box[0],box[1],box[2],box[3],box[4],box[5]));
            level.getField("collisions").set(world,collision);
            Object dimensions=size.getMethod("b",float.class,float.class).invoke(null,sizes[s][0],sizes[s][1]);
            Object vector=vec.getConstructor(double.class,double.class,double.class).newInstance(points[p][0],points[p][1],points[p][2]);
            Object result=Class.forName("eqe").getMethod("a",vec,Class.forName("aqu"),Class.forName("bsr"),size).invoke(null,vector,world,null,dimensions);
            System.out.printf(Locale.ROOT,"portal_collision %d %d %d %.17g %.17g %.17g%n",s,p,b,vec.getField("c").get(result),vec.getField("d").get(result),vec.getField("e").get(result));
        }
        Random random=new Random(1234);
        for(int test=0;test<256;test++) {
            Object world=ut.getMethod("allocateInstance",Class.class).invoke(unsafe,level);
            level.getField("border").set(world,Class.forName("dut").getConstructor().newInstance());
            float width=(random.nextInt(32)+1)/8f,height=(random.nextInt(32)+1)/8f;
            double x=random.nextInt(33)/8.0-2,y=random.nextInt(33)/8.0-2,z=random.nextInt(33)/8.0-2;
            List<Object> collision=new ArrayList<>();List<double[]> inputs=new ArrayList<>();
            for(int b=0;b<4;b++) {
                double[] box={random.nextInt(49)/8.0-3,random.nextInt(49)/8.0-3,random.nextInt(49)/8.0-3,0,0,0};
                for(int i=0;i<3;i++)box[i+3]=box[i]+(random.nextInt(16)+1)/8.0;
                inputs.add(box);collision.add(Class.forName("exs").getMethod("b",double.class,double.class,double.class,double.class,double.class,double.class).invoke(null,box[0],box[1],box[2],box[3],box[4],box[5]));
            }
            level.getField("collisions").set(world,collision);
            Object dimensions=size.getMethod("b",float.class,float.class).invoke(null,width,height);
            Object vector=vec.getConstructor(double.class,double.class,double.class).newInstance(x,y,z);
            Object result=Class.forName("eqe").getMethod("a",vec,Class.forName("aqu"),Class.forName("bsr"),size).invoke(null,vector,world,null,dimensions);
            System.out.printf(Locale.ROOT,"portal_collision_random %.9g %.9g %.17g %.17g %.17g",width,height,x,y,z);
            for(double[] box:inputs)for(double v:box)System.out.printf(Locale.ROOT," %.17g",v);
            System.out.printf(Locale.ROOT," %.17g %.17g %.17g%n",vec.getField("c").get(result),vec.getField("d").get(result),vec.getField("e").get(result));
        }
    }
}
