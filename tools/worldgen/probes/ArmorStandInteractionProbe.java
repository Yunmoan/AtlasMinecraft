package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.util.*;

/** Original 1.21.1 interactAt, item-slot and serverbound codec vectors. */
public final class ArmorStandInteractionProbe {
 static Class<?> cls(String n)throws Exception{return Class.forName(n);}
 static Object call(Object o,String c,String n,Class<?>[] t,Object...v)throws Exception{Method m;try{m=cls(c).getMethod(n,t);}catch(NoSuchMethodException e){m=cls(c).getDeclaredMethod(n,t);m.setAccessible(true);}return m.invoke(o,v);}
 static Object item(int id,int count)throws Exception {
  if(count==0)return cls("cuq").getField("l").get(null);
  Object value=call(cls("lt").getField("g").get(null),"jz","a",new Class<?>[]{int.class},id);
  return cls("cuq").getConstructor(cls("dcv"),int.class).newInstance(value,count);
 }
 static int id(String name)throws Exception {
  Object r=cls("lt").getField("g").get(null),v=call(r,"jz","a",new Class<?>[]{cls("akr")},call(null,"akr","a",new Class<?>[]{String.class},"minecraft:"+name));
  return (int)call(r,"jz","a",new Class<?>[]{Object.class},v);
 }
 static String stack(Object s)throws Exception {
  if((boolean)call(s,"cuq","e",new Class<?>[]{}))return "0:0";
  return call(cls("lt").getField("g").get(null),"jz","a",new Class<?>[]{Object.class},call(s,"cuq","g",new Class<?>[]{}))+":"+call(s,"cuq","H",new Class<?>[]{});
 }
 public static void main(String[]args)throws Exception {
  InitialAnimalSpawnProbe.main(new String[]{"setup"});
  Object level=InitialAnimalSpawnProbe.probeLevel,registry=cls("lt").getField("f").get(null);
  Object access=cls("ka$c").getConstructor(List.class).newInstance(List.of(registry,cls("lt").getField("g").get(null)));
  class HelperLoader extends ClassLoader {
   HelperLoader(){super(ArmorStandInteractionProbe.class.getClassLoader());}
   Class<?> helper(String name)throws Exception {byte[] data=java.nio.file.Files.readAllBytes(java.nio.file.Path.of(System.getProperty("atlas.root"),"artifacts/worldgen-probes/"+name+".class"));return defineClass(name,data,0,data.length);}
  }
  HelperLoader helpers=new HelperLoader();
  InitialAnimalSpawnProbe.field(level,"aqu","O",helpers.helper("AtlasProbeGameEventDispatcher").getConstructor(cls("aqu")).newInstance(level));
  Object type=call(registry,"jz","a",new Class<?>[]{cls("akr")},call(null,"akr","a",new Class<?>[]{String.class},"minecraft:armor_stand"));
  Object stand=call(type,"bsx","a",new Class<?>[]{cls("dcw")},level);
  Object[] slots=(Object[])call(null,"bsy","values",new Class<?>[]{});
  Object hand=cls("bqq").getField("a").get(null);
  int itemCount=0;for(Object ignored:(Iterable<?>)cls("lt").getField("g").get(null))itemCount++;
  for(int i=1;i<itemCount;i++) {
   Object slot=call(stand,"btn","h",new Class<?>[]{cls("cuq")},item(i,1));
   System.out.printf("armor_item_slot %d %d%n",i,((Enum<?>)slot).ordinal());
  }
  Field u=sun.misc.Unsafe.class.getDeclaredField("theUnsafe");u.setAccessible(true);var unsafe=(sun.misc.Unsafe)u.get(null);
  Class<?> playerClass=helpers.helper("AtlasArmorStandPlayer");
  int[] equip={id("diamond_sword"),id("shield"),id("iron_boots"),id("iron_leggings"),id("iron_chestplate"),id("iron_helmet")};
  int[] heldIDs={0,equip[0],equip[1],equip[2],equip[3],equip[4],equip[5],id("elytra"),id("carved_pumpkin"),id("name_tag"),id("stone")};
  double[] heights={0,.099999,.1,.399999,.4,.549999,.55,.899999,.9,1.199999,1.2,1.399999,1.4,1.599999,1.6,1.899999,1.9};
  Random r=new Random(767);int cases=4096;
  for(int index=0;index<cases;index++) {
   stand=call(type,"bsx","a",new Class<?>[]{cls("dcw")},level);
   boolean small=r.nextBoolean(),arms=r.nextBoolean(),marker=index%37==0,creative=r.nextBoolean(),spectator=index%43==0;
   int mask=r.nextInt(64),disabled=index%5==0?0:1<<r.nextInt(24),heldID=heldIDs[r.nextInt(heldIDs.length)],count=heldID==0?0:r.nextBoolean()?1:5;
   double scale=new double[]{.5,1,2}[r.nextInt(3)],y=heights[r.nextInt(heights.length)]*scale*(small?.5:1);
   call(stand,"ciw","u",new Class<?>[]{boolean.class},small);call(stand,"ciw","a",new Class<?>[]{boolean.class},arms);call(stand,"ciw","v",new Class<?>[]{boolean.class},marker);
   InitialAnimalSpawnProbe.field(stand,"ciw","cg",disabled);
   Object attr=call(stand,"btn","a",new Class<?>[]{cls("jm")},cls("buw").getField("y").get(null));if(attr!=null)call(attr,"bus","a",new Class<?>[]{double.class},scale);
   scale=(float)call(stand,"btn","eb",new Class<?>[]{});
   for(int i=0;i<6;i++)call(stand,"ciw","a",new Class<?>[]{cls("bsy"),cls("cuq")},slots[i],item(equip[i],(mask&(1<<i))==0?0:1));
   Object player=unsafe.allocateInstance(playerClass);playerClass.getField("level").set(player,level);playerClass.getField("held").set(player,item(heldID,count));playerClass.getField("creative").setBoolean(player,creative);playerClass.getField("spectator").setBoolean(player,spectator);
   Object point=cls("exc").getConstructor(double.class,double.class,double.class).newInstance(0d,y,0d);
   Object result=call(stand,"ciw","a",new Class<?>[]{cls("cmx"),cls("exc"),cls("bqq")},player,point,hand);
   System.out.printf(Locale.ROOT,"armor_interact %d %b %b %b %d %.1f %d %d %b %b %.9f %d %s %s",index,small,arms,marker,disabled,scale,heldID,count,creative,spectator,y,mask,((Enum<?>)result).name(),stack(playerClass.getField("held").get(player)));
   for(int i=0;i<6;i++)System.out.print(" "+stack(call(stand,"ciw","a",new Class<?>[]{cls("bsy")},slots[i])));
   System.out.println();
  }
  call(stand,"bsr","e",new Class<?>[]{int.class},77);
  for(boolean secondary:new boolean[]{false,true})for(int action=0;action<3;action++)for(int h=0;h<2;h++) {
   Object currentHand=slots[0];currentHand=((Object[])call(null,"bqq","values",new Class<?>[]{}))[h];
   Object point=cls("exc").getConstructor(double.class,double.class,double.class).newInstance(.1d,1.2d,-.2d),packet;
   if(action==0)packet=call(null,"ahd","a",new Class<?>[]{cls("bsr"),boolean.class,cls("bqq")},stand,secondary,currentHand);
   else if(action==1)packet=call(null,"ahd","a",new Class<?>[]{cls("bsr"),boolean.class},stand,secondary);
   else packet=call(null,"ahd","a",new Class<?>[]{cls("bsr"),boolean.class,cls("bqq"),cls("exc")},stand,secondary,currentHand,point);
   System.out.printf("armor_interact_packet %d %d %b %s%n",action,h,secondary,GeneratedEntityPacketProbe.encode("ahd",packet,access));
  }
 }
}
