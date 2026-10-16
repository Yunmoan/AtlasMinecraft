package atlasprobe;
import java.io.*;import java.lang.reflect.*;import java.util.*;
public final class DrownedActorProbe {
 static Class<?> cls(String name)throws Exception{return Class.forName(name);}
 static Object call(Object receiver,String owner,String name,Class<?>[] types,Object...v)throws Exception{return cls(owner).getMethod(name,types).invoke(receiver,v);}
 public static void main(String[]args)throws Exception {
  InitialAnimalSpawnProbe.main(new String[]{"setup"});Object level=InitialAnimalSpawnProbe.probeLevel,access=call(level,"dcw","H_",new Class<?>[]{}),registry=cls("lt").getField("f").get(null),type=call(registry,"jz","a",new Class<?>[]{cls("akr")},call(null,"akr","a",new Class<?>[]{String.class},"minecraft:drowned"));
  class HelperLoader extends ClassLoader {HelperLoader(){super(DrownedActorProbe.class.getClassLoader());}Class<?> helper()throws Exception {byte[] bytes=java.nio.file.Files.readAllBytes(java.nio.file.Path.of("artifacts/worldgen-probes/AtlasProbeGameEventDispatcher.class"));return defineClass("AtlasProbeGameEventDispatcher",bytes,0,bytes.length);}}
  InitialAnimalSpawnProbe.field(level,"aqu","O",new HelperLoader().helper().getConstructor(cls("aqu")).newInstance(level));
  Object initial=call(type,"bsx","a",new Class<?>[]{cls("dcw")},level);call(initial,"bsr","a_",new Class<?>[]{UUID.class},new UUID(0,0));Object saved=cls("ub").getConstructor().newInstance();call(initial,"bsr","e",new Class<?>[]{cls("ub")},saved);GeneratedEntityProbe.print("drowned_actor_default","minecraft:drowned",0,saved);GeneratedEntityProbe.metadata("drowned_actor_default_metadata","minecraft:drowned",0,initial,access);System.out.printf("drowned_actor_type minecraft:drowned %s%n",call(registry,"jz","a",new Class<?>[]{Object.class},type));
  for(int index=0;index<4096;index++) {
   long seed=index*987654321L,actorSeed=seed^0x123456789abcdefL;Object random=cls("dzz").getConstructor(long.class).newInstance(seed),actor=cls("dyz").getConstructor(long.class).newInstance(actorSeed),biome=InitialAnimalSpawnProbe.probeBiomes.get("minecraft:ocean");List<Object> extras=new ArrayList<>();
   Object world=Proxy.newProxyInstance(cls("ddl").getClassLoader(),new Class<?>[]{cls("ddl")},(p,m,v)->{
    if(m.getName().equals("E_"))return random;if(m.getName().equals("E"))return level;if(m.getName().equals("H_"))return access;if(m.getName().equals("d_"))return InitialAnimalSpawnProbe.probeDifficulty;if(m.getName().equals("t")||m.getName().equals("getNoiseBiome"))return biome;
    if(m.getName().equals("a")&&v!=null&&v.length==3&&v[0] instanceof Class)return List.of();
    if(m.getName().equals("b")&&v!=null&&v.length==1&&cls("bsr").isInstance(v[0])) {extras.add(v[0]);return true;}
    if(m.getName().equals("a_")&&v!=null&&v.length==1&&cls("jd").isInstance(v[0]))return InitialAnimalSpawnProbe.probeAir;
    if(m.isDefault())return InvocationHandler.invokeDefault(p,m,v);throw new UnsupportedOperationException(m.toString());
   });
   Object entity=call(type,"bsx","a",new Class<?>[]{cls("dcw")},level);InitialAnimalSpawnProbe.field(entity,"bsr","ah",actor);call(entity,"bsr","a_",new Class<?>[]{UUID.class},new UUID(0,0));call(entity,"btp","fR",new Class<?>[]{});call(entity,"bsr","b",new Class<?>[]{double.class,double.class,double.class,float.class,float.class},-15.5,50.0,16.5,0f,0f);
   call(entity,"btp","a",new Class<?>[]{cls("ddl"),cls("bqp"),cls("btr"),cls("buh")},world,InitialAnimalSpawnProbe.probeDifficulty,cls("btr").getField("d").get(null),null);
   saved=cls("ub").getConstructor().newInstance();call(entity,"bsr","d",new Class<?>[]{cls("ub")},saved);GeneratedEntityProbe.print("drowned_actor_finalized","minecraft:drowned",seed,saved);GeneratedEntityProbe.metadata("drowned_actor_finalized_metadata","minecraft:drowned",seed,entity,access);
   System.out.printf("drowned_actor_next %d %d %s %s %d%n",seed,actorSeed,call(random,"dzz","g",new Class<?>[]{}),call(actor,"dyz","g",new Class<?>[]{}),extras.size());
   for(Object extra:extras){saved=cls("ub").getConstructor().newInstance();call(extra,"bsr","e",new Class<?>[]{cls("ub")},saved);GeneratedEntityProbe.print("drowned_actor_extra","minecraft:chicken",seed,saved);}
  }
 }
}
