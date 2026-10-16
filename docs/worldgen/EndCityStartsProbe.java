package atlasprobe;

import java.io.*;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.*;
import java.util.function.Predicate;
import java.util.jar.JarFile;
import java.util.stream.Stream;

/** Actual EndCityPieces recursion and EndCityStructure terrain/biome gate. */
public final class EndCityStartsProbe {
 static Class<?> cls(String name)throws Exception{return Class.forName(name);}
 static String hash(String text)throws Exception{return HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(text.getBytes(StandardCharsets.UTF_8)));}
 public static void main(String[]args)throws Exception {
  Object version=cls("t").getMethod("a").invoke(null);cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);
  Object registries=IglooProbeSupport.registryAccess();RuinedPortalStartsProbe.bindBiomeTags(registries);List<String> names=new ArrayList<>();
  try(JarFile jar=new JarFile(new File(cls("ent").getProtectionDomain().getCodeSource().getLocation().toURI()))) {
   var entries=jar.entries();while(entries.hasMoreElements()){String n=entries.nextElement().getName();if(n.startsWith("data/minecraft/structure/end_city/")&&n.endsWith(".nbt"))names.add(n.substring("data/minecraft/structure/".length(),n.length()-4));}
  }
  Object manager=IglooProbeSupport.manager(names.toArray(String[]::new)),serialization=cls("ekh").getConstructor(cls("aue"),cls("ka"),cls("enu")).newInstance(null,registries,manager);
  System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
  System.setErr(new PrintStream(new FileOutputStream(FileDescriptor.err),true));
  long[] seeds={0,1,-1,1234,Long.MIN_VALUE,Long.MAX_VALUE,9212937586943976949L};
  for(long seed:seeds)for(int rotation=0;rotation<4;rotation++)for(int[]origin:new int[][]{{0,64,0},{15,100,-16},{-400,61,257},{-30,200,-31}}) {
   Object random=cls("dyz").getConstructor(long.class).newInstance(seed);List<Object> pieces=new ArrayList<>();Object position=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(origin[0],origin[1],origin[2]);
   cls("elo").getMethod("a",cls("enu"),cls("jd"),cls("dmm"),List.class,cls("ayw")).invoke(null,manager,position,cls("dmm").getEnumConstants()[rotation],pieces,random);
   StringBuilder saved=new StringBuilder();for(Object piece:pieces)saved.append(IglooProbeSupport.canonicalNBT(cls("ejv").getMethod("a",cls("ekh")).invoke(piece,serialization))).append('\n');
   System.out.printf("end_city_layout %d %d %d %d %d %d %s %s%n",seed,rotation,origin[0],origin[1],origin[2],pieces.size(),hash(saved.toString()),cls("ayw").getMethod("g").invoke(random));
   if(seed==0&&rotation==0&&origin[0]==0)System.err.println("end_city_pieces "+saved);
  }
  Object lookup=cls("jo$a").getMethod("b").invoke(registries);var registry=cls("jn$a").getMethod("b",cls("akq"));
  Object structures=registry.invoke(lookup,cls("lu").getField("aR").get(null)),biomes=registry.invoke(lookup,cls("lu").getField("aF").get(null)),settingsLookup=registry.invoke(lookup,cls("lu").getField("aO").get(null));
  Object structure=null;Map<String,Object> biomeHolders=new HashMap<>();
  for(Object holder:((Stream<?>)cls("jo").getMethod("b").invoke(structures)).toList())if(cls("akq").getMethod("a").invoke(cls("jm$c").getMethod("h").invoke(holder)).toString().equals("minecraft:end_city"))structure=cls("jm").getMethod("a").invoke(holder);
  for(Object holder:((Stream<?>)cls("jo").getMethod("b").invoke(biomes)).toList())biomeHolders.put(cls("akq").getMethod("a").invoke(cls("jm$c").getMethod("h").invoke(holder)).toString(),holder);
  Object key=cls("dzd").getField("g").get(null),settings=cls("jn").getMethod("b",cls("akq")).invoke(settingsLookup,key),heights=cls("dcy").getMethod("e",int.class,int.class).invoke(null,0,256);
  for(long seed:seeds) {
   Object noise=cls("dzm").getMethod("a",cls("jn$a"),cls("akq"),long.class).invoke(null,lookup,key,seed);
   for(String name:new String[]{"minecraft:end_highlands","minecraft:the_void"}) {
    Object source=cls("deh").getConstructor(cls("jm")).newInstance(biomeHolders.get(name)),generator=cls("dzb").getConstructor(cls("dea"),cls("jm")).newInstance(source,settings);
    for(int[]chunk:new int[][]{{0,0},{1,-1},{-1,1},{17,-9},{-41,23},{31,31},{-33,32},{145,-62},{-9,-15},{19,5},{-63,-27},{72,13},{5,6},{-8,-5},{101,-31},{-121,23}}) {
     Object chunkPos=cls("dcd").getConstructor(int.class,int.class).newInstance(chunk[0],chunk[1]),allowed=cls("ejr").getMethod("a").invoke(structure);
     Predicate<Object> predicate=holder->{try{return(Boolean)cls("jq").getMethod("a",cls("jm")).invoke(allowed,holder);}catch(Exception e){throw new IllegalStateException(e);}};
     Object context=cls("ejr$a").getConstructor(cls("ka"),cls("duz"),cls("dea"),cls("dzm"),cls("enu"),long.class,cls("dcd"),cls("dcy"),Predicate.class).newInstance(registries,generator,source,noise,manager,seed,chunkPos,heights,predicate);
     Optional<?> stub=(Optional<?>)cls("ejr").getMethod("b",cls("ejr$a")).invoke(structure,context);int count=0;String digest="-";
     if(stub.isPresent()) {
      Object builder=cls("ejr$b").getMethod("a").invoke(stub.get()),container=cls("ekj").getMethod("a").invoke(builder);count=((List<?>)cls("ekg").getMethod("c").invoke(container)).size();
      Object start=cls("ejz").getConstructor(cls("ejr"),cls("dcd"),int.class,cls("ekg")).newInstance(structure,chunkPos,0,container),saved=cls("ejz").getMethod("a",cls("ekh"),cls("dcd")).invoke(start,serialization,chunkPos);digest=hash(IglooProbeSupport.canonicalNBT(saved));
     }
     Object random=cls("ejr$a").getMethod("f").invoke(context);
     System.out.printf("end_city_start %d %d %d %s %d %s %s%n",seed,chunk[0],chunk[1],name,count,digest,cls("ayw").getMethod("g").invoke(random));
    }
   }
  }
 }
}
