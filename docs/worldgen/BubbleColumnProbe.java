package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.*;

/** Actual 1.21.1 BubbleColumnBlock.updateColumn and magma/soul-sand
 * updateShape, with the real block/fluid states and Level.setBlock's
 * unchanged-state return contract. No Go bubble rules are mirrored here. */
public final class BubbleColumnProbe {
 static Class<?> cls(String name)throws Exception{return Class.forName(name);}
 static Object call(Object receiver,String owner,String name,Class<?>[]types,Object...args)throws Exception{return cls(owner).getMethod(name,types).invoke(receiver,args);}
 static Object property(Object state,String field,Comparable<?> value)throws Exception {
  return call(state,"dte","a",new Class<?>[]{cls("duf"),Comparable.class},cls("dts").getField(field).get(null),value);
 }
 static String key(Object p)throws Exception{return call(p,"jd","u",new Class<?>[]{})+" "+call(p,"jd","v",new Class<?>[]{})+" "+call(p,"jd","w",new Class<?>[]{});}
 static Object position(int x,int y,int z)throws Exception{return cls("jd").getConstructor(int.class,int.class,int.class).newInstance(x,y,z);}
 public static void main(String[]args)throws Exception {
  Object version=call(null,"t","a",new Class<?>[]{});call(null,"ab","a",new Class<?>[]{cls("ae")},version);call(null,"akt","a",new Class<?>[]{});
  System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
  System.setErr(new PrintStream(new FileOutputStream(FileDescriptor.err),true));
  Object registry=cls("lt").getField("e").get(null);
  Map<String,Object> states=new HashMap<>();
  for(String name:new String[]{"air","water","magma_block","soul_sand","bubble_column","stone","kelp","kelp_plant","seagrass","tall_seagrass","oak_slab","lava"}) {
   Object block=call(registry,"jz","a",new Class<?>[]{cls("akr")},call(null,"akr","a",new Class<?>[]{String.class},"minecraft:"+name));
   states.put(name,call(block,"dfy","o",new Class<?>[]{}));
  }
  states.put("bubble_down",states.get("bubble_column"));
  states.put("bubble_up",call(states.get("bubble_column"),"dte","a",new Class<?>[]{cls("duf"),Comparable.class},cls("dge").getField("b").get(null),false));
  states.put("wet_slab",property(states.get("oak_slab"),"C",true));
  for(int level=0;level<16;level++)states.put("water_"+level,property(states.get("water"),"aP",level));
  Method update=cls("dge").getMethod("b",cls("dcx"),cls("jd"),cls("dtc"));
  int[][] origins={{8,40,8},{15,40,15},{16,40,16},{-1,40,-16},{8,319,8},{8,-63,8}};
  for(String support:new String[]{"magma_block","soul_sand","bubble_down","bubble_up","stone"})for(int mode=0;mode<32;mode++)for(int[]origin:origins) {
   int x=origin[0],y=origin[1],z=origin[2];
   Map<String,Object> blocks=new HashMap<>(),changed=new HashMap<>();
   blocks.put(x+" "+(y-1)+" "+z,states.get(support));
   int top=mode==31?320:Math.min(320,y+8);
   for(int height=y;height<top;height++) {
    blocks.put(x+" "+height+" "+z,states.get("water"));
    blocks.put((x+1)+" "+height+" "+z,states.get("water"));
   }
   if(mode<16)blocks.put(x+" "+y+" "+z,states.get("water_"+mode));
   if(mode==17)blocks.put(x+" "+(y+3)+" "+z,states.get("water_1"));
   if(mode==18)blocks.put(x+" "+y+" "+z,states.get("kelp"));
   if(mode>=19&&mode<=26)blocks.put(x+" "+(y+3)+" "+z,states.get(new String[]{"kelp_plant","seagrass","tall_seagrass","wet_slab","bubble_down","bubble_up","stone","lava"}[mode-19]));
   if(mode==27)blocks.put(x+" "+y+" "+z,states.get("air"));
   if(mode==28)blocks.put(x+" "+(y+3)+" "+z,states.get("air"));
   if(mode==29)blocks.put(x+" "+y+" "+z,states.get("bubble_up"));
   if(mode==30)for(int height=y;height<top;height++)blocks.put(x+" "+height+" "+z,states.get("bubble_down"));
   Object world=Proxy.newProxyInstance(cls("dcx").getClassLoader(),new Class<?>[]{cls("dcx")},(p,m,v)->{
    if(v!=null&&v.length>0&&cls("jd").isInstance(v[0])) {
     String k=key(v[0]);int height=(int)call(v[0],"jd","v",new Class<?>[]{});
     Object old=height < -64||height>=320?states.get("air"):blocks.getOrDefault(k,states.get("air"));
     if(m.getName().equals("a_"))return old;
     if(m.getName().equals("a")&&v.length==3&&cls("dtc").isInstance(v[1])) {
      if((int)v[2]!=2)throw new IllegalStateException("unexpected bubble write flags "+v[2]);
      if(height < -64||height>=320||old==v[1])return false;
      blocks.put(k,v[1]);changed.put(k,v[1]);return true;
     }
    }
    throw new UnsupportedOperationException(m.toString());
   });
   update.invoke(null,world,position(x,y,z),states.get(support));
   List<String> rows=new ArrayList<>();
   for(var entry:changed.entrySet())rows.add(entry.getKey()+" "+call(null,"dfy","i",new Class<?>[]{cls("dtc")},entry.getValue()));
   rows.sort(String::compareTo);
   String hash=HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(String.join("\n",rows).getBytes(StandardCharsets.UTF_8)));
   System.out.printf("bubble_column %s %d %d %d %d %d %s%n",support,mode,x,y,z,changed.size(),hash);
  }
  String[] directionNames={"down","up","north","south","west","east"};
  String[] directionFields={"a","b","c","d","e","f"};
  for(String support:new String[]{"magma_block","soul_sand"})for(int side=0;side<6;side++)for(String above:new String[]{"water","water_1","kelp","air"}) {
   List<Integer> delays=new ArrayList<>();
   Object world=Proxy.newProxyInstance(cls("dcx").getClassLoader(),new Class<?>[]{cls("dcx")},(p,m,v)->{
    if(m.getName().equals("a")&&v!=null&&v.length==3&&cls("dfy").isInstance(v[1])&&v[2] instanceof Integer) {
     delays.add((int)v[2]);return null;
    }
    throw new UnsupportedOperationException(m.toString());
   });
   Object state=states.get(support),block=call(state,"dtb$a","b",new Class<?>[]{}),direction=cls("ji").getField(directionFields[side]).get(null);
   Method shape=block.getClass().getDeclaredMethod("a",cls("dtc"),cls("ji"),cls("dtc"),cls("dcx"),cls("jd"),cls("jd"));shape.setAccessible(true);
   Object result=shape.invoke(block,state,direction,states.get(above),world,position(8,39,8),position(8,40,8));
   if(result!=state)throw new IllegalStateException("support state changed");
   System.out.printf("bubble_neighbor %s %s %s %d %d%n",support,directionNames[side],above,delays.size(),delays.isEmpty()?0:delays.getFirst());
  }
 }
}
