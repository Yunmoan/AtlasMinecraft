package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.util.*;
import java.util.jar.JarFile;

/** Calls original MansionPiece postProcess on full and clipped synthetic worlds with raw brightness 15. */
public final class MansionBlocksProbe {
    static Class<?> cls(String n)throws Exception{return Class.forName(n);}
    static Object call(Object r,String c,String n,Class<?>[]t,Object...v)throws Exception{return cls(c).getMethod(n,t).invoke(r,v);}
    @SuppressWarnings("unchecked") static String actorNBT(Object actor)throws Exception {
        Object tag=cls("ub").getConstructor().newInstance();call(actor,"bsr","e",new Class<?>[]{cls("ub")},tag);
        call(tag,"ub","r",new Class<?>[]{String.class},"UUID");
        List<Object> attributes=(List<Object>)call(tag,"ub","c",new Class<?>[]{String.class,int.class},"attributes",10);
        attributes.sort(Comparator.comparing(a->{try{return (String)call(a,"ub","l",new Class<?>[]{String.class},"id");}catch(Exception e){throw new IllegalStateException(e);}}));
        return IglooProbeSupport.canonicalNBT(tag);
    }
    public static void main(String[]args)throws Exception {
        Object version=call(null,"t","a",new Class<?>[]{});call(null,"ab","a",new Class<?>[]{cls("ae")},version);call(null,"akt","a",new Class<?>[]{});
        RuinedPortalProbeSupport.bindTags();
        Object registries=IglooProbeSupport.registryAccess(),level=MonumentBlocksProbe.level(registries);
        Object difficulty=cls("bqp").getConstructor(cls("bqo"),long.class,long.class,float.class).newInstance(cls("bqo").getField("c").get(null),0L,0L,0f);
        List<String> names=new ArrayList<>();
        try(JarFile jar=new JarFile(new File(cls("ent").getProtectionDomain().getCodeSource().getLocation().toURI()))) {
            var entries=jar.entries();while(entries.hasMoreElements()){String n=entries.nextElement().getName();if(n.startsWith("data/minecraft/structure/woodland_mansion/")&&n.endsWith(".nbt"))names.add(n.substring("data/minecraft/structure/".length(),n.length()-4));}
        }
        if(args.length>1)names.removeIf(n->!n.equals("woodland_mansion/"+args[1]));
        Collections.sort(names);Object manager=IglooProbeSupport.manager(names.toArray(String[]::new));
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));System.setErr(new PrintStream(new FileOutputStream(FileDescriptor.err),true));
        Object registry=cls("lt").getField("e").get(null);Map<String,Object> terrain=new HashMap<>();
        for(String name:new String[]{"air","stone","water"})terrain.put(name,call(call(registry,"jz","a",new Class<?>[]{cls("akr")},call(null,"akr","a",new Class<?>[]{String.class},"minecraft:"+name)),"dfy","o",new Class<?>[]{}));
        var getX=cls("jd").getMethod("u");var getY=cls("jd").getMethod("v");var getZ=cls("jd").getMethod("w");
        int cases=0;
        for(long seed:new long[]{0,1234})for(int rotation=0;rotation<4;rotation++)for(int mirror=0;mirror<3;mirror++)for(int scenario=0;scenario<2;scenario++)for(String fullName:names)for(int clipMode=0;clipMode<2;clipMode++) {
            String name=fullName.substring("woodland_mansion/".length());
            Object origin=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(7,70,-9);
            Object piece=cls("emn$i").getConstructor(cls("enu"),String.class,cls("jd"),cls("dmm"),cls("dkv")).newInstance(manager,name,origin,cls("dmm").getEnumConstants()[rotation],cls("dkv").getEnumConstants()[mirror]);
            Object random=cls("dzx").getConstructor(cls("ayw")).newInstance(cls("dzz").getConstructor(long.class).newInstance(seed));
            Object regionRandom=cls("dzz").getConstructor(long.class).newInstance(seed^0x7654321L);
            Map<String,Object> changed=new HashMap<>(),blockEntities=new HashMap<>();List<Object> actors=new ArrayList<>();Set<String> ticks=new TreeSet<>();final int mode=scenario;
            java.util.function.Function<int[],Object> read=p->{String key=p[0]+" "+p[1]+" "+p[2];if(changed.containsKey(key))return changed.get(key);if(p[1]<=63)return terrain.get("stone");if(mode==1&&p[1]<=110)return terrain.get("water");return terrain.get("air");};
            Object world=Proxy.newProxyInstance(cls("dds").getClassLoader(),new Class<?>[]{cls("dds")},(proxy,method,values)->{
                String m=method.getName();if(m.equals("I_"))return -64;if(m.equals("an"))return 320;if(m.equals("am")||m.equals("J_"))return 384;if(m.equals("H_"))return registries;if(m.equals("C"))return seed;if(m.equals("E"))return level;if(m.equals("x_"))return false;if(m.equals("E_"))return regionRandom;if(m.equals("d_"))return difficulty;if(m.equals("D_"))return call(level,"dcw","D_",new Class<?>[]{});
                if(values!=null&&values.length==1&&cls("bsr").isInstance(values[0])){actors.add(values[0]);return method.getReturnType()==boolean.class?true:null;}
                if(values!=null&&values.length>0&&cls("jd").isInstance(values[0])) {
                    int x=(int)getX.invoke(values[0]),y=(int)getY.invoke(values[0]),z=(int)getZ.invoke(values[0]);String key=x+" "+y+" "+z;
                    if(m.equals("b")&&values.length==2&&values[1] instanceof Integer)return 15;
                    if(m.equals("h")&&values.length==1&&method.getReturnType()==boolean.class)return true;
                    if(m.equals("a_"))return read.apply(new int[]{x,y,z});if(m.equals("b_"))return call(read.apply(new int[]{x,y,z}),"dtb$a","u",new Class<?>[]{});if(m.equals("c_"))return blockEntities.get(key);
                    if(m.equals("a")&&values.length==3&&cls("epd").isInstance(values[1])){ticks.add(key);return null;}
                    if(m.equals("a")&&values.length==3&&cls("dtc").isInstance(values[1])) {
                        Object oldBlock=call(read.apply(new int[]{x,y,z}),"dtb$a","b",new Class<?>[]{}),block=call(values[1],"dtb$a","b",new Class<?>[]{});changed.put(key,values[1]);
                        if(cls("diq").isInstance(block)){if(oldBlock!=block||!blockEntities.containsKey(key))blockEntities.put(key,call(block,"diq","a",new Class<?>[]{cls("jd"),cls("dtc")},values[0],values[1]));}else blockEntities.remove(key);return true;
                    }
                }
                if(method.getReturnType()==void.class)return null;if(method.isDefault())return InvocationHandler.invokeDefault(proxy,method,values);throw new UnsupportedOperationException(method.toString());
            });
            int minX=clipMode==0?-128:0,minZ=clipMode==0?-128:-16,maxX=clipMode==0?128:15,maxZ=clipMode==0?128:-1;
            Object clip=cls("ejj").getConstructor(int.class,int.class,int.class,int.class,int.class,int.class).newInstance(minX,-64,minZ,maxX,319,maxZ),chunk=cls("dcd").getConstructor(int.class,int.class).newInstance(0,-1);
            call(piece,"ejv","a",new Class<?>[]{cls("dds"),cls("ddq"),cls("duz"),cls("ayw"),cls("ejj"),cls("dcd"),cls("jd")},world,null,null,random,clip,chunk,origin);
            List<String> edits=new ArrayList<>(),entities=new ArrayList<>(),actorRows=new ArrayList<>();
            for(var e:changed.entrySet())edits.add(e.getKey()+" "+call(null,"dfy","i",new Class<?>[]{cls("dtc")},e.getValue()));
            for(var e:blockEntities.entrySet())entities.add(e.getKey()+" "+IglooProbeSupport.canonicalNBT(call(e.getValue(),"dqh","b",new Class<?>[]{cls("jo$a")},registries)));
            for(Object actor:actors)actorRows.add(actorNBT(actor));
            System.out.printf("mansion_blocks %d %d %d %d %s %d %d %s %d %s %d %s %d %s %s %s%n",seed,rotation,mirror,scenario,name,clipMode,edits.size(),MonumentBlocksProbe.digest(edits),entities.size(),MonumentBlocksProbe.digest(entities),ticks.size(),MonumentBlocksProbe.digest(ticks),actors.size(),MonumentBlocksProbe.digest(actorRows),call(random,"ayw","g",new Class<?>[]{}),call(regionRandom,"ayw","g",new Class<?>[]{}));
            if(args.length>1){System.err.printf("case %d %d %d %d %s %d%n",seed,rotation,mirror,scenario,name,clipMode);edits.sort(String::compareTo);for(String row:edits)System.err.println("block "+row);for(String row:entities)System.err.println("entity "+row);}
            if(++cases%100==0)System.err.println("completed "+cases+" mansion block cases");
        }
    }
}
