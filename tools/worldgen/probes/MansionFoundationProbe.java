package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.util.*;
import java.util.jar.JarFile;

/** Actual WoodlandMansionStructure.afterPlace with synthetic floors/fluids. */
public final class MansionFoundationProbe {
    static Class<?> cls(String n)throws Exception{return Class.forName(n);}
    static Object call(Object r,String c,String n,Class<?>[]t,Object...v)throws Exception{return cls(c).getMethod(n,t).invoke(r,v);}
    public static void main(String[]args)throws Exception {
        Object version=call(null,"t","a",new Class<?>[]{});call(null,"ab","a",new Class<?>[]{cls("ae")},version);call(null,"akt","a",new Class<?>[]{});
        List<String> names=new ArrayList<>();
        try(JarFile jar=new JarFile(new File(cls("ent").getProtectionDomain().getCodeSource().getLocation().toURI()))) {
            var entries=jar.entries();while(entries.hasMoreElements()){String n=entries.nextElement().getName();if(n.startsWith("data/minecraft/structure/woodland_mansion/")&&n.endsWith(".nbt"))names.add(n.substring("data/minecraft/structure/".length(),n.length()-4));}
        }
        Object manager=IglooProbeSupport.manager(names.toArray(String[]::new));
        var uf=cls("sun.misc.Unsafe").getDeclaredField("theUnsafe");uf.setAccessible(true);
        Object structure=cls("sun.misc.Unsafe").getMethod("allocateInstance",Class.class).invoke(uf.get(null),cls("emo"));
        Map<String,Object> states=new HashMap<>();Object registry=cls("lt").getField("e").get(null);
        for(String name:new String[]{"air","stone","water","dark_oak_planks"})states.put(name,call(call(registry,"jz","a",new Class<?>[]{cls("akr")},call(null,"akr","a",new Class<?>[]{String.class},"minecraft:"+name)),"dfy","o",new Class<?>[]{}));
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        for(long seed:new long[]{0,1,1234})for(int rotation=0;rotation<4;rotation++) {
            Object random=cls("dyz").getConstructor(long.class).newInstance(seed),origin=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(7,70,-9);
            List<Object> pieces=new ArrayList<>();call(null,"emn","a",new Class<?>[]{cls("enu"),cls("jd"),cls("dmm"),List.class,cls("ayw")},manager,origin,cls("dmm").getEnumConstants()[rotation],pieces,random);
            Object container=cls("ekg").getConstructor(List.class).newInstance(pieces),box=call(container,"ekg","b",new Class<?>[]{});
            int minX=(int)call(box,"ejj","h",new Class<?>[]{}),maxX=(int)call(box,"ejj","k",new Class<?>[]{}),minZ=(int)call(box,"ejj","j",new Class<?>[]{}),maxZ=(int)call(box,"ejj","m",new Class<?>[]{}),floorY=(int)call(box,"ejj","i",new Class<?>[]{});
            for(int scenario=0;scenario<4;scenario++)for(int clipMode=0;clipMode<3;clipMode++) {
                final int mode=scenario;Map<String,Object> edits=new HashMap<>();
                java.util.function.Function<int[],Object> read=p->{String key=p[0]+" "+p[1]+" "+p[2];if(edits.containsKey(key))return edits.get(key);
                    if(p[1]==floorY)return states.get(Math.floorMod(p[0]+p[2],3)==0?"air":"dark_oak_planks");
                    int floor=mode==0?60:mode==1?30:mode==2?47+Math.floorMod(p[0]*3+p[2],7):-65;
                    if(p[1]<=floor)return states.get("stone");return states.get((mode==1||mode==2)&&p[1]<floorY?"water":"air");};
                Object world=Proxy.newProxyInstance(cls("dds").getClassLoader(),new Class<?>[]{cls("dds")},(p,m,v)->{
                    if(m.getName().equals("I_"))return -64;
                    if(v!=null&&v.length>0&&cls("jd").isInstance(v[0])) {
                        int x=(int)call(v[0],"jd","u",new Class<?>[]{}),y=(int)call(v[0],"jd","v",new Class<?>[]{}),z=(int)call(v[0],"jd","w",new Class<?>[]{});
                        if(m.getName().equals("a_"))return read.apply(new int[]{x,y,z});
                        if(m.getName().equals("a")&&v.length==3&&cls("dtc").isInstance(v[1])){edits.put(x+" "+y+" "+z,v[1]);return true;}
                    }
                    if(m.isDefault())return InvocationHandler.invokeDefault(p,m,v);throw new UnsupportedOperationException(m.toString());
                });
                int x1=minX-5,z1=minZ-5,x2=maxX+5,z2=maxZ+5;
                if(clipMode==1){x1=0;z1=-16;x2=15;z2=-1;}else if(clipMode==2){x1=(minX>>4)*16;z1=(minZ>>4)*16;x2=x1+15;z2=z1+15;}
                Object clip=cls("ejj").getConstructor(int.class,int.class,int.class,int.class,int.class,int.class).newInstance(x1,-64,z1,x2,319,z2),chunk=cls("dcd").getConstructor(int.class,int.class).newInstance(x1>>4,z1>>4);
                call(structure,"emo","a",new Class<?>[]{cls("dds"),cls("ddq"),cls("duz"),cls("ayw"),cls("ejj"),cls("dcd"),cls("ekg")},world,null,null,random,clip,chunk,container);
                List<String> blocks=new ArrayList<>();for(var e:edits.entrySet())blocks.add(e.getKey()+" "+call(null,"dfy","i",new Class<?>[]{cls("dtc")},e.getValue()));
                System.out.printf("mansion_foundation %d %d %d %d %d %s%n",seed,rotation,scenario,clipMode,blocks.size(),MonumentBlocksProbe.digest(blocks));
            }
        }
    }
}
