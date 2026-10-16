package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.util.*;
import java.util.jar.JarFile;

/** Actual template Entity.load/save and STRUCTURE Mob.finalizeSpawn for every
 * authored Outpost actor, with full saved NBT, metadata and region RNG. */
public final class OutpostActorProbe {
 static Class<?> cls(String name)throws Exception{return Class.forName(name);}
 static Object call(Object receiver,String owner,String name,Class<?>[] types,Object...v)throws Exception{return cls(owner).getMethod(name,types).invoke(receiver,v);}
 // AttributeInstanceMap iteration depends on identity hashes, outside the
 // world seed. Canonicalize only this unordered map in saved snapshots;
 // authored tags, modifier lists, inventory and all random draws stay intact.
 @SuppressWarnings("unchecked") static void printSaved(String label,String id,long seed,Object tag)throws Exception {
  List<Object> attributes=(List<Object>)call(tag,"ub","c",new Class<?>[]{String.class,int.class},"attributes",10);
  attributes.sort(Comparator.comparing(value->{try{return (String)call(value,"ub","l",new Class<?>[]{String.class},"id");}catch(Exception e){throw new RuntimeException(e);}}));
  GeneratedEntityProbe.print(label,id,seed,tag);
 }
 public static void main(String[]args)throws Exception {
  InitialAnimalSpawnProbe.main(new String[]{"setup"});
  Object level=InitialAnimalSpawnProbe.probeLevel,access=call(level,"dcw","H_",new Class<?>[]{}),registry=cls("lt").getField("f").get(null);
  InitialAnimalSpawnProbe.field(level,"dcw","H",cls("dcw").getField("i").get(null));
  Object difficulty=cls("bqp").getConstructor(cls("bqo"),long.class,long.class,float.class).newInstance(cls("bqo").getField("c").get(null),0L,0L,0f);
  Map<String,List<Object>> templates=new TreeMap<>();Set<String> ids=new TreeSet<>();
  try(JarFile jar=new JarFile(new File(cls("ent").getProtectionDomain().getCodeSource().getLocation().toURI()))) {
   var entries=jar.entries();while(entries.hasMoreElements()) {
    var entry=entries.nextElement();String name=entry.getName(),prefix="data/minecraft/structure/pillager_outpost/";
    if(!name.startsWith(prefix)||!name.endsWith(".nbt"))continue;
    Object root;try(InputStream input=jar.getInputStream(entry)){root=call(null,"uo","a",new Class<?>[]{InputStream.class,cls("uk")},input,call(null,"uk","a",new Class<?>[]{}));}
    Object entities=call(root,"ub","c",new Class<?>[]{String.class,int.class},"entities",10);List<Object> actors=new ArrayList<>();
    for(int i=0;i<(int)call(entities,"uh","size",new Class<?>[]{});i++) {
     Object actor=call(entities,"uh","a",new Class<?>[]{int.class},i),nbt=call(actor,"ub","p",new Class<?>[]{String.class},"nbt");actors.add(nbt);ids.add((String)call(nbt,"ub","l",new Class<?>[]{String.class},"id"));
    }
    if(!actors.isEmpty())templates.put("minecraft:pillager_outpost/"+name.substring(prefix.length(),name.length()-4),actors);
   }
  }
  for(String id:ids) {
   Object type=call(registry,"jz","a",new Class<?>[]{cls("akr")},call(null,"akr","a",new Class<?>[]{String.class},id));
   Object entity=call(type,"bsx","a",new Class<?>[]{cls("dcw")},level);call(entity,"bsr","a_",new Class<?>[]{UUID.class},new UUID(0,0));
   // LivingEntity's constructor yaw uses process-global Math.random,
   // outside all world-seeded streams. Canonicalize default yaw and UUID
   // only; authored template Rotation below remains untouched.
   call(entity,"bsr","t",new Class<?>[]{float.class},0f);
   Object saved=cls("ub").getConstructor().newInstance();call(entity,"bsr","e",new Class<?>[]{cls("ub")},saved);
   printSaved("outpost_actor_default",id,0,saved);GeneratedEntityProbe.metadata("outpost_actor_default_metadata",id,0,entity,access);
   System.out.printf("outpost_actor_type %s %s%n",id,call(registry,"jz","a",new Class<?>[]{Object.class},type));
  }
  for(var template:templates.entrySet())for(int actorIndex=0;actorIndex<template.getValue().size();actorIndex++) {
   Object original=template.getValue().get(actorIndex);String id=(String)call(original,"ub","l",new Class<?>[]{String.class},"id");
   GeneratedEntityProbe.print("outpost_actor_authored",template.getKey(),actorIndex,original);
   Object type=call(registry,"jz","a",new Class<?>[]{cls("akr")},call(null,"akr","a",new Class<?>[]{String.class},id));
   for(long seed:new long[]{0,1,-1,42,1234,987654321})for(int rotation=0;rotation<4;rotation++) {
    Object random=cls("dzz").getConstructor(long.class).newInstance(seed),biome=InitialAnimalSpawnProbe.probeBiomes.get("minecraft:plains");
    Object world=Proxy.newProxyInstance(cls("ddl").getClassLoader(),new Class<?>[]{cls("ddl")},(p,m,v)->{
     if(m.getName().equals("E_"))return random;if(m.getName().equals("E"))return level;if(m.getName().equals("H_"))return access;if(m.getName().equals("d_"))return difficulty;
     if(m.getName().equals("t")||m.getName().equals("getNoiseBiome"))return biome;if(m.isDefault())return InvocationHandler.invokeDefault(p,m,v);throw new UnsupportedOperationException(m.toString());
    });
    Object entity=call(type,"bsx","a",new Class<?>[]{cls("dcw")},level);call(entity,"bsr","g",new Class<?>[]{cls("ub")},call(original,"ub","i",new Class<?>[]{}));
    call(entity,"bsr","a_",new Class<?>[]{UUID.class},new UUID(0,0));
    float yaw=(float)call(entity,"bsr","a",new Class<?>[]{cls("dmm")},cls("dmm").getEnumConstants()[rotation]);
    call(entity,"bsr","b",new Class<?>[]{double.class,double.class,double.class,float.class,float.class},-15.5,70.0,16.5,yaw,0f);
    Object input=cls("ub").getConstructor().newInstance();call(entity,"bsr","e",new Class<?>[]{cls("ub")},input);
    System.out.printf("outpost_actor_case %s %d %d %d%n",template.getKey(),actorIndex,seed,rotation);printSaved("outpost_actor_input",id,seed,input);
    call(entity,"btp","a",new Class<?>[]{cls("ddl"),cls("bqp"),cls("btr"),cls("buh")},world,difficulty,cls("btr").getField("d").get(null),null);
    Object saved=cls("ub").getConstructor().newInstance();call(entity,"bsr","e",new Class<?>[]{cls("ub")},saved);printSaved("outpost_actor_finalized",id,seed,saved);GeneratedEntityProbe.metadata("outpost_actor_finalized_metadata",id,seed,entity,access);
    System.out.printf("outpost_actor_next %s %d %d %d %s%n",template.getKey(),actorIndex,seed,rotation,call(random,"dzz","g",new Class<?>[]{}));
   }
  }
  System.err.println("Outpost authored actor types="+ids+" actor templates="+templates.size());
 }
}
