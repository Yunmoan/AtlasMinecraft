package atlasprobe;
import java.io.*;import java.nio.file.*;import java.util.*;
public final class GeneratedEntityEquipmentProbe {
 static Class<?> cls(String name)throws Exception{return Class.forName(name);}
 static Object call(Object receiver,String owner,String name,Class<?>[] types,Object...v)throws Exception{return cls(owner).getMethod(name,types).invoke(receiver,v);}
 public static void main(String[]args)throws Exception {
  InitialAnimalSpawnProbe.main(new String[]{"setup"});Object level=InitialAnimalSpawnProbe.probeLevel,registry=cls("lt").getField("f").get(null);Object access=cls("ka$c").getConstructor(List.class).newInstance(List.of(registry,cls("lt").getField("g").get(null)));InitialAnimalSpawnProbe.field(level,"dcw","I",access);int count=0;
  for(String fixture:new String[]{"other_initial_animals_1.21.1.txt","structure_bastion_actors_1.21.1.txt","structure_drowned_actors_1.21.1.txt"})for(String line:Files.readAllLines(Path.of("docs/worldgen/"+fixture))) {
   String[] f=line.split(" ");if(f.length!=4||!(f[0].equals("animal_finalized")&&f[1].equals("minecraft:fox")||f[0].equals("bastion_actor_finalized")||f[0].equals("drowned_actor_finalized")))continue;
   var bytes=io.netty.buffer.Unpooled.wrappedBuffer(HexFormat.of().parseHex(f[3]));Object buffer=cls("vw").getConstructor(io.netty.buffer.ByteBuf.class).newInstance(bytes),tag=call(buffer,"vw","o",new Class<?>[]{});bytes.release();
   Object type=call(registry,"jz","a",new Class<?>[]{cls("akr")},call(null,"akr","a",new Class<?>[]{String.class},f[1])),entity=call(type,"bsx","a",new Class<?>[]{cls("dcw")},level);call(entity,"bsr","g",new Class<?>[]{cls("ub")},tag);
   List<Object> pairs=new ArrayList<>();boolean populated=false;
   for(Object slot:(Object[])call(null,"bsy","values",new Class<?>[]{})) {Object stack=call(entity,"btn","a",new Class<?>[]{cls("bsy")},slot);populated|=!(boolean)call(stack,"cuq","e",new Class<?>[]{});pairs.add(com.mojang.datafixers.util.Pair.of(slot,stack));}
   if(!populated)continue;
   GeneratedEntityProbe.print("actor_equipment_nbt",f[1],count,tag);
   Object packet=cls("afc").getConstructor(int.class,List.class).newInstance(77,pairs);
   System.out.printf("actor_equipment %s %d %s%n",f[1],count,GeneratedEntityPacketProbe.encode("afc",packet,access));count++;
  }
 }
}
