package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.util.*;
import java.util.stream.Stream;
import com.google.gson.*;
import com.mojang.serialization.Lifecycle;

/** Actual subtype finalization, pack data, variants and saved appearance. */
public final class OtherInitialAnimalProbe {
 static Class<?> cls(String name)throws Exception{return Class.forName(name);}
 static Object call(Object receiver,String owner,String name,Class<?>[] types,Object...v)throws Exception{return cls(owner).getMethod(name,types).invoke(receiver,v);}
 static List<String> tagMembers(String folder,String name)throws Exception {
  List<String> out=new ArrayList<>();String[] parts=name.split(":",2);String path="data/"+parts[0]+"/tags/"+folder+"/"+parts[1]+".json";
  try(var input=OtherInitialAnimalProbe.class.getClassLoader().getResourceAsStream(path)){if(input==null)return out;JsonArray values=JsonParser.parseReader(new InputStreamReader(input)).getAsJsonObject().getAsJsonArray("values");for(JsonElement value:values){String text=value.isJsonObject()?value.getAsJsonObject().get("id").getAsString():value.getAsString();if(text.startsWith("#"))out.addAll(tagMembers(folder,text.substring(1)));else out.add(text);}}
  return out;
 }
 static List<Object> registryOrder(Object key)throws Exception {
  var raw=io.netty.buffer.Unpooled.wrappedBuffer(java.nio.file.Files.readAllBytes(java.nio.file.Path.of("internal/registry/registries.bin")));Object container=cls("vw").getConstructor(io.netty.buffer.ByteBuf.class).newInstance(raw);int count=(int)call(container,"vw","l",new Class<?>[]{});List<Object> result=new ArrayList<>();
  for(int i=0;i<count;i++){int length=(int)call(container,"vw","l",new Class<?>[]{});byte[] payload=new byte[length];raw.readBytes(payload);var bytes=io.netty.buffer.Unpooled.wrappedBuffer(payload);Object buffer=cls("vw").getConstructor(io.netty.buffer.ByteBuf.class).newInstance(bytes),decoded=call(cls("abe").getField("a").get(null),"yx","decode",new Class<?>[]{Object.class},buffer);if(call(decoded,"abe","b",new Class<?>[]{}).equals(key))for(Object entry:(List<?>)call(decoded,"abe","e",new Class<?>[]{}))result.add(call(entry,"kd$a","a",new Class<?>[]{}));bytes.release();}
  raw.release();return result;
 }
 static void metadata(String label,String id,long seed,Object entity,Object access,Object frogs)throws Exception {
  if(id.equals("minecraft:frog")){Object tag=cls("ub").getConstructor().newInstance();call(entity,"bsr","e",new Class<?>[]{cls("ub")},tag);String variant=(String)call(tag,"ub","l",new Class<?>[]{String.class},"variant");Object holder=((Optional<?>)call(frogs,"jz","c",new Class<?>[]{cls("akr")},call(null,"akr","a",new Class<?>[]{String.class},variant))).orElseThrow();call(entity,"cgy","i",new Class<?>[]{cls("jm")},holder);}
  GeneratedEntityProbe.metadata(label,id,seed,entity,access);
 }
 public static void main(String[]args)throws Exception {
  InitialAnimalSpawnProbe.main(new String[]{"setup"});Object level=InitialAnimalSpawnProbe.probeLevel,lookup=InitialAnimalSpawnProbe.probeLookup;Map<String,Object> biomes=InitialAnimalSpawnProbe.probeBiomes;
  Object biomeLookup=call(lookup,"jn$a","b",new Class<?>[]{cls("akq")},cls("lu").getField("aF").get(null));Map<Object,List<Object>> holderTags=new IdentityHashMap<>();for(Object h:biomes.values())holderTags.put(h,new ArrayList<>());
  Method bindHolder=cls("jm$c").getDeclaredMethod("a",Collection.class);bindHolder.setAccessible(true);Method bindSet=cls("jq$c").getDeclaredMethod("b",List.class);bindSet.setAccessible(true);
  for(Field f:cls("awd").getFields())if(f.getType()==cls("awu")) {Object tag=f.get(null);String name=call(tag,"awu","b",new Class<?>[]{}).toString();List<Object> holders=new ArrayList<>();for(String id:tagMembers("worldgen/biome",name)){Object h=biomes.get(id);if(h!=null){holders.add(h);holderTags.get(h).add(tag);}}Optional<?> set=(Optional<?>)call(biomeLookup,"jn","a",new Class<?>[]{cls("awu")},tag);if(set.isPresent())bindSet.invoke(set.get(),holders);System.out.printf("animal_biome_tag %s %s%n",name,String.join(",",tagMembers("worldgen/biome",name)));}
  for(var entry:holderTags.entrySet())bindHolder.invoke(entry.getKey(),entry.getValue());
  Object blockRegistry=cls("lt").getField("e").get(null);Map<Object,List<Object>> blockTags=new LinkedHashMap<>();
  for(Field field:cls("awe").getFields())if(field.getType()==cls("awu")) {Object tag=field.get(null);List<Object> values=new ArrayList<>();for(String id:tagMembers("block",call(tag,"awu","b",new Class<?>[]{}).toString())){Object block=call(blockRegistry,"jz","a",new Class<?>[]{cls("akr")},call(null,"akr","a",new Class<?>[]{String.class},id));values.add(call(blockRegistry,"jz","e",new Class<?>[]{Object.class},block));}blockTags.put(tag,values);}
  call(blockRegistry,"jz","a",new Class<?>[]{Map.class},blockTags);
  Object wolfKey=cls("lu").getField("m").get(null),wolves=cls("ju").getConstructor(cls("akq"),Lifecycle.class).newInstance(wolfKey,Lifecycle.stable()),wolfLookup=call(lookup,"jn$a","b",new Class<?>[]{cls("akq")},wolfKey);
  for(Object resource:registryOrder(wolfKey)){Object key=call(null,"akq","a",new Class<?>[]{cls("akq"),cls("akr")},wolfKey,resource),h=call(wolfLookup,"jn","b",new Class<?>[]{cls("akq")},key);call(null,"jz","a",new Class<?>[]{cls("jz"),cls("akq"),Object.class},wolves,key,call(h,"jm","a",new Class<?>[]{}));}
  call(wolves,"jz","m",new Class<?>[]{});
  Object frogKey=cls("lu").getField("F").get(null),frogs=cls("lt").getField("aj").get(null);
  Map<Object,Object> overrides=Map.of(wolfKey,wolves,frogKey,frogs);
  Object original=call(level,"dcw","H_",new Class<?>[]{}),access=Proxy.newProxyInstance(cls("ka").getClassLoader(),new Class<?>[]{cls("ka")},(p,m,v)->{if(v!=null&&v.length==1&&overrides.containsKey(v[0])){if(m.getName().equals("d"))return overrides.get(v[0]);if(m.getName().equals("c"))return Optional.of(overrides.get(v[0]));}return m.invoke(original,v);});InitialAnimalSpawnProbe.field(level,"dcw","I",access);
  for(Object wolf:(Iterable<?>)wolves)System.out.printf("animal_registry minecraft:wolf_variant %s %s%n",call(wolves,"jz","b",new Class<?>[]{Object.class},wolf),call(wolves,"jz","a",new Class<?>[]{Object.class},wolf));
  for(Object frog:(Iterable<?>)frogs)System.out.printf("animal_registry minecraft:frog_variant %s %s%n",call(frogs,"jz","b",new Class<?>[]{Object.class},frog),call(frogs,"jz","a",new Class<?>[]{Object.class},frog));
  Object registry=cls("lt").getField("f").get(null);List<Object> types=new ArrayList<>();
  for(String id:new String[]{"minecraft:armadillo","minecraft:fox","minecraft:frog","minecraft:goat","minecraft:llama","minecraft:mooshroom","minecraft:panda","minecraft:parrot","minecraft:polar_bear","minecraft:rabbit","minecraft:turtle","minecraft:wolf"}) {
   Object type=call(registry,"jz","a",new Class<?>[]{cls("akr")},call(null,"akr","a",new Class<?>[]{String.class},id));
   types.add(type);
   Object value=call(type,"bsx","a",new Class<?>[]{cls("dcw")},level);call(value,"bsr","b",new Class<?>[]{double.class,double.class,double.class,float.class,float.class},-15.5,63.0,16.5,0f,0f);call(value,"bsr","a_",new Class<?>[]{UUID.class},new UUID(0,0));Object tag=cls("ub").getConstructor().newInstance();call(value,"bsr","e",new Class<?>[]{cls("ub")},tag);GeneratedEntityProbe.print("animal_default",id,0,tag);metadata("animal_default_metadata",id,0,value,access,frogs);
   System.out.printf("animal_type %s %s %s %s %s %s%n",id,call(registry,"jz","a",new Class<?>[]{Object.class},type),call(type,"bsx","l",new Class<?>[]{}),call(type,"bsx","m",new Class<?>[]{}),call(type,"bsx","c",new Class<?>[]{}),call(null,"buk","b",new Class<?>[]{cls("bsx")},type));
   for(String biomeName:new String[]{"minecraft:plains","minecraft:taiga","minecraft:snowy_taiga","minecraft:desert","minecraft:savanna_plateau","minecraft:swamp","minecraft:mangrove_swamp"})for(long seed:new long[]{0,1,-1,42,1234,987654321}) {
    Object random=cls("dzz").getConstructor(long.class).newInstance(seed),biome=biomes.get(biomeName);Object group=null;
    Object world=Proxy.newProxyInstance(cls("ddl").getClassLoader(),new Class<?>[]{cls("ddl")},(p,m,v)->{if(m.getName().equals("E_"))return random;if(m.getName().equals("E"))return level;if(m.getName().equals("t")||m.getName().equals("getNoiseBiome"))return biome;if(m.getName().equals("z_"))return 63;if(m.getName().equals("H_"))return access;if(m.getName().equals("d_"))return InitialAnimalSpawnProbe.probeDifficulty;if(m.getName().equals("a_")&&v!=null&&v.length==1&&cls("jd").isInstance(v[0]))return InitialAnimalSpawnProbe.probeGrass;if(m.isDefault())return InvocationHandler.invokeDefault(p,m,v);throw new UnsupportedOperationException(m.toString());});
    for(int member=0;member<4;member++) {Object animal=call(type,"bsx","a",new Class<?>[]{cls("dcw")},level);call(animal,"bsr","b",new Class<?>[]{double.class,double.class,double.class,float.class,float.class},-15.5,63.0,16.5,0f,0f);call(animal,"bsr","a_",new Class<?>[]{UUID.class},new UUID(0,0));group=call(animal,"btp","a",new Class<?>[]{cls("ddl"),cls("bqp"),cls("btr"),cls("buh")},world,InitialAnimalSpawnProbe.probeDifficulty,cls("btr").getField("b").get(null),group);Object saved=cls("ub").getConstructor().newInstance();call(animal,"bsr","e",new Class<?>[]{cls("ub")},saved);System.out.printf("animal_finalized_case %s %s %d %d%n",id,biomeName,seed,member);GeneratedEntityProbe.print("animal_finalized",id,seed,saved);metadata("animal_finalized_metadata",id,seed,animal,access,frogs);}
    System.out.printf("animal_finalized_next %s %s %d %s%n",id,biomeName,seed,call(random,"dzz","g",new Class<?>[]{}));
   }
  }
  Object empty=cls("dcl").getField("a").get(null),origin=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(0,0,0);
  for(int stateID=0;stateID<26684;stateID++) {Object state=call(null,"dfy","a",new Class<?>[]{int.class},stateID),fluid=call(state,"dtb$a","u",new Class<?>[]{});int ground=0,space=0;
   for(int i=0;i<types.size();i++) {Object type=types.get(i);if((boolean)call(state,"dtb$a","a",new Class<?>[]{cls("dcc"),cls("jd"),cls("bsx")},empty,origin,type))ground|=1<<i;if((boolean)call(null,"ddh","a",new Class<?>[]{cls("dcc"),cls("jd"),cls("dtc"),cls("epe"),cls("bsx")},empty,origin,state,fluid,type))space|=1<<i;}
   System.out.printf("animal_state %d %d %d%n",stateID,ground,space);
   System.out.printf("animal_path %d %s%n",stateID,call(state,"dtb$a","a",new Class<?>[]{cls("ept")},cls("ept").getField("a").get(null)));
  }
  packs(level,biomes,access);
 }
 static Object state(String name)throws Exception {
  Object block=call(cls("lt").getField("e").get(null),"jz","a",new Class<?>[]{cls("akr")},call(null,"akr","a",new Class<?>[]{String.class},name));
  return call(block,"dfy","o",new Class<?>[]{});
 }
 static void packs(Object level,Map<String,Object> biomes,Object access)throws Exception {
  String[][] contexts={{"savanna","grass_block"},{"snowy_taiga","grass_block"},{"swamp","grass_block"},{"mangrove_swamp","mud"},{"snowy_slopes","stone"},{"savanna_plateau","grass_block"},{"mushroom_fields","mycelium"},{"bamboo_jungle","grass_block"},{"jungle","grass_block"},{"snowy_plains","grass_block"},{"beach","sand"},{"desert","sand"},{"frozen_ocean","ice"},{"old_growth_pine_taiga","grass_block"},{"old_growth_spruce_taiga","grass_block"},{"forest","grass_block"},{"badlands","red_sand"}};
  for(String[] context:contexts)for(int mode=0;mode<5;mode++)for(int index=0;index<256;index++) {
   String biomeName="minecraft:"+context[0],groundName="minecraft:"+context[1];long seed=index*987654321L;final int scenario=mode;
   Object placement=cls("dyz").getConstructor(long.class).newInstance(seed),random=cls("dzz").getConstructor(long.class).newInstance(seed^0x123456789abcdefL),biome=biomes.get(biomeName),ground=state(groundName);List<Object> entities=new ArrayList<>();
   Object world=Proxy.newProxyInstance(cls("ddl").getClassLoader(),new Class<?>[]{cls("ddl")},(p,m,v)->{
    if(m.getName().equals("E_"))return random;if(m.getName().equals("E"))return level;if(m.getName().equals("C_"))return InitialAnimalSpawnProbe.probeBorder;if(m.getName().equals("D_"))return InitialAnimalSpawnProbe.probeDimension;if(m.getName().equals("d_"))return InitialAnimalSpawnProbe.probeDifficulty;if(m.getName().equals("H_"))return access;
    if(m.getName().equals("a_")&&v!=null&&v.length==1&&cls("bsr").isInstance(v[0])){entities.add(v[0]);return null;}
    if(m.getName().equals("a_")&&v!=null&&v.length==1&&cls("jd").isInstance(v[0])){int y=(int)call(v[0],"kh","v",new Class<?>[]{}),x=(int)call(v[0],"kh","u",new Class<?>[]{});return y<64?ground:scenario==1&&x<-24?ground:InitialAnimalSpawnProbe.probeAir;}
    if(m.getName().equals("b_")&&v!=null&&v.length==1&&cls("jd").isInstance(v[0]))return call(InitialAnimalSpawnProbe.probeAir,"dtb$a","u",new Class<?>[]{});
    if(m.getName().equals("a")&&v!=null&&v.length==3&&cls("dyy$a").isInstance(v[0]))return 64;
    if(m.getName().equals("b")&&v!=null&&v.length==1&&cls("ewx").isInstance(v[0]))return scenario!=2;
    if(m.getName().equals("b")&&v!=null&&v.length==2&&cls("jd").isInstance(v[0])&&v[1] instanceof Integer)return scenario==3?0:15;
    if(m.getName().equals("f")&&v!=null&&v.length==1&&cls("bsr").isInstance(v[0]))return scenario!=4;
    if(m.getName().equals("t")||m.getName().equals("getNoiseBiome"))return biome;
    if(m.getName().equals("B_"))return 0;if(m.getName().equals("z_"))return 63;
    if(m.isDefault())return InvocationHandler.invokeDefault(p,m,v);throw new UnsupportedOperationException(m.toString());
   });
   call(null,"ddh","a",new Class<?>[]{cls("ddl"),cls("jm"),cls("dcd"),cls("ayw")},world,biome,cls("dcd").getConstructor(int.class,int.class).newInstance(-2,3),placement);
   System.out.printf("other_animal_pack %s %s %d -2 3 %d %d %s %s%n",biomeName,groundName,seed,scenario,entities.size(),call(placement,"dyz","g",new Class<?>[]{}),call(random,"dzz","g",new Class<?>[]{}));
   for(Object entity:entities){call(entity,"bsr","a_",new Class<?>[]{UUID.class},new UUID(0,0));Object tag=cls("ub").getConstructor().newInstance();call(entity,"bsr","e",new Class<?>[]{cls("ub")},tag);GeneratedEntityProbe.print("other_animal_nbt",biomeName,seed,tag);}
  }
 }
}
