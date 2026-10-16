package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.util.*;
import java.util.jar.JarFile;

/** Independent audit and original loading/finalization oracle for every
 * authored actor in the five village template families and common animals. */
public final class VillageActorProbe {
 static final java.nio.file.Path ROOT=java.nio.file.Path.of(System.getProperty("atlas.root",".")).toAbsolutePath();
 static Class<?> cls(String name)throws Exception{return Class.forName(name);}
 static Object call(Object receiver,String owner,String name,Class<?>[] types,Object...v)throws Exception{return cls(owner).getMethod(name,types).invoke(receiver,v);}
 // AttributeInstanceMap iteration uses identity hashes and differs by JVM
 // process. Canonicalize this unordered registry map only; preserve all item,
 // pose, inventory and other list orders in the actual saved NBT.
 @SuppressWarnings("unchecked") static void canonicalAttributes(Object tag)throws Exception {
  List<Object> attributes=(List<Object>)call(tag,"ub","c",new Class<?>[]{String.class,int.class},"attributes",10);
  attributes.sort(Comparator.comparing(value->{try{return (String)call(value,"ub","l",new Class<?>[]{String.class},"id");}catch(Exception e){throw new RuntimeException(e);}}));
  for(Object passenger:(List<Object>)call(tag,"ub","c",new Class<?>[]{String.class,int.class},"Passengers",10))canonicalAttributes(passenger);
 }
 static void print(String label,String id,long seed,Object tag)throws Exception {
  canonicalAttributes(tag);
  // Chicken's constructor uses its independently created process-global RNG
  // for EggLayTime. Keep that constructor stage outside the controlled Mob
  // finalization stream; the runtime supplies its own deterministic seed.
  if(id.equals("minecraft:chicken"))call(tag,"ub","r",new Class<?>[]{String.class},"EggLayTime");
  GeneratedEntityProbe.print(label,id,seed,tag);
 }
 // Audit actual saved/reloaded state separately from fresh finalization.
 // Camel's inherited Tame flag and horse Health differ after Entity.load.
 static void loadedMetadata(String label,String id,long seed,Object tag,Object level,Object access,Object registry)throws Exception {
  Object type=call(registry,"jz","a",new Class<?>[]{cls("akr")},call(null,"akr","a",new Class<?>[]{String.class},id));
  Object loaded=call(type,"bsx","a",new Class<?>[]{cls("dcw")},level);
  call(loaded,"bsr","g",new Class<?>[]{cls("ub")},tag);
  GeneratedEntityProbe.metadata(label,id,seed,loaded,access);
 }
 static Map<String,List<Object>> templates()throws Exception {
  Map<String,List<Object>> result=new TreeMap<>();
  try(JarFile jar=new JarFile(ROOT.resolve("vanilla/server-inner.jar").toFile())) {
   var entries=jar.entries();while(entries.hasMoreElements()) {
    var entry=entries.nextElement();String name=entry.getName(),prefix="data/minecraft/structure/village/";
    if(!name.startsWith(prefix)||!name.endsWith(".nbt"))continue;
    Object root;try(InputStream input=jar.getInputStream(entry)){root=call(null,"uo","a",new Class<?>[]{InputStream.class,cls("uk")},input,call(null,"uk","a",new Class<?>[]{}));}
    Object entities=call(root,"ub","c",new Class<?>[]{String.class,int.class},"entities",10);List<Object> actors=new ArrayList<>();
    for(int i=0;i<(int)call(entities,"uh","size",new Class<?>[]{});i++) {
     Object actor=call(entities,"uh","a",new Class<?>[]{int.class},i),nbt=call(actor,"ub","p",new Class<?>[]{String.class},"nbt");actors.add(nbt);
    }
    if(!actors.isEmpty())result.put("minecraft:village/"+name.substring(prefix.length(),name.length()-4),actors);
   }
  }
  return result;
 }
 public static void main(String[]args)throws Exception {
  Map<String,List<Object>> templates=templates();Map<String,Integer> counts=new TreeMap<>(),families=new TreeMap<>();
  for(var template:templates.entrySet())for(int index=0;index<template.getValue().size();index++) {
   Object actor=template.getValue().get(index);String id=(String)call(actor,"ub","l",new Class<?>[]{String.class},"id");
   counts.merge(id,1,Integer::sum);String family=template.getKey().split("/")[1];families.merge(family,1,Integer::sum);
   print("village_actor_authored",template.getKey(),index,actor);
   System.err.printf("village_actor_audit %s %d %s%n",template.getKey(),index,actor);
  }
  System.err.printf("village_actor_summary templates=%d types=%s families=%s%n",templates.size(),counts,families);
  if(args.length>0&&args[0].equals("audit"))return;
  InitialAnimalSpawnProbe.main(new String[]{"setup"});
  Object level=InitialAnimalSpawnProbe.probeLevel,registry=cls("lt").getField("f").get(null);
  Object access=cls("ka$c").getConstructor(List.class).newInstance(List.of(cls("lt").getField("ai").get(null),cls("lt").getField("w").get(null),cls("lt").getField("x").get(null),cls("lt").getField("g").get(null),registry));
  InitialAnimalSpawnProbe.field(level,"dcw","H",cls("dcw").getField("i").get(null));
  long[] gameTime={0L};
  Object levelData=Proxy.newProxyInstance(cls("erm").getClassLoader(),new Class<?>[]{cls("erm")},(p,m,v)->{
   if(m.getName().equals("q"))return cls("bqo").getField("c").get(null);
   if(m.getName().equals("c"))return gameTime[0];
   if(m.getReturnType()==long.class)return 0L;if(m.getReturnType()==boolean.class)return false;if(m.getReturnType()==int.class)return 0;
   throw new UnsupportedOperationException(m.toString());
  });
  InitialAnimalSpawnProbe.field(level,"dcw","A",levelData);
  class HelperLoader extends ClassLoader {
   HelperLoader(){super(VillageActorProbe.class.getClassLoader());}
   Class<?> helper(String name)throws Exception {byte[] bytes=java.nio.file.Files.readAllBytes(ROOT.resolve("artifacts/worldgen-probes/"+name+".class"));return defineClass(name,bytes,0,bytes.length);}
  }
  InitialAnimalSpawnProbe.field(level,"aqu","ad",new HelperLoader().helper("AtlasEntityStructureManager").getConstructor(cls("ejz")).newInstance(cls("ejz").getField("b").get(null)));
  InitialAnimalSpawnProbe.field(level,"aqu","O",new HelperLoader().helper("AtlasProbeGameEventDispatcher").getConstructor(cls("aqu")).newInstance(level));
  Object catRegistry=cls("lt").getField("ai").get(null);Map<Object,List<Object>> catTags=new HashMap<>();
  for(String name:new String[]{"default_spawns","full_moon_spawns"}) {
   String text;try(var input=VillageActorProbe.class.getClassLoader().getResourceAsStream("data/minecraft/tags/cat_variant/"+name+".json")){text=new String(input.readAllBytes(),java.nio.charset.StandardCharsets.UTF_8);}
   var matches=java.util.regex.Pattern.compile("\"minecraft:[^\"]+\"").matcher(text);List<Object> values=new ArrayList<>();
   if(name.equals("full_moon_spawns"))values.addAll(catTags.get(cls("awf").getField("a").get(null)));
   while(matches.find()) {String resource=matches.group().replace("\"","");if(resource.equals("minecraft:default_spawns"))continue;Object variant=call(catRegistry,"jz","a",new Class<?>[]{cls("akr")},call(null,"akr","a",new Class<?>[]{String.class},resource));values.add(call(catRegistry,"jz","e",new Class<?>[]{Object.class},variant));}
   catTags.put(cls("awf").getField(name.equals("default_spawns")?"a":"b").get(null),values);
  }
  call(catRegistry,"jz","a",new Class<?>[]{Map.class},catTags);
  for(var entry:new TreeMap<>(InitialAnimalSpawnProbe.probeBiomes).entrySet())System.out.printf("village_actor_biome %s minecraft:%s%n",entry.getKey(),call(null,"cmp","a",new Class<?>[]{cls("jm")},entry.getValue()));
  Set<String> defaultTypes=new TreeSet<>(counts.keySet());defaultTypes.add("minecraft:chicken");
  for(String id:defaultTypes) {
   Object type=call(registry,"jz","a",new Class<?>[]{cls("akr")},call(null,"akr","a",new Class<?>[]{String.class},id));
   Object entity=call(type,"bsx","a",new Class<?>[]{cls("dcw")},level);call(entity,"bsr","a_",new Class<?>[]{UUID.class},new UUID(0,0));
   // Constructor yaw uses an independent process-global seed. Give the
   // constructor snapshot a controlled origin/rotation, just as template
   // Entity.load and moveTo do for every real actor below.
   call(entity,"bsr","b",new Class<?>[]{double.class,double.class,double.class,float.class,float.class},0d,0d,0d,0f,0f);
   if(id.equals("minecraft:villager")||id.equals("minecraft:zombie_villager")) {
    Object villageType=call(cls("lt").getField("w").get(null),"jz","a",new Class<?>[]{cls("akr")},call(null,"akr","a",new Class<?>[]{String.class},"minecraft:plains"));
    Object profession=call(cls("lt").getField("x").get(null),"jz","a",new Class<?>[]{cls("akr")},call(null,"akr","a",new Class<?>[]{String.class},"minecraft:none"));
    Object data=cls("cml").getConstructor(cls("cmp"),cls("cmn"),int.class).newInstance(villageType,profession,1);
    call(entity,id.equals("minecraft:villager")?"cmk":"ckx","a",new Class<?>[]{cls("cml")},data);
   }
   Object saved=cls("ub").getConstructor().newInstance();call(entity,"bsr","e",new Class<?>[]{cls("ub")},saved);
   print(id.equals("minecraft:chicken")?"village_actor_support_default":"village_actor_default",id,0,saved);
   GeneratedEntityProbe.metadata("village_actor_default_metadata",id,0,entity,access);
   loadedMetadata("village_actor_loaded_metadata",id,0,saved,level,access,registry);
   System.out.printf("village_actor_type %s %s%n",id,call(registry,"jz","a",new Class<?>[]{Object.class},type));
  }
  for(var template:templates.entrySet())for(int actorIndex=0;actorIndex<template.getValue().size();actorIndex++) {
   Object original=template.getValue().get(actorIndex);String id=(String)call(original,"ub","l",new Class<?>[]{String.class},"id");
   Object type=call(registry,"jz","a",new Class<?>[]{cls("akr")},call(null,"akr","a",new Class<?>[]{String.class},id));
   String family=template.getKey().split("/")[1];
   List<String> biomes=family.equals("common")?List.of("minecraft:plains","minecraft:desert","minecraft:savanna","minecraft:snowy_plains","minecraft:taiga"):List.of("minecraft:"+(family.equals("snowy")?"snowy_plains":family));
   List<Long> seeds=new ArrayList<>(List.of(0L,1L,-1L,42L,1234L,987654321L));
   if(template.getKey().equals("minecraft:village/desert/zombie/villagers/unemployed"))for(int i=1;i<=4096;i++)seeds.add(i*987654321L);
   for(String biomeName:biomes)for(int seedIndex=0;seedIndex<seeds.size();seedIndex++)for(int rotation=0;rotation<(seedIndex<6?4:1);rotation++) {
    long seed=seeds.get(seedIndex);
    long actorSeed=seed^0x123456789abcdefL;boolean fullMoon=(seed&1)==0;Object random=cls("dzz").getConstructor(long.class).newInstance(seed),actorRandom=cls("dyz").getConstructor(long.class).newInstance(actorSeed),biome=InitialAnimalSpawnProbe.probeBiomes.get(biomeName);List<Object> extras=new ArrayList<>();
    Object world=Proxy.newProxyInstance(cls("ddl").getClassLoader(),new Class<?>[]{cls("ddl")},(p,m,v)->{
     if(m.getName().equals("E_"))return random;if(m.getName().equals("E"))return level;if(m.getName().equals("H_"))return access;if(m.getName().equals("d_"))return InitialAnimalSpawnProbe.probeDifficulty;
     if(m.getName().equals("t")||m.getName().equals("getNoiseBiome"))return biome;
     if(m.getName().equals("aq"))return fullMoon?1f:.5f;
     if(m.getName().equals("a")&&v!=null&&v.length==3&&v[0] instanceof Class)return List.of();
     if(m.getName().equals("b")&&v!=null&&v.length==1&&cls("bsr").isInstance(v[0])){extras.add(v[0]);return true;}
     if(m.getName().equals("a_")&&v!=null&&v.length==1&&cls("jd").isInstance(v[0]))return InitialAnimalSpawnProbe.probeAir;
     if(m.isDefault())return InvocationHandler.invokeDefault(p,m,v);throw new UnsupportedOperationException(m.toString());
    });
    Object entity=call(type,"bsx","a",new Class<?>[]{cls("dcw")},level);InitialAnimalSpawnProbe.field(entity,"bsr","ah",actorRandom);call(entity,"bsr","g",new Class<?>[]{cls("ub")},call(original,"ub","i",new Class<?>[]{}));
    call(entity,"bsr","a_",new Class<?>[]{UUID.class},new UUID(0,0));
    float yaw=(float)call(entity,"bsr","a",new Class<?>[]{cls("dmm")},cls("dmm").getEnumConstants()[rotation]);
    call(entity,"bsr","b",new Class<?>[]{double.class,double.class,double.class,float.class,float.class},-15.5,70.0,16.5,yaw,0f);
    Object input=cls("ub").getConstructor().newInstance();call(entity,"bsr","e",new Class<?>[]{cls("ub")},input);
    System.out.printf("village_actor_case %s %d %s %d %d %d %d%n",template.getKey(),actorIndex,biomeName,seed,actorSeed,rotation,fullMoon?1:0);print("village_actor_input",id,seed,input);
    if(cls("btp").isInstance(entity))call(entity,"btp","a",new Class<?>[]{cls("ddl"),cls("bqp"),cls("btr"),cls("buh")},world,InitialAnimalSpawnProbe.probeDifficulty,cls("btr").getField("d").get(null),null);
    Object saved=cls("ub").getConstructor().newInstance();call(entity,"bsr","d",new Class<?>[]{cls("ub")},saved);print("village_actor_finalized",id,seed,saved);GeneratedEntityProbe.metadata("village_actor_metadata",id,seed,entity,access);
    System.out.printf("village_actor_next %s %s %s %d%n",id,call(random,"dzz","g",new Class<?>[]{}),call(actorRandom,"dyz","g",new Class<?>[]{}),extras.size());
    for(Object extra:extras) {
     call(extra,"bsr","a_",new Class<?>[]{UUID.class},new UUID(0,0));
     Object extraSaved=cls("ub").getConstructor().newInstance();call(extra,"bsr","e",new Class<?>[]{cls("ub")},extraSaved);print("village_actor_extra","minecraft:chicken",seed,extraSaved);
    }
    if(id.equals("minecraft:armor_stand")||id.equals("minecraft:zombie_villager")) {
     List<Object> equipment=new ArrayList<>();boolean populated=false;
     for(Object slot:(Object[])call(null,"bsy","values",new Class<?>[]{})) {Object stack=call(entity,"btn","a",new Class<?>[]{cls("bsy")},slot);populated|=!(boolean)call(stack,"cuq","e",new Class<?>[]{});equipment.add(com.mojang.datafixers.util.Pair.of(slot,stack));}
     if(populated)System.out.printf("village_actor_equipment %s %d %s%n",id,seed,GeneratedEntityPacketProbe.encode("afc",cls("afc").getConstructor(int.class,List.class).newInstance(77,equipment),access));
    }
    loadedMetadata("village_actor_loaded_metadata",id,seed,saved,level,access,registry);
   }
  }
  Object camelType=call(registry,"jz","a",new Class<?>[]{cls("akr")},call(null,"akr","a",new Class<?>[]{String.class},"minecraft:camel"));
  for(long time:new long[]{Long.MIN_VALUE,-24000,-1,0,1,52,53,54,6000,24000,Long.MAX_VALUE}) {
   gameTime[0]=time;Object random=cls("dzz").getConstructor(long.class).newInstance(42L);
   Object world=Proxy.newProxyInstance(cls("ddl").getClassLoader(),new Class<?>[]{cls("ddl")},(p,m,v)->{if(m.getName().equals("E_"))return random;if(m.getName().equals("E"))return level;throw new UnsupportedOperationException(m.toString());});
   Object camel=call(camelType,"bsx","a",new Class<?>[]{cls("dcw")},level);call(camel,"bsr","g",new Class<?>[]{cls("ub")},templates.get("minecraft:village/desert/camel_spawn").get(0));call(camel,"bsr","a_",new Class<?>[]{UUID.class},new UUID(0,0));
   call(camel,"btp","a",new Class<?>[]{cls("ddl"),cls("bqp"),cls("btr"),cls("buh")},world,InitialAnimalSpawnProbe.probeDifficulty,cls("btr").getField("d").get(null),null);
   Object saved=cls("ub").getConstructor().newInstance();call(camel,"bsr","e",new Class<?>[]{cls("ub")},saved);print("village_actor_camel_time","minecraft:camel",time,saved);GeneratedEntityProbe.metadata("village_actor_camel_time_metadata","minecraft:camel",time,camel,access);
   loadedMetadata("village_actor_loaded_metadata","minecraft:camel",time,saved,level,access,registry);
  }
 }
}
