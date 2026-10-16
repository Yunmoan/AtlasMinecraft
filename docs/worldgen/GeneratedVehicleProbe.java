package atlasprobe;

import java.io.*;
import java.util.*;

/** Real chest minecart Entity.save/load and passenger packet codecs. */
public final class GeneratedVehicleProbe {
 static Class<?> cls(String name)throws Exception{return Class.forName(name);}
 static Object call(Object receiver,String owner,String name,Class<?>[]types,Object...v)throws Exception{return cls(owner).getMethod(name,types).invoke(receiver,v);}
 public static void main(String[]args)throws Exception {
  InitialAnimalSpawnProbe.main(new String[]{"setup"});
  Object level=InitialAnimalSpawnProbe.probeLevel,registry=cls("lt").getField("f").get(null),access=cls("ka$c").getConstructor(List.class).newInstance(List.of(registry));
  Object type=call(registry,"jz","a",new Class<?>[]{cls("akr")},call(null,"akr","a",new Class<?>[]{String.class},"minecraft:chest_minecart"));
  System.out.printf("vehicle_type minecraft:chest_minecart %s%n",call(registry,"jz","a",new Class<?>[]{Object.class},type));
  for(int scenario=0;scenario<4;scenario++) {
   Object entity=call(type,"bsx","a",new Class<?>[]{cls("dcw")},level);call(entity,"bsr","a_",new Class<?>[]{UUID.class},new UUID(0,0));
   Object saved=cls("ub").getConstructor().newInstance();call(entity,"bsr","e",new Class<?>[]{cls("ub")},saved);
   if(scenario>0){call(saved,"ub","a",new Class<?>[]{String.class,String.class},"LootTable","minecraft:chests/abandoned_mineshaft");call(saved,"ub","a",new Class<?>[]{String.class,long.class},"LootTableSeed",scenario==1?1L:scenario==2?-1L:Long.MAX_VALUE);call(entity,"bsr","g",new Class<?>[]{cls("ub")},saved);saved=cls("ub").getConstructor().newInstance();call(entity,"bsr","e",new Class<?>[]{cls("ub")},saved);}
   GeneratedEntityProbe.print(scenario==0?"vehicle_default":"vehicle_saved","minecraft:chest_minecart",scenario,saved);
   GeneratedEntityProbe.metadata("vehicle_metadata","minecraft:chest_minecart",scenario,entity,access);
  }
  var unsafeField=sun.misc.Unsafe.class.getDeclaredField("theUnsafe");unsafeField.setAccessible(true);var unsafe=(sun.misc.Unsafe)unsafeField.get(null);
  for(int vehicle:new int[]{0,77,128,1234567})for(int[]passengers:new int[][]{new int[]{},new int[]{78},new int[]{78,129,1234568}}) {
   Object packet=unsafe.allocateInstance(cls("afg"));InitialAnimalSpawnProbe.field(packet,"afg","b",vehicle);InitialAnimalSpawnProbe.field(packet,"afg","c",passengers);
   String ids=passengers.length==0?"-":Arrays.toString(passengers).replace("[","").replace("]","").replace(" ","");
   System.out.printf("vehicle_passengers %d %s %s%n",vehicle,ids,GeneratedEntityPacketProbe.encode("afg",packet,access));
  }
 }
}
