package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.*;
import java.util.jar.JarFile;

/** Real EndCityPiece template placement, marker actors, saved NBT and codecs. */
public final class EndCityBlocksProbe {
 static Class<?> cls(String n)throws Exception{return Class.forName(n);}
 static Object call(Object r,String c,String n,Class<?>[]t,Object...v)throws Exception{return cls(c).getMethod(n,t).invoke(r,v);}
 static String digest(List<String> rows)throws Exception {rows.sort(String::compareTo);return HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(String.join("\n",rows).getBytes(StandardCharsets.UTF_8)));}
 public static void main(String[]args)throws Exception {
  InitialAnimalSpawnProbe.main(new String[]{"setup"});
  Object level=InitialAnimalSpawnProbe.probeLevel,registries=IglooProbeSupport.registryAccess();
  InitialAnimalSpawnProbe.field(level,"dcw","I",registries);
  InitialAnimalSpawnProbe.field(level,"dcw","H",cls("dcw").getField("j").get(null));
  InitialAnimalSpawnProbe.field(level,"dcw","f",cls("dyz").getConstructor(long.class).newInstance(0L));
  var unsafeField=sun.misc.Unsafe.class.getDeclaredField("theUnsafe");unsafeField.setAccessible(true);var unsafe=(sun.misc.Unsafe)unsafeField.get(null);
  Object playerList=unsafe.allocateInstance(cls("apm"));InitialAnimalSpawnProbe.field(playerList,"aur","l",List.of());
  var serverField=cls("aqu").getDeclaredField("J");serverField.setAccessible(true);InitialAnimalSpawnProbe.field(serverField.get(level),"net.minecraft.server.MinecraftServer","Q",playerList);
  class HelperLoader extends ClassLoader {HelperLoader(){super(EndCityBlocksProbe.class.getClassLoader());}Class<?> helper()throws Exception {byte[] bytes=java.nio.file.Files.readAllBytes(java.nio.file.Path.of("artifacts/worldgen-probes/AtlasProbeGameEventDispatcher.class"));return defineClass("AtlasProbeGameEventDispatcher",bytes,0,bytes.length);}}
  InitialAnimalSpawnProbe.field(level,"aqu","O",new HelperLoader().helper().getConstructor(cls("aqu")).newInstance(level));
  List<String> names=new ArrayList<>();
  try(JarFile jar=new JarFile(new File(cls("ent").getProtectionDomain().getCodeSource().getLocation().toURI()))) {var entries=jar.entries();while(entries.hasMoreElements()){String n=entries.nextElement().getName();if(n.startsWith("data/minecraft/structure/end_city/")&&n.endsWith(".nbt"))names.add(n.substring("data/minecraft/structure/".length(),n.length()-4));}}
  Collections.sort(names);Object manager=IglooProbeSupport.manager(names.toArray(String[]::new));
  Object blockRegistry=cls("lt").getField("e").get(null),entityRegistry=cls("lt").getField("f").get(null);
  Object metadataAccess=cls("ka$c").getConstructor(List.class).newInstance(List.of(entityRegistry,cls("lt").getField("g").get(null)));
  for(String id:new String[]{"minecraft:shulker","minecraft:item_frame"}) {
   Object type=call(entityRegistry,"jz","a",new Class<?>[]{cls("akr")},call(null,"akr","a",new Class<?>[]{String.class},id));
   Object entity=call(type,"bsx","a",new Class<?>[]{cls("dcw")},level);call(entity,"bsr","a_",new Class<?>[]{UUID.class},new UUID(0,0));
   // LivingEntity uses process-global Math.random for its initial yaw,
   // outside every seeded worldgen stream. Canonicalize this and UUID only.
   if(id.equals("minecraft:shulker"))call(entity,"bsr","t",new Class<?>[]{float.class},0f);
   Object tag=cls("ub").getConstructor().newInstance();call(entity,"bsr","e",new Class<?>[]{cls("ub")},tag);
   GeneratedEntityProbe.print("end_city_actor_default",id,0,tag);GeneratedEntityProbe.metadata("end_city_actor_default_metadata",id,0,entity,metadataAccess);
   System.out.printf("end_city_actor_type %s %s%n",id,call(entityRegistry,"jz","a",new Class<?>[]{Object.class},type));
  }
  Map<String,Object> terrain=new HashMap<>();for(String name:new String[]{"air","end_stone","water"})terrain.put(name,call(call(blockRegistry,"jz","a",new Class<?>[]{cls("akr")},call(null,"akr","a",new Class<?>[]{String.class},"minecraft:"+name)),"dfy","o",new Class<?>[]{}));
  var getX=cls("jd").getMethod("u");var getY=cls("jd").getMethod("v");var getZ=cls("jd").getMethod("w");
  int cases=0,actorCases=0;
  for(long seed:new long[]{0,1234})for(int rotation=0;rotation<4;rotation++)for(boolean overwrite:new boolean[]{false,true})for(int scenario=0;scenario<2;scenario++)for(String fullName:names)for(int[]anchor:new int[][]{{0,64,0},{15,100,-16}}) {
   String name=fullName.substring("end_city/".length());Object base=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(anchor[0],anchor[1],anchor[2]);
   Object sample=cls("elo$a").getConstructor(cls("enu"),String.class,cls("jd"),cls("dmm"),boolean.class).newInstance(manager,name,base,cls("dmm").getEnumConstants()[rotation],overwrite);
   Object box=call(sample,"ejv","f",new Class<?>[]{});int minX=(int)call(box,"ejj","h",new Class<?>[]{}),maxX=(int)call(box,"ejj","k",new Class<?>[]{}),minZ=(int)call(box,"ejj","j",new Class<?>[]{}),maxZ=(int)call(box,"ejj","m",new Class<?>[]{});
   for(int chunkX=(minX>>4);chunkX<=(maxX>>4);chunkX++)for(int chunkZ=(minZ>>4);chunkZ<=(maxZ>>4);chunkZ++) {
    Object piece=cls("elo$a").getConstructor(cls("enu"),String.class,cls("jd"),cls("dmm"),boolean.class).newInstance(manager,name,base,cls("dmm").getEnumConstants()[rotation],overwrite);
    Map<String,Object> changed=new HashMap<>(),blockEntities=new HashMap<>();List<Object> actors=new ArrayList<>();final int mode=scenario;
    java.util.function.Function<int[],Object> read=p->{String key=p[0]+" "+p[1]+" "+p[2];if(changed.containsKey(key))return changed.get(key);if(p[1]<=63)return terrain.get("end_stone");if(mode==1&&p[1]<=110)return terrain.get("water");return terrain.get("air");};
    Object world=Proxy.newProxyInstance(cls("dds").getClassLoader(),new Class<?>[]{cls("dds")},(proxy,method,values)->{
     String m=method.getName();if(m.equals("I_"))return 0;if(m.equals("an"))return 256;if(m.equals("am"))return 256;if(m.equals("H_"))return registries;if(m.equals("C"))return seed;if(m.equals("E"))return level;if(m.equals("x_"))return false;
     if(values!=null&&values.length==1&&cls("bsr").isInstance(values[0])){actors.add(values[0]);return method.getReturnType()==boolean.class?true:null;}
     if(values!=null&&values.length>0&&cls("jd").isInstance(values[0])) {
      int x=(int)getX.invoke(values[0]),y=(int)getY.invoke(values[0]),z=(int)getZ.invoke(values[0]);String key=x+" "+y+" "+z;
      if(m.equals("a_"))return read.apply(new int[]{x,y,z});if(m.equals("b_"))return call(read.apply(new int[]{x,y,z}),"dtb$a","u",new Class<?>[]{});if(m.equals("c_"))return blockEntities.get(key);
      if(m.equals("a")&&values.length==3&&cls("dtc").isInstance(values[1])) {
       Object oldBlock=call(read.apply(new int[]{x,y,z}),"dtb$a","b",new Class<?>[]{}),block=call(values[1],"dtb$a","b",new Class<?>[]{});changed.put(key,values[1]);
       if(cls("diq").isInstance(block)){if(oldBlock!=block||!blockEntities.containsKey(key))blockEntities.put(key,call(block,"diq","a",new Class<?>[]{cls("jd"),cls("dtc")},values[0],values[1]));}else blockEntities.remove(key);return true;
      }
     }
     if(method.getReturnType()==void.class)return null;if(method.isDefault())return InvocationHandler.invokeDefault(proxy,method,values);throw new UnsupportedOperationException(method.toString());
    });
    Object clip=cls("ejj").getConstructor(int.class,int.class,int.class,int.class,int.class,int.class).newInstance(chunkX*16,0,chunkZ*16,chunkX*16+15,255,chunkZ*16+15),random=cls("dzx").getConstructor(cls("ayw")).newInstance(cls("dzz").getConstructor(long.class).newInstance(seed)),chunk=cls("dcd").getConstructor(int.class,int.class).newInstance(chunkX,chunkZ);
    call(piece,"ejv","a",new Class<?>[]{cls("dds"),cls("ddq"),cls("duz"),cls("ayw"),cls("ejj"),cls("dcd"),cls("jd")},world,null,null,random,clip,chunk,base);
    List<String> edits=new ArrayList<>(),entities=new ArrayList<>(),actorRows=new ArrayList<>();for(var e:changed.entrySet())edits.add(e.getKey()+" "+call(null,"dfy","i",new Class<?>[]{cls("dtc")},e.getValue()));
    for(var e:blockEntities.entrySet()) {Object tag=call(e.getValue(),"dqh","b",new Class<?>[]{cls("jo$a")},registries);entities.add(e.getKey()+" "+IglooProbeSupport.canonicalNBT(tag));if(seed==0&&rotation==0&&scenario==0&&overwrite&&anchor[0]==0&&name.equals("ship"))System.err.println("End City ship block entity "+e.getKey()+" "+tag);}
    for(Object actor:actors) {
     call(actor,"bsr","a_",new Class<?>[]{UUID.class},new UUID(0,0));if(cls("ckj").isInstance(actor))call(actor,"bsr","t",new Class<?>[]{float.class},0f);Object tag=cls("ub").getConstructor().newInstance();call(actor,"bsr","e",new Class<?>[]{cls("ub")},tag);actorRows.add(IglooProbeSupport.canonicalNBT(tag));
     if(seed==0&&scenario==0&&overwrite&&anchor[0]==0&&name.equals("ship")) {String id=(String)call(tag,"ub","l",new Class<?>[]{String.class},"id");GeneratedEntityProbe.print("end_city_actor_saved",id,rotation,tag);GeneratedEntityProbe.metadata("end_city_actor_metadata",id,rotation,actor,metadataAccess);if(id.equals("minecraft:item_frame")){call(actor,"bsr","e",new Class<?>[]{int.class},77);Object packet=call(actor,"cja","a",new Class<?>[]{cls("aqt")},(Object)null);System.out.printf("end_city_actor_spawn %s %d %s%n",id,rotation,GeneratedEntityPacketProbe.encode("abv",packet,metadataAccess));}}
    }
    System.out.printf("end_city_blocks %d %d %d %d %s %d %d %d %d %d %d %s %d %s %d %s %s%n",seed,rotation,overwrite?1:0,scenario,name,anchor[0],anchor[1],anchor[2],chunkX,chunkZ,edits.size(),digest(edits),entities.size(),digest(entities),actors.size(),digest(actorRows),call(random,"ayw","g",new Class<?>[]{}));cases++;actorCases+=actors.size();
   }
  }
  System.err.printf("End City blocks: %d calls, %d actors%n",cases,actorCases);
 }
}
