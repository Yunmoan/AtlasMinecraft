package atlasprobe;

import com.mojang.serialization.Lifecycle;
import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.lang.reflect.Proxy;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.HexFormat;
import java.util.List;
import java.util.Map;
import java.util.TreeSet;
import java.util.stream.Stream;

/** Original SwampHutPiece.postProcess against defined terrain, all four
 * orientations and chunk clipping. Witch/Cat are already true in the probe
 * input so only the still-independent block stage is measured.
 */
public final class StructureHutBlocksProbe {
    static Class<?> cls(String n)throws Exception{return Class.forName(n);}
    static String digest(List<String> values)throws Exception {
        values.sort(String::compareTo);
        return HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(String.join("\n",values).getBytes(StandardCharsets.UTF_8)));
    }
    public static void main(String[] args)throws Exception { run(false); }
    public static void run(boolean pyramid)throws Exception { run(pyramid?1:0); }
    public static void run(int type)throws Exception {
        boolean pyramid=type==1,jungle=type==2;
        Object version=cls("t").getMethod("a").invoke(null);
        cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);
        Object registries=cls("pd").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Object lookup=cls("jo$a").getMethod("b").invoke(registries);
        Object pyramidStructure=null;
        if(pyramid) {
            Object structureLookup=cls("jn$a").getMethod("b",cls("akq")).invoke(lookup,cls("lu").getField("aR").get(null));
            for(Object holder:((Stream<?>)cls("jo").getMethod("b").invoke(structureLookup)).toList()) {
                Object key=cls("jm$c").getMethod("h").invoke(holder);
                if(cls("akq").getMethod("a").invoke(key).toString().equals("minecraft:desert_pyramid"))pyramidStructure=cls("jm").getMethod("a").invoke(holder);
            }
        }
        Object biomeLookup=cls("jn$a").getMethod("b",cls("akq")).invoke(lookup,cls("lu").getField("aF").get(null));
        Object biomeRegistry=cls("ju").getConstructor(cls("akq"),Lifecycle.class).newInstance(cls("lu").getField("aF").get(null),Lifecycle.stable());
        for(Object holder:((Stream<?>)cls("jo").getMethod("b").invoke(biomeLookup)).toList())
            cls("jz").getMethod("a",cls("jz"),cls("akq"),Object.class).invoke(null,biomeRegistry,cls("jm$c").getMethod("h").invoke(holder),cls("jm").getMethod("a").invoke(holder));
        Object heightAccessor=cls("dcy").getMethod("e",int.class,int.class).invoke(null,-64,384);
        Object blocksRegistry=cls("lt").getField("e").get(null);
        Map<String,Object> blocks=new HashMap<>();
        for(String name:new String[]{"air","stone","grass_block","water","oak_leaves"}) {
            Object block=cls("jz").getMethod("a",cls("akr")).invoke(blocksRegistry,cls("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name));
            blocks.put(name,cls("dfy").getMethod("o").invoke(block));
        }
        java.lang.reflect.Method getX=cls("jd").getMethod("u"),getY=cls("jd").getMethod("v"),getZ=cls("jd").getMethod("w");
        java.lang.reflect.Method fluid=cls("dtb$a").getMethod("u"),solid=cls("dtb$a").getMethod("d"),emptyFluid=cls("epe").getMethod("c");
        var stateID=cls("dfy").getMethod("i",cls("dtc"));
        for(long worldSeed:pyramid||jungle?new long[]{0,1,-1,1234}:new long[]{0})for(int orientation=0;orientation<4;orientation++)for(int scenario=0;scenario<(pyramid?4:5);scenario++)for(int[] anchor:new int[][]{{0,0},{15,15},{-16,-16}}) {
            int width=jungle?12:pyramid?21:7,depth=jungle?15:pyramid?21:9;
            int xSize=orientation%2==0?width:depth,zSize=orientation%2==0?depth:width,pieceHeight=jungle?10:pyramid?15:7;
            int[] box={anchor[0],64,anchor[1],anchor[0]+xSize-1,64+pieceHeight-1,anchor[1]+zSize-1};
            for(int dx=0;dx<=1;dx++)for(int dz=0;dz<=1;dz++) {
                int chunkX=(anchor[0]>>4)+dx,chunkZ=(anchor[1]>>4)+dz;
                Object tag=cls("ub").getConstructor().newInstance();
                cls("ub").getMethod("a",String.class,String.class).invoke(tag,"id",jungle?"minecraft:tejp":pyramid?"minecraft:tedp":"minecraft:tesh");
                cls("ub").getMethod("a",String.class,int[].class).invoke(tag,"BB",box);
                for(var entry:Map.of("O",orientation,"GD",0,"Width",width,"Height",pieceHeight,"Depth",depth,"HPos",-1).entrySet())
                    cls("ub").getMethod("a",String.class,int.class).invoke(tag,entry.getKey(),entry.getValue());
                if(!pyramid&&!jungle)for(String field:new String[]{"Witch","Cat"})cls("ub").getMethod("a",String.class,boolean.class).invoke(tag,field,true);
                Object piece=cls(jungle?"elt":pyramid?"elm":"eml").getConstructor(cls("ub")).newInstance(tag);
                Object worldRandom=cls("dyz").getConstructor(long.class).newInstance(worldSeed^0x123456789abcdefL);
                Map<String,Object> changed=new HashMap<>(),chunks=new HashMap<>(),entities=new HashMap<>();
                final int mode=scenario;
                java.util.function.Function<int[],Object> read=p->{
                    String key=p[0]+" "+p[1]+" "+p[2];if(changed.containsKey(key))return changed.get(key);
                    int floor=mode==1?60+Math.floorMod(p[0]+2*p[2],5):mode==2||mode==3?60:mode==4?-65:63;
                    if(p[1]<=floor)return blocks.get(p[1]==floor?"grass_block":"stone");
                    if(mode==2&&p[1]<=66)return blocks.get("water");
                    if(mode==3&&p[1]==72)return blocks.get("oak_leaves");
                    return blocks.get("air");
                };
                Object world=Proxy.newProxyInstance(cls("dds").getClassLoader(),new Class<?>[]{cls("dds")},(proxy,method,values)->{
                    String name=method.getName();
                    if(name.equals("I_"))return -64;if(name.equals("an"))return 384;if(name.equals("am"))return 320;
                    if(name.equals("C"))return worldSeed;if(name.equals("E_"))return worldRandom;
                    if(values!=null&&values.length==2&&cls("dyy$a").isInstance(values[0])&&cls("jd").isInstance(values[1])) {
                        int x=(Integer)getX.invoke(values[1]),z=(Integer)getZ.invoke(values[1]);
                        if(changed.isEmpty()) {
                            int floor=mode==1?60+Math.floorMod(x+2*z,5):mode==2||mode==3?60:mode==4?-65:63;
                            int height=mode==2?67:Math.max(-64,floor+1);
                            return cls("jd").getConstructor(int.class,int.class,int.class).newInstance(x,height,z);
                        }
                        for(int y=319;y>=-64;y--) {
                            Object s=read.apply(new int[]{x,y,z});
                            if(s==blocks.get("oak_leaves"))continue;
                            if((Boolean)solid.invoke(s)||!(Boolean)emptyFluid.invoke(fluid.invoke(s)))return cls("jd").getConstructor(int.class,int.class,int.class).newInstance(x,y+1,z);
                        }
                        return cls("jd").getConstructor(int.class,int.class,int.class).newInstance(x,-64,z);
                    }
                    if(values!=null&&values.length>0&&cls("jd").isInstance(values[0])) {
                        int x=(Integer)getX.invoke(values[0]),y=(Integer)getY.invoke(values[0]),z=(Integer)getZ.invoke(values[0]);
                        if(name.equals("a_"))return read.apply(new int[]{x,y,z});
                        if(name.equals("b_"))return fluid.invoke(read.apply(new int[]{x,y,z}));
                        if(name.equals("c_"))return entities.get(x+" "+y+" "+z);
                        if(name.equals("a")&&values.length==2&&cls("dqj").isInstance(values[1]))return java.util.Optional.ofNullable(entities.get(x+" "+y+" "+z));
                        if(name.equals("a")&&values.length==3&&cls("dtc").isInstance(values[1])) {
                            String key=x+" "+y+" "+z;changed.put(key,values[1]);
                            Object block=cls("dtb$a").getMethod("b").invoke(values[1]);
                            if(block==cls("dga").getField("cv").get(null)||block==cls("dga").getField("J").get(null)||block==cls("dga").getField("aU").get(null)) {
                                if(!entities.containsKey(key))entities.put(key,cls(block==cls("dga").getField("cv").get(null)?"dqo":block==cls("dga").getField("aU").get(null)?"dra":"dql").getConstructor(cls("jd"),cls("dtc")).newInstance(values[0],values[1]));
                            }else entities.remove(key);
                            return true;
                        }
                        if(name.equals("y")) {
                            String key=(x>>4)+" "+(z>>4);
                            if(!chunks.containsKey(key))chunks.put(key,cls("dvs").getConstructor(cls("dcd"),cls("dvv"),cls("dcy"),cls("jz"),cls("eab"))
                                .newInstance(cls("dcd").getConstructor(int.class,int.class).newInstance(x>>4,z>>4),cls("dvv").getField("a").get(null),heightAccessor,biomeRegistry,null));
                            return chunks.get(key);
                        }
                        if(method.getReturnType()==void.class)return null;
                    }
                    throw new UnsupportedOperationException(method.toString());
                });
                Object clip=cls("ejj").getConstructor(int.class,int.class,int.class,int.class,int.class,int.class)
                    .newInstance(chunkX*16,-64,chunkZ*16,chunkX*16+15,319,chunkZ*16+15);
                Object random=pyramid||jungle?cls("dzx").getConstructor(cls("ayw")).newInstance(cls("dzz").getConstructor(long.class).newInstance(worldSeed)):cls("dyz").getConstructor(long.class).newInstance(0L);
                Object chunk=cls("dcd").getConstructor(int.class,int.class).newInstance(chunkX,chunkZ);
                Object origin=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(0,64,0);
                cls("ejv").getMethod("a",cls("dds"),cls("ddq"),cls("duz"),cls("ayw"),cls("ejj"),cls("dcd"),cls("jd"))
                    .invoke(piece,world,null,null,random,clip,chunk,origin);
                if(pyramid)cls("ejr").getMethod("a",cls("dds"),cls("ddq"),cls("duz"),cls("ayw"),cls("ejj"),cls("dcd"),cls("ekg"))
                    .invoke(pyramidStructure,world,null,null,random,clip,chunk,cls("ekg").getConstructor(List.class).newInstance(List.of(piece)));
                Object saved=cls("ejv").getMethod("a",cls("ekh")).invoke(piece,(Object)null);
                int height=(Integer)cls("ub").getMethod("h",String.class).invoke(saved,"HPos");
                List<String> edits=new ArrayList<>(),post=new ArrayList<>();
                for(var entry:changed.entrySet())edits.add(entry.getKey()+" "+stateID.invoke(null,entry.getValue()));
                TreeSet<String> marks=new TreeSet<>();
                for(var entry:chunks.entrySet()) {
                    String[] pos=entry.getKey().split(" ");int cx=Integer.parseInt(pos[0]),cz=Integer.parseInt(pos[1]);
                    Object[] sections=(Object[])cls("duy").getMethod("n").invoke(entry.getValue());
                    for(int section=0;section<sections.length;section++)if(sections[section]!=null)for(Object value:(Iterable<?>)sections[section]) {
                        int v=((Short)value)&65535;marks.add((cx*16+(v&15))+" "+(-64+section*16+((v>>4)&15))+" "+(cz*16+((v>>8)&15)));
                    }
                }
                post.addAll(marks);
                if(!pyramid&&!jungle)System.out.printf("hut %d %d %d %d %d %d %d %d %s %d %s%n",orientation,scenario,anchor[0],anchor[1],chunkX,chunkZ,height,edits.size(),digest(edits),post.size(),digest(post));
                else {
                    List<String> entityValues=new ArrayList<>();
                    for(var entry:entities.entrySet()) {
                        Object entityTag=cls("dqh").getMethod("b",cls("jo$a")).invoke(entry.getValue(),registries);
                        entityValues.add(entry.getKey()+" "+cls("ub").getMethod("l",String.class).invoke(entityTag,"id")+" "+cls("ub").getMethod("l",String.class).invoke(entityTag,"LootTable")+" "+cls("ub").getMethod("i",String.class).invoke(entityTag,"LootTableSeed"));
                    }
                    String flags="";for(String field:jungle?new String[]{"placedMainChest","placedHiddenChest","placedTrap1","placedTrap2"}:new String[]{"hasPlacedChest0","hasPlacedChest1","hasPlacedChest2","hasPlacedChest3"})flags+=((Boolean)cls("ub").getMethod("q",String.class).invoke(saved,field))?"1":"0";
                    int[] savedBox=(int[])cls("ub").getMethod("n",String.class).invoke(saved,"BB");
                    System.out.printf("%s %d %d %d %d %d %d %d %d %d %s %d %s %d %s %s %d %d %d%n",jungle?"jungle":"pyramid",worldSeed,orientation,scenario,anchor[0],anchor[1],chunkX,chunkZ,height,edits.size(),digest(edits),post.size(),digest(post),entities.size(),digest(entityValues),flags,savedBox[1],
                        (Long)cls("ayw").getMethod("g").invoke(random),(Long)cls("ayw").getMethod("g").invoke(worldRandom));
                }
            }
        }
    }
}
