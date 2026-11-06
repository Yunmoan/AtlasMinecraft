package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.util.*;
import java.util.stream.Stream;

/** Actual BiomeManager + LevelReader.getBiome + BiomeFilter at nonuniform
 * river/ocean, Nether and End quart boundaries. No Go zoom calculation. */
public final class BiomeFilterBoundaryProbe {
 static Class<?> cls(String name)throws Exception{return Class.forName(name);}
 static Object call(Object receiver,String owner,String name,Class<?>[] types,Object...args)throws Exception{return cls(owner).getMethod(name,types).invoke(receiver,args);}
 static Map<String,Object> holders(Object lookup,String field)throws Exception {
  Object registry=call(lookup,"jn$a","b",new Class<?>[]{cls("akq")},cls("lu").getField(field).get(null));
  Map<String,Object> result=new HashMap<>();
  for(Object holder:((Stream<?>)call(registry,"jo","b",new Class<?>[]{})).toList()) {
   Object key=((Optional<?>)call(holder,"jm","e",new Class<?>[]{})).orElseThrow();
   result.put(call(key,"akq","a",new Class<?>[]{}).toString(),holder);
  }
  return result;
 }
 static boolean left(int mode,int qx,int qy,int qz,int ax,int az) {
  return mode==0?qx<ax:mode==1?qz<az:Math.floorMod(qx+qy+qz,2)==0;
 }
 public static void main(String[]args)throws Exception {
  Object version=call(null,"t","a",new Class<?>[]{});call(null,"ab","a",new Class<?>[]{cls("ae")},version);call(null,"akt","a",new Class<?>[]{});
  Object registries=call(null,"pd","a",new Class<?>[]{}),lookup=call(registries,"jo$a","b",new Class<?>[]{});
  System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
  System.setErr(new PrintStream(new FileOutputStream(FileDescriptor.err),true));
  Map<String,Object> biomes=holders(lookup,"aF"),features=holders(lookup,"aQ");
  Object settings=call(call(lookup,"jn$a","b",new Class<?>[]{cls("akq")},cls("lu").getField("aO").get(null)),"jn","b",new Class<?>[]{cls("akq")},cls("dzd").getField("c").get(null));
  Object source=cls("deh").getConstructor(cls("jm")).newInstance(biomes.get("minecraft:plains"));
  Object generator=cls("dzb").getConstructor(cls("dea"),cls("jm")).newInstance(source,settings);
  Object filter=call(null,"eii","a",new Class<?>[]{});
  Method shouldPlace=cls("eii").getDeclaredMethod("a",cls("eiw"),cls("ayw"),cls("jd"));shouldPlace.setAccessible(true);
  String[][] groups={
   {"overworld","kelp_cold","river","ocean"},
   {"overworld","seagrass_river","river","ocean"},
   {"nether","crimson_forest_vegetation","nether_wastes","crimson_forest"},
   {"end","end_island_decorated","end_highlands","small_end_islands"}
  };
  for(String[]group:groups)for(long seed:new long[]{0,1,-1,4511451270791457643L})for(int modeIndex=0;modeIndex<3;modeIndex++)for(int[]anchor:new int[][]{{0,0},{-248,1072}}) {
   final int mode=modeIndex,ax=anchor[0]>>2,az=anchor[1]>>2;
   int[] selected=new int[3];String[] selectedBiome={null};
   Object noise=Proxy.newProxyInstance(cls("ddy$a").getClassLoader(),new Class<?>[]{cls("ddy$a")},(p,m,v)->{
    if(!m.getName().equals("getNoiseBiome"))throw new UnsupportedOperationException(m.toString());
    int qx=(int)v[0],qy=(int)v[1],qz=(int)v[2];selected[0]=qx;selected[1]=qy;selected[2]=qz;
    selectedBiome[0]="minecraft:"+group[left(mode,qx,qy,qz,ax,az)?2:3];return biomes.get(selectedBiome[0]);
   });
   Object manager=cls("ddy").getConstructor(cls("ddy$a"),long.class).newInstance(noise,call(null,"ddy","a",new Class<?>[]{long.class},seed));
   Object world=Proxy.newProxyInstance(cls("dds").getClassLoader(),new Class<?>[]{cls("dds")},(p,m,v)->{
    if(m.getName().equals("F_"))return manager;
    if(m.getName().equals("I_"))return -64;
    if(m.getName().equals("J_"))return 384;
    if(m.isDefault())return InvocationHandler.invokeDefault(p,m,v);
    throw new UnsupportedOperationException(m.toString());
   });
   Object feature=call(features.get("minecraft:"+group[1]),"jm","a",new Class<?>[]{});
   Object context=cls("eiw").getConstructor(cls("dds"),cls("duz"),Optional.class).newInstance(world,generator,Optional.of(feature));
   for(int y:new int[]{40,63,255})for(int dx=-4;dx<4;dx++)for(int dz=-4;dz<4;dz++) {
    int x=anchor[0]+dx,z=anchor[1]+dz;Object pos=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(x,y,z);
    Object random=cls("dzx").getConstructor(cls("ayw")).newInstance(cls("dzz").getConstructor(long.class).newInstance(seed));
    boolean allowed=(boolean)shouldPlace.invoke(filter,context,random,pos);
    System.out.printf("biome_filter %s minecraft:%s %d %d %d %d %d %d %d %d %d %d %s %s %s%n",group[0],group[1],seed,mode,anchor[0],anchor[1],x,y,z,selected[0],selected[1],selected[2],selectedBiome[0],allowed,call(random,"dzx","g",new Class<?>[]{}));
   }
  }
 }
}
