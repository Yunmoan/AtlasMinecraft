package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.util.*;
import java.util.function.Supplier;
import java.util.stream.Stream;

/** Runs the official NaturalSpawner generation pass against a controlled flat region. */
public final class InitialAnimalSpawnProbe {
 static Object probeLevel,probeLookup,probeGrass,probeAir,probeBorder,probeDimension,probeDifficulty;
 static Map<String,Object> probeBiomes;
 static Class<?> cls(String name)throws Exception{return Class.forName(name);}
 static Object call(Object receiver,String owner,String name,Class<?>[] types,Object...values)throws Exception{return cls(owner).getMethod(name,types).invoke(receiver,values);}
 static void field(Object receiver,String owner,String name,Object value)throws Exception {Field f=cls(owner).getDeclaredField(name);f.setAccessible(true);f.set(receiver,value);}
 public static void main(String[] args)throws Exception {
  Object version=call(null,"t","a",new Class<?>[]{});call(null,"ab","a",new Class<?>[]{cls("ae")},version);call(null,"akt","a",new Class<?>[]{});
  Object registries=call(null,"pd","a",new Class<?>[]{}),lookup=call(registries,"jo$a","b",new Class<?>[]{});
  System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));System.setErr(new PrintStream(new FileOutputStream(FileDescriptor.err),true));
  Object biomes=call(lookup,"jn$a","b",new Class<?>[]{cls("akq")},cls("lu").getField("aF").get(null)),dimensions=call(lookup,"jn$a","b",new Class<?>[]{cls("akq")},cls("lu").getField("aK").get(null));
  Map<String,Object> holders=new LinkedHashMap<>();Object dimension=null;
  for(Object holder:((Stream<?>)call(biomes,"jo","b",new Class<?>[]{})).toList()) {Object key=((Optional<?>)call(holder,"jm","e",new Class<?>[]{})).orElseThrow();String name=call(key,"akq","a",new Class<?>[]{}).toString();holders.put(name,holder);Object biome=call(holder,"jm","a",new Class<?>[]{}),settings=call(biome,"ddw","b",new Class<?>[]{}),list=call(settings,"dei","a",new Class<?>[]{cls("btq")},cls("btq").getField("b").get(null));System.out.printf("creature_settings %s %s %s%n",name,call(settings,"dei","a",new Class<?>[]{}),call(list,"bpf","e",new Class<?>[]{}));}
  for(Object holder:((Stream<?>)call(dimensions,"jo","b",new Class<?>[]{})).toList()) {Object key=((Optional<?>)call(holder,"jm","e",new Class<?>[]{})).orElseThrow();if(call(key,"akq","a",new Class<?>[]{}).toString().equals("minecraft:overworld"))dimension=holder;}
  Field unsafeField=sun.misc.Unsafe.class.getDeclaredField("theUnsafe");unsafeField.setAccessible(true);var unsafe=(sun.misc.Unsafe)unsafeField.get(null);
  Object server=unsafe.allocateInstance(cls("apn")),level=unsafe.allocateInstance(cls("aqu"));
  Object worldData=Proxy.newProxyInstance(cls("erl").getClassLoader(),new Class<?>[]{cls("erl")},(p,m,v)->{if(m.getName().equals("K"))return cls("cpn").getField("g").get(null);throw new UnsupportedOperationException(m.toString());});
  field(server,"net.minecraft.server.MinecraftServer","j",worldData);field(level,"aqu","J",server);field(level,"dcw","D",dimension);field(level,"dcw","I",call(null,"ka","a",new Class<?>[]{cls("jz")},cls("lt").getField("a").get(null)));field(level,"dcw","E",(Supplier<Object>)()->{try{return cls("bnc").getField("a").get(null);}catch(Exception e){throw new RuntimeException(e);}});
  Object levelData=Proxy.newProxyInstance(cls("erm").getClassLoader(),new Class<?>[]{cls("erm")},(p,m,v)->{if(m.getReturnType()==long.class)return 0L;if(m.getReturnType()==boolean.class)return false;if(m.getReturnType()==int.class)return 0;throw new UnsupportedOperationException(m.toString());});field(level,"dcw","A",levelData);
  Object grass=call(cls("dga").getField("i").get(null),"dfy","o",new Class<?>[]{}),air=call(cls("dga").getField("a").get(null),"dfy","o",new Class<?>[]{}),border=cls("dut").getConstructor().newInstance(),dim=call(dimension,"jm","a",new Class<?>[]{}),difficulty=cls("bqp").getConstructor(cls("bqo"),long.class,long.class,float.class).newInstance(cls("bqo").getField("c").get(null),0L,0L,0f);
  Object blockRegistry=cls("lt").getField("e").get(null),grassHolder=call(blockRegistry,"jz","e",new Class<?>[]{Object.class},cls("dga").getField("i").get(null));
  List<Object> rails=new ArrayList<>();for(String name:new String[]{"rail","powered_rail","detector_rail","activator_rail"}) {Object block=call(blockRegistry,"jz","a",new Class<?>[]{cls("akr")},call(null,"akr","a",new Class<?>[]{String.class},"minecraft:"+name));rails.add(call(blockRegistry,"jz","e",new Class<?>[]{Object.class},block));}
  call(blockRegistry,"jz","a",new Class<?>[]{Map.class},Map.of(cls("awe").getField("bX").get(null),List.of(grassHolder),cls("awe").getField("aY").get(null),rails));
  probeLevel=level;probeLookup=lookup;probeGrass=grass;probeAir=air;probeBorder=border;probeDimension=dim;probeDifficulty=difficulty;probeBiomes=holders;
  if(args.length>0&&args[0].equals("setup"))return;
  Object entityRegistry=cls("lt").getField("f").get(null);
  List<Object> types=new ArrayList<>();
  for(String id:new String[]{"minecraft:sheep","minecraft:pig","minecraft:chicken","minecraft:cow","minecraft:horse","minecraft:donkey"}) {
   Object type=call(entityRegistry,"jz","a",new Class<?>[]{cls("akr")},call(null,"akr","a",new Class<?>[]{String.class},id));
   types.add(type);
   Object entity=call(type,"bsx","a",new Class<?>[]{cls("dcw")},level);call(entity,"bsr","a_",new Class<?>[]{UUID.class},new UUID(0,0));
   Object tag=cls("ub").getConstructor().newInstance();call(entity,"bsr","e",new Class<?>[]{cls("ub")},tag);GeneratedEntityProbe.print("animal_default",id,0,tag);
   Object access=call(level,"dcw","H_",new Class<?>[]{});GeneratedEntityProbe.metadata("animal_default_metadata",id,0,entity,access);
   for(int scenario=0;scenario<16;scenario++) {
    Object altered=cls("ub").getConstructor().newInstance();call(entity,"bsr","e",new Class<?>[]{cls("ub")},altered);
    call(altered,"ub","a",new Class<?>[]{String.class,float.class},"Health",scenario==0?1f:3.5f);
    call(altered,"ub","a",new Class<?>[]{String.class,boolean.class},"NoAI",(scenario&1)!=0);call(altered,"ub","a",new Class<?>[]{String.class,boolean.class},"LeftHanded",(scenario&2)!=0);call(altered,"ub","a",new Class<?>[]{String.class,int.class},"Age",(scenario&4)!=0?-24000:0);
    if(id.equals("minecraft:sheep")){call(altered,"ub","a",new Class<?>[]{String.class,byte.class},"Color",(byte)scenario);call(altered,"ub","a",new Class<?>[]{String.class,boolean.class},"Sheared",(scenario&8)!=0);}
    if(id.equals("minecraft:pig"))call(altered,"ub","a",new Class<?>[]{String.class,boolean.class},"Saddle",(scenario&8)!=0);
    if(id.equals("minecraft:horse")||id.equals("minecraft:donkey"))call(altered,"ub","a",new Class<?>[]{String.class,boolean.class},"Tame",(scenario&8)!=0);
    if(id.equals("minecraft:horse"))call(altered,"ub","a",new Class<?>[]{String.class,int.class},"Variant",scenario%7|(scenario%5)<<8);
    if(id.equals("minecraft:donkey"))call(altered,"ub","a",new Class<?>[]{String.class,boolean.class},"ChestedHorse",(scenario&8)!=0);
    Object changed=call(type,"bsx","a",new Class<?>[]{cls("dcw")},level);call(changed,"bsr","g",new Class<?>[]{cls("ub")},altered);Object saved=cls("ub").getConstructor().newInstance();call(changed,"bsr","e",new Class<?>[]{cls("ub")},saved);GeneratedEntityProbe.print("animal_changed",id,scenario,saved);GeneratedEntityProbe.metadata("animal_changed_metadata",id,scenario,changed,access);
   }
   System.out.printf("animal_type %s %s %s %s %s%n",id,call(entityRegistry,"jz","a",new Class<?>[]{Object.class},type),call(type,"bsx","l",new Class<?>[]{}),call(type,"bsx","m",new Class<?>[]{}),call(type,"bsx","c",new Class<?>[]{}));
  }
  Object empty=cls("dcl").getField("a").get(null),origin=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(0,0,0);
  for(int stateID=0;stateID<26684;stateID++) {Object state=call(null,"dfy","a",new Class<?>[]{int.class},stateID),fluid=call(state,"dtb$a","u",new Class<?>[]{});int ground=0,space=0;
   for(int i=0;i<types.size();i++) {Object type=types.get(i);if((boolean)call(state,"dtb$a","a",new Class<?>[]{cls("dcc"),cls("jd"),cls("bsx")},empty,origin,type))ground|=1<<i;if((boolean)call(null,"ddh","a",new Class<?>[]{cls("dcc"),cls("jd"),cls("dtc"),cls("epe"),cls("bsx")},empty,origin,state,fluid,type))space|=1<<i;}
   System.out.printf("animal_state %d %d %d%n",stateID,ground,space);
  }
  for(String biomeName:new String[]{"minecraft:plains","minecraft:birch_forest","minecraft:river"})for(int scenario=0;scenario<5;scenario++)for(int index=0;index<128;index++) {
   long seed=index*987654321L;final int mode=scenario;
   Object random=cls("dyz").getConstructor(long.class).newInstance(seed),worldRandom=cls("dzz").getConstructor(long.class).newInstance(seed^0x123456789abcdefL),biome=holders.get(biomeName);List<Object> entities=new ArrayList<>();
   Object spawnWorld=Proxy.newProxyInstance(cls("ddl").getClassLoader(),new Class<?>[]{cls("ddl")},(p,m,v)->{
    if(m.getName().equals("E_"))return worldRandom;if(m.getName().equals("E"))return level;if(m.getName().equals("C_"))return border;if(m.getName().equals("D_"))return dim;if(m.getName().equals("d_"))return difficulty;
    if(m.getName().equals("a_")&&v!=null&&v.length==1&&cls("bsr").isInstance(v[0])){entities.add(v[0]);return null;}
    if(m.getName().equals("a_")&&v!=null&&v.length==1&&cls("jd").isInstance(v[0])) {int y=(int)call(v[0],"kh","v",new Class<?>[]{}),x=(int)call(v[0],"kh","u",new Class<?>[]{});return y<64?grass:(mode==1&&x<-24?grass:air);}
    if(m.getName().equals("a")&&v!=null&&v.length==3&&cls("dyy$a").isInstance(v[0]))return 64;
    if(m.getName().equals("b")&&v!=null&&v.length==1&&cls("ewx").isInstance(v[0]))return mode!=2;
    if(m.getName().equals("b")&&v!=null&&v.length==2&&cls("jd").isInstance(v[0])&&v[1] instanceof Integer)return mode==3?0:15;
    if(m.getName().equals("f")&&v!=null&&v.length==1&&cls("bsr").isInstance(v[0]))return mode!=4;
    if(m.getName().equals("t")||m.getName().equals("getNoiseBiome"))return biome;
    if(m.getName().equals("B_"))return 0;
    if(m.isDefault())return InvocationHandler.invokeDefault(p,m,v);
    throw new UnsupportedOperationException(m.toString());
   });
   call(null,"ddh","a",new Class<?>[]{cls("ddl"),cls("jm"),cls("dcd"),cls("ayw")},spawnWorld,biome,cls("dcd").getConstructor(int.class,int.class).newInstance(-2,3),random);
   System.out.printf("animal_pack %s %d -2 3 %d %d %s %s%n",biomeName,seed,scenario,entities.size(),call(random,"dyz","g",new Class<?>[]{}),call(worldRandom,"dzz","g",new Class<?>[]{}));
   for(Object entity:entities){call(entity,"bsr","a_",new Class<?>[]{UUID.class},new UUID(0,0));Object tag=cls("ub").getConstructor().newInstance();call(entity,"bsr","e",new Class<?>[]{cls("ub")},tag);GeneratedEntityProbe.print("animal_nbt",biomeName,seed,tag);}
  }
 }
}
