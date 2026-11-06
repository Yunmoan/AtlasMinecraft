package atlasprobe;

import com.google.gson.*;
import com.mojang.serialization.*;
import io.netty.buffer.*;
import java.io.*;
import java.lang.reflect.*;
import java.nio.charset.StandardCharsets;
import java.nio.file.*;
import java.util.*;
import java.util.function.Function;

/** Development oracle: actual LootTable.fill and protocol-767 menu codecs. */
public final class ChestLootProbe {
 static Class<?> cls(String n)throws Exception{return Class.forName(n);}
 static void field(Object o,String type,String n,Object v)throws Exception{Field f=cls(type).getDeclaredField(n);f.setAccessible(true);f.set(o,v);}
 static Object call(Object o,String c,String n,Class<?>[]types,Object...args)throws Exception{return cls(c).getMethod(n,types).invoke(o,args);}
 static Object location(String n)throws Exception{return call(null,"akr","a",new Class<?>[]{String.class},n);}
 static String encoded(Object tag)throws Exception{ByteArrayOutputStream b=new ByteArrayOutputStream();call(null,"uo","a",new Class<?>[]{cls("uy"),java.io.DataOutput.class},tag,new DataOutputStream(b));return Base64.getEncoder().encodeToString(b.toByteArray());}
 static String menuSnapshot(Object menu,Object inventory,Object access)throws Exception {
  Object saved=cls("uh").getConstructor().newInstance();
  for(int slot=0;slot<65;slot++) {
   Object stack=slot<63?call(call(menu,"cpu","b",new Class<?>[]{int.class},slot),"crq","g",new Class<?>[]{}):slot==63?call(menu,"cpu","g",new Class<?>[]{}):call(inventory,"cmw","a",new Class<?>[]{int.class},40);
   if((Boolean)call(stack,"cuq","e",new Class<?>[]{}))continue;
   Object tag=call(stack,"cuq","b",new Class<?>[]{cls("jo$a")},access);call(tag,"ub","a",new Class<?>[]{String.class,byte.class},"Slot",(byte)slot);call(saved,"uh","add",new Class<?>[]{Object.class},tag);
  }
  return encoded(saved);
 }
 static void menuFixtures(Object level,Object access,sun.misc.Unsafe unsafe,String directory)throws Exception {
  class HelperLoader extends ClassLoader {HelperLoader(){super(ChestLootProbe.class.getClassLoader());}Class<?> helper()throws Exception{byte[] b=Files.readAllBytes(Path.of(directory,"AtlasChestMenuPlayer.class"));return defineClass("AtlasChestMenuPlayer",b,0,b.length);}}
  Class<?> helper=new HelperLoader().helper();
  Object itemRegistry=cls("lt").getField("g").get(null);
  Object[] ids=new Object[3];int[] limits={64,16,1};int i=0;
  for(String name:new String[]{"stone","snowball","diamond_sword"})ids[i++]=call(itemRegistry,"jz","a",new Class<?>[]{cls("akr")},location("minecraft:"+name));
  Random random=new Random(7671211);int vectors=1024;
  for(int scenario=0;scenario<vectors;scenario++) {
   Object player=unsafe.allocateInstance(helper);helper.getField("level").set(player,level);boolean creative=scenario%3!=0;helper.getField("creative").setBoolean(player,creative);
   Object inventory=cls("cmw").getConstructor(cls("cmx")).newInstance(player);helper.getField("inventory").set(player,inventory);
   Object chest=cls("bra").getConstructor(int.class).newInstance(27),menu=call(null,"cqc","a",new Class<?>[]{int.class,cls("cmw"),cls("bqk")},37,inventory,chest);
   for(int slot=0;slot<65;slot++) {
    if(random.nextInt(4)==0)continue;int type=random.nextInt(ids.length),count=1+random.nextInt(limits[type]);if(slot%7==0)count=limits[type];
    Object stack=cls("cuq").getConstructor(cls("dcv"),int.class).newInstance(ids[type],count);
    if(slot<63)call(call(menu,"cpu","b",new Class<?>[]{int.class},slot),"crq","e",new Class<?>[]{cls("cuq")},stack);
    else if(slot==63)call(menu,"cpu","b",new Class<?>[]{cls("cuq")},stack);
    else call(inventory,"cmw","a",new Class<?>[]{int.class,cls("cuq")},40,stack);
   }
   List<int[]> actions=new ArrayList<>();int mode=new int[]{0,1,2,3,5,6}[scenario%6],slot=random.nextInt(63),button=random.nextInt(2);
   if(mode==2)button=scenario%7==0?40:random.nextInt(9);
   if(mode==3)button=2;
   if(mode==5) {
    int kind=scenario/6%3;actions.add(new int[]{-999,kind<<2,5});int count=1+scenario/18%5;
    for(int add=0;add<count;add++)actions.add(new int[]{random.nextInt(63),(kind<<2)|1,5});
    actions.add(new int[]{-999,(kind<<2)|2,5});
   } else actions.add(new int[]{slot,button,mode});
   String before=menuSnapshot(menu,inventory,access);StringJoiner encodedActions=new StringJoiner(",");
   for(int[] action:actions){encodedActions.add(action[0]+":"+action[1]+":"+action[2]);call(menu,"cpu","a",new Class<?>[]{int.class,int.class,cls("cqe"),cls("cmx")},action[0],action[1],cls("cqe").getEnumConstants()[action[2]],player);}
   System.out.printf("chest_menu %d %b %s %s %s%n",scenario,creative,encodedActions,before,menuSnapshot(menu,inventory,access));
  }
  System.err.printf("original ChestMenu.clicked: %d action vectors%n",vectors);
 }
 static void bindTags(Object access)throws Exception {
  var factory=cls("ats").getDeclaredMethod("b");factory.setAccessible(true);
  Object resources=cls("atx").getConstructor(cls("ass"),List.class).newInstance(cls("ass").getField("b").get(null),List.of(factory.invoke(null)));
  for(String[] spec:new String[][]{{"aL","enchantment"},{"I","instrument"}}) {
   Object key=cls("lu").getField(spec[0]).get(null),registry=call(access,"ka","d",new Class<?>[]{cls("akq")},key);
   Function<Object,Optional<?>> lookup=x->{try{return(Optional<?>)call(registry,"jz","c",new Class<?>[]{cls("akr")},x);}catch(Exception e){throw new RuntimeException(e);}};
   Object loader=cls("awv").getConstructor(Function.class,String.class).newInstance(lookup,"tags/"+spec[1]);
   Map<?,?> loaded=(Map<?,?>)call(loader,"awv","b",new Class<?>[]{cls("aue")},resources);
   Map<Object,Object> tags=new HashMap<>();
   for(var entry:loaded.entrySet())tags.put(call(null,"awu","a",new Class<?>[]{cls("akq"),cls("akr")},key,entry.getKey()),new ArrayList<>((Collection<?>)entry.getValue()));
   call(registry,"jz","a",new Class<?>[]{Map.class},tags);
  }
 }
 public static void main(String[]args)throws Exception {
  Object version=call(null,"t","a",new Class<?>[]{});call(null,"ab","a",new Class<?>[]{cls("ae")},version);call(null,"akt","a",new Class<?>[]{});
  Object access=IglooProbeSupport.registryAccess();bindTags(access);
  var uf=sun.misc.Unsafe.class.getDeclaredField("theUnsafe");uf.setAccessible(true);var unsafe=(sun.misc.Unsafe)uf.get(null);
  Object server=unsafe.allocateInstance(cls("apn")),level=unsafe.allocateInstance(cls("aqu")),resources=unsafe.allocateInstance(cls("alc"));
  Object frozen=call(access,"ka","d",new Class<?>[]{});
  field(resources,"alc","c",cls("alb$b").getConstructor(cls("ka$b")).newInstance(frozen));
  var holderConstructor=cls("net.minecraft.server.MinecraftServer$a").getDeclaredConstructor(cls("atu"),cls("alc"));holderConstructor.setAccessible(true);Object holder=holderConstructor.newInstance(null,resources);field(server,"net.minecraft.server.MinecraftServer","az",holder);
  Object worldData=Proxy.newProxyInstance(cls("erl").getClassLoader(),new Class<?>[]{cls("erl")},(p,m,v)->{if(m.getName().equals("K"))return cls("cpn").getField("g").get(null);throw new UnsupportedOperationException(m.toString());});field(server,"net.minecraft.server.MinecraftServer","j",worldData);
  field(level,"aqu","J",server);field(level,"dcw","I",access);field(level,"dcw","H",cls("dcw").getField("i").get(null));
  Object ops=call(null,"akp","a",new Class<?>[]{DynamicOps.class,cls("jo$a")},JsonOps.INSTANCE,access);
  System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
  System.setErr(new PrintStream(new FileOutputStream(FileDescriptor.err),true));
  Object instruments=call(access,"ka","d",new Class<?>[]{cls("akq")},cls("lu").getField("I").get(null));
  TreeMap<Integer,String> instrumentNames=new TreeMap<>();
  for(Object instrument:(Iterable<?>)instruments)instrumentNames.put((Integer)call(instruments,"jz","a",new Class<?>[]{Object.class},instrument),(String)call(instruments,"jz","b",new Class<?>[]{Object.class},instrument).toString());
  Files.writeString(Path.of(args[2]),new Gson().toJson(instrumentNames.values()),StandardCharsets.UTF_8);
  Object blockEntities=cls("lt").getField("j").get(null);
  Object trapped=call(blockEntities,"jz","a",new Class<?>[]{cls("akr")},location("minecraft:trapped_chest"));
  if(trapped==null)throw new AssertionError("missing original trapped chest block entity type");
  Object trappedEntity=call(trapped,"dqj","a",new Class<?>[]{cls("jd"),cls("dtc")},cls("jd").getConstructor(int.class,int.class,int.class).newInstance(0,0,0),RuinedPortalProbeSupport.block("trapped_chest"));
  if(!call(blockEntities,"jz","b",new Class<?>[]{Object.class},call(trappedEntity,"dqh","r",new Class<?>[]{})).toString().equals("minecraft:trapped_chest"))throw new AssertionError("wrong trapped chest saved type");
  Object blockRegistry=cls("lt").getField("e").get(null),air=RuinedPortalProbeSupport.block("air"),position=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(0,0,0);
  Object getter=Proxy.newProxyInstance(cls("dcc").getClassLoader(),new Class<?>[]{cls("dcc")},(p,m,v)->{if(m.getName().equals("a_"))return air;if(m.getName().equals("c_"))return null;if(m.getName().equals("b_"))return call(air,"dtb$a","u",new Class<?>[]{});if(m.getName().equals("I_"))return -64;if(m.getName().equals("an"))return 384;if(m.isDefault())return InvocationHandler.invokeDefault(p,m,v);throw new UnsupportedOperationException(m.toString());});
  BitSet conductor=new BitSet(26684);int states=0;
  for(Object state:(Iterable<?>)cls("dfy").getField("q").get(null)) {
   int id=(Integer)call(null,"dfy","i",new Class<?>[]{cls("dtc")},state);states++;
   if((Boolean)call(state,"dtb$a","g",new Class<?>[]{cls("dcc"),cls("jd")},getter,position))conductor.set(id);
  }
  byte[] bits=Arrays.copyOf(conductor.toByteArray(),(states+7)/8);Files.write(Path.of(args[0]),bits);
  Object component=call(null,"wz","c",new Class<?>[]{String.class},"container.chest");
  Object open=cls("adr").getConstructor(int.class,cls("crc"),cls("wz")).newInstance(37,cls("crc").getField("c").get(null),component);
  ByteBuf raw=Unpooled.buffer();Object buffer=cls("wk").getConstructor(ByteBuf.class,cls("ka")).newInstance(raw,access);
  call(cls("adr").getField("a").get(null),"yz","encode",new Class<?>[]{Object.class,Object.class},buffer,open);byte[] packet=new byte[raw.readableBytes()];raw.readBytes(packet);raw.release();System.out.println("open_screen "+HexFormat.of().formatHex(packet));
  Object stoneItem=call(cls("lt").getField("g").get(null),"jz","a",new Class<?>[]{cls("akr")},location("minecraft:stone"));
  Object offhand=cls("acr").getConstructor(int.class,int.class,int.class,cls("cuq")).newInstance(-2,17,40,cls("cuq").getConstructor(cls("dcv"),int.class).newInstance(stoneItem,6));
  raw=Unpooled.buffer();buffer=cls("wk").getConstructor(ByteBuf.class,cls("ka")).newInstance(raw,access);
  call(cls("acr").getField("a").get(null),"yz","encode",new Class<?>[]{Object.class,Object.class},buffer,offhand);packet=new byte[raw.readableBytes()];raw.readBytes(packet);raw.release();System.out.println("offhand_slot "+HexFormat.of().formatHex(packet));
  menuFixtures(level,access,unsafe,args[3]);
  List<String> tables=new ArrayList<>();
  try(var jar=new java.util.jar.JarFile(Path.of(args[1]).toFile())) {var e=jar.entries();while(e.hasMoreElements()){String n=e.nextElement().getName();if(n.startsWith("data/minecraft/loot_table/chests/village/")||n.equals("data/minecraft/loot_table/chests/pillager_outpost.json"))tables.add(n);}}
  Collections.sort(tables);
  for(String path:tables) {
   String text;try(var input=ChestLootProbe.class.getClassLoader().getResourceAsStream(path)){text=new String(input.readAllBytes(),StandardCharsets.UTF_8);}
   Object table=((Codec<?>)cls("erw").getField("d").get(null)).parse((DynamicOps)ops,JsonParser.parseString(text)).getOrThrow();
   for(int scenario=0;scenario<64;scenario++) {
    long seed=scenario<4?new long[]{1,-1,1234,Long.MIN_VALUE}[scenario]:0x123456789abcdefL+scenario*341873128712L;
    float luck=new float[]{0f,-2f,5f}[scenario%3];
    Object origin=cls("exc").getConstructor(double.class,double.class,double.class).newInstance(0d,64d,0d);
    Object params=cls("eru").getConstructor(cls("aqu"),Map.class,Map.class,float.class).newInstance(level,Map.of(cls("eun").getField("f").get(null),origin),Map.of(),luck);
    Object container=cls("bra").getConstructor(int.class).newInstance(27);
    call(table,"erw","a",new Class<?>[]{cls("bqk"),cls("eru"),long.class},container,params,seed);
    Object saved=cls("uh").getConstructor().newInstance();
    for(int slot=0;slot<27;slot++) {
     Object stack=call(container,"bra","a",new Class<?>[]{int.class},slot);if((Boolean)call(stack,"cuq","e",new Class<?>[]{}))continue;
     Object tag=call(stack,"cuq","b",new Class<?>[]{cls("jo$a")},access);call(tag,"ub","a",new Class<?>[]{String.class,byte.class},"Slot",(byte)slot);call(saved,"uh","add",new Class<?>[]{Object.class},tag);
    }
    System.out.printf("chest_loot minecraft:%s %d %s %s%n",path.substring("data/minecraft/loot_table/".length(),path.length()-5),seed,Float.toString(luck),encoded(saved));
   }
  }
  System.err.printf("original LootTable.fill: %d tables, %d vectors; %d redstone-conductor state checks%n",tables.size(),tables.size()*64,states);
 }
}
