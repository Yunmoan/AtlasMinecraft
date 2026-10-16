package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.lang.reflect.Field;
import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Proxy;
import java.util.HashMap;
import java.util.List;

/** Original BlockRot/ignore/capped ocean archaeology processor chain. */
public final class OceanRuinProcessorsProbe {
    static Class<?> cls(String name)throws Exception{return Class.forName(name);}
    static void field(Object receiver,String owner,String name,Object value)throws Exception{Field field=cls(owner).getDeclaredField(name);field.setAccessible(true);field.set(receiver,value);}
    static Object seedLevel(long seed)throws Exception {
        Field field=sun.misc.Unsafe.class.getDeclaredField("theUnsafe");field.setAccessible(true);var unsafe=(sun.misc.Unsafe)field.get(null);Object server=unsafe.allocateInstance(cls("apn")),level=unsafe.allocateInstance(cls("aqu"));
        Object options=cls("dzw").getConstructor(long.class,boolean.class,boolean.class).newInstance(seed,true,false),data=Proxy.newProxyInstance(cls("erl").getClassLoader(),new Class<?>[]{cls("erl")},(proxy,method,args)->{if(method.getName().equals("y"))return options;if(method.getName().equals("K"))return cls("cpn").getField("g").get(null);throw new UnsupportedOperationException(method.toString());});
        field(server,"net.minecraft.server.MinecraftServer","j",data);field(level,"aqu","J",server);return level;
    }
    public static void main(String[] args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);RuinedPortalProbeSupport.bindTags();Object registries=IglooProbeSupport.registryAccess(),manager=IglooProbeSupport.manager(OceanRuinProbeSupport.names());System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        for(String name:OceanRuinProbeSupport.names())for(int rotation=0;rotation<4;rotation++)for(float integrity:new float[]{0.5f,0.7f,0.8f,0.9f})for(int cold=0;cold<2;cold++)for(long seed:new long[]{0,1,-1})for(int[] anchor:new int[][]{{0,60,0},{-16,64,-16},{15,69,15}}) {
            Object origin=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(anchor[0],anchor[1],anchor[2]),key=cls("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name),piece=cls("emd$a").getConstructor(cls("enu"),cls("akr"),cls("jd"),cls("dmm"),float.class,cls("eme$a"),boolean.class).newInstance(manager,key,origin,cls("dmm").getEnumConstants()[rotation],integrity,cls("eme$a").getEnumConstants()[cold],false),template=cls("ekb").getMethod("c").invoke(piece);
            Field settingsField=cls("ekb").getDeclaredField("c");settingsField.setAccessible(true);Object settings=settingsField.get(piece);
            Field palettesField=cls("ent").getDeclaredField("l");palettesField.setAccessible(true);Object palette=cls("enp").getMethod("a",List.class,cls("jd")).invoke(settings,palettesField.get(template),origin);List<?> blocks=(List<?>)cls("ent$a").getMethod("a").invoke(palette);
            Object delegate=RuinedPortalProbeSupport.world(0,new HashMap<>(),new HashMap<>(),registries),level=seedLevel(seed),world=Proxy.newProxyInstance(cls("dds").getClassLoader(),new Class<?>[]{cls("dds")},(proxy,method,values)->{if(method.getName().equals("E")&&method.getParameterCount()==0)return level;try{return method.invoke(delegate,values);}catch(InvocationTargetException error){throw error.getCause();}});
            List<?> processed=(List<?>)cls("ent").getMethod("a",cls("ddl"),cls("jd"),cls("jd"),cls("enp"),List.class).invoke(null,world,origin,origin,settings,blocks);StringBuilder text=new StringBuilder();
            for(Object block:processed){Object position=cls("ent$c").getMethod("a").invoke(block),state=cls("ent$c").getMethod("b").invoke(block),tag=cls("ent$c").getMethod("c").invoke(block);text.append(cls("jd").getMethod("u").invoke(position)).append(' ').append(cls("jd").getMethod("v").invoke(position)).append(' ').append(cls("jd").getMethod("w").invoke(position)).append(' ').append(cls("dfy").getMethod("i",cls("dtc")).invoke(null,state)).append(' ').append(tag==null?"none":IglooProbeSupport.canonicalNBT(tag)).append('\n');}
            System.out.printf("oceanprocess %s %d %s %d %d %d %d %d %d %s%n",name,rotation,Float.toString(integrity),cold,seed,anchor[0],anchor[1],anchor[2],processed.size(),RuinedPortalStartsProbe.digest(text.toString()));
        }
    }
}
