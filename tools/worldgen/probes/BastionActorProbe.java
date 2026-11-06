package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.util.*;

/** Loads original authored actor NBT into actual mobs, applies STRUCTURE
 * finalization, then captures full saved NBT and the independent region RNG. */
public final class BastionActorProbe {
 static Class<?> cls(String name)throws Exception{return Class.forName(name);}
 static Object call(Object receiver,String owner,String name,Class<?>[] types,Object...v)throws Exception{return cls(owner).getMethod(name,types).invoke(receiver,v);}
 public static void main(String[]args)throws Exception {
  InitialAnimalSpawnProbe.main(new String[]{"setup"});
  Object level=InitialAnimalSpawnProbe.probeLevel,access=call(level,"dcw","H_",new Class<?>[]{}),registry=cls("lt").getField("f").get(null);
  InitialAnimalSpawnProbe.field(level,"dcw","H",cls("dcw").getField("i").get(null));
  Object randomHolder=cls("bqp").getConstructor(cls("bqo"),long.class,long.class,float.class).newInstance(cls("bqo").getField("c").get(null),0L,0L,0f);
  for(String id:new String[]{"minecraft:piglin","minecraft:piglin_brute","minecraft:hoglin"}) {
   Object type=call(registry,"jz","a",new Class<?>[]{cls("akr")},call(null,"akr","a",new Class<?>[]{String.class},id));
   Object entity=call(type,"bsx","a",new Class<?>[]{cls("dcw")},level);call(entity,"bsr","a_",new Class<?>[]{UUID.class},new UUID(0,0));
   Object saved=cls("ub").getConstructor().newInstance();call(entity,"bsr","e",new Class<?>[]{cls("ub")},saved);
   GeneratedEntityProbe.print("bastion_actor_default",id,0,saved);GeneratedEntityProbe.metadata("bastion_actor_default_metadata",id,0,entity,access);
   System.out.printf("bastion_actor_type %s %s%n",id,call(registry,"jz","a",new Class<?>[]{Object.class},type));
  }
  for(String template:new String[]{"sword_piglin","crossbow_piglin","melee_piglin","melee_piglin_always","hoglin"}) {
   Object original;
   try(InputStream input=BastionActorProbe.class.getResourceAsStream("/data/minecraft/structure/bastion/mobs/"+template+".nbt")) {
    Object root=call(null,"uo","a",new Class<?>[]{InputStream.class,cls("uk")},input,call(null,"uk","a",new Class<?>[]{}));
    Object entities=call(root,"ub","c",new Class<?>[]{String.class,int.class},"entities",10),first=call(entities,"uh","a",new Class<?>[]{int.class},0);original=call(first,"ub","p",new Class<?>[]{String.class},"nbt");
   }
   String id=(String)call(original,"ub","l",new Class<?>[]{String.class},"id");Object type=call(registry,"jz","a",new Class<?>[]{cls("akr")},call(null,"akr","a",new Class<?>[]{String.class},id));
   for(long seed:new long[]{0,1,-1,42,1234,987654321})for(int rotation=0;rotation<4;rotation++) {
    Object random=cls("dzz").getConstructor(long.class).newInstance(seed),biome=InitialAnimalSpawnProbe.probeBiomes.get("minecraft:nether_wastes");
    Object world=Proxy.newProxyInstance(cls("ddl").getClassLoader(),new Class<?>[]{cls("ddl")},(p,m,v)->{
     if(m.getName().equals("E_"))return random;if(m.getName().equals("E"))return level;if(m.getName().equals("H_"))return access;if(m.getName().equals("d_"))return randomHolder;
     if(m.getName().equals("t")||m.getName().equals("getNoiseBiome"))return biome;if(m.isDefault())return InvocationHandler.invokeDefault(p,m,v);throw new UnsupportedOperationException(m.toString());
    });
    Object entity=call(type,"bsx","a",new Class<?>[]{cls("dcw")},level);call(entity,"bsr","g",new Class<?>[]{cls("ub")},call(original,"ub","i",new Class<?>[]{}));
    call(entity,"bsr","a_",new Class<?>[]{UUID.class},new UUID(0,0));
    float yaw=(float)call(entity,"bsr","a",new Class<?>[]{cls("dmm")},cls("dmm").getEnumConstants()[rotation]);
    call(entity,"bsr","b",new Class<?>[]{double.class,double.class,double.class,float.class,float.class},-15.5,70.0,16.5,yaw,0f);
    Object input=cls("ub").getConstructor().newInstance();call(entity,"bsr","e",new Class<?>[]{cls("ub")},input);
    System.out.printf("bastion_actor_case %s %d %d%n",template,seed,rotation);GeneratedEntityProbe.print("bastion_actor_input",id,seed,input);
    call(entity,"btp","a",new Class<?>[]{cls("ddl"),cls("bqp"),cls("btr"),cls("buh")},world,randomHolder,cls("btr").getField("d").get(null),null);
    Object saved=cls("ub").getConstructor().newInstance();call(entity,"bsr","e",new Class<?>[]{cls("ub")},saved);GeneratedEntityProbe.print("bastion_actor_finalized",id,seed,saved);GeneratedEntityProbe.metadata("bastion_actor_finalized_metadata",id,seed,entity,access);
    System.out.printf("bastion_actor_next %s %d %d %s%n",template,seed,rotation,call(random,"dzz","g",new Class<?>[]{}));
   }
  }
 }
}
