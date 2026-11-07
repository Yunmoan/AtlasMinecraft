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

/** Original StrongholdPiece.postProcess, complete state/NBT/RNG and clipped writes. */
public final class StrongholdBlocksProbe {
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
        Object lava=cls("dfy").getMethod("o").invoke(cls("dga").getField("H").get(null));
        java.lang.reflect.Method getX=cls("jd").getMethod("u"),getY=cls("jd").getMethod("v"),getZ=cls("jd").getMethod("w");
        java.lang.reflect.Method fluid=cls("dtb$a").getMethod("u"),solid=cls("dtb$a").getMethod("d"),emptyFluid=cls("epe").getMethod("c");
        var stateID=cls("dfy").getMethod("i",cls("dtc"));
        String[][] specifications={
            {"n","shs","5","5","7"},{"h","shph","9","5","11"},{"d","shlt","5","5","5"},{"i","shrt","5","5","5"},
            {"j","shrc","11","7","11"},{"o","shssd","5","11","8"},{"l","shsd","5","11","5"},{"c","sh5c","10","9","11"},
            {"a","shcc","5","5","7"},{"e","shli","14","11","15"},{"e","shli_small","14","6","15"},{"g","shpr","11","8","16"},
            {"b","shfc","5","5","3"},{"m","shstart","5","11","5"}
        };
        for(String[] specification:specifications)for(long worldSeed:new long[]{0,1234,99})for(int orientation=0;orientation<4;orientation++)for(int scenario=0;scenario<4;scenario++)for(int[] anchor:new int[][]{{0,0},{15,15},{-16,-16}}) {
            int width=Integer.parseInt(specification[2]),depth=Integer.parseInt(specification[4]);
            int xSize=orientation%2==0?width:depth,zSize=orientation%2==0?depth:width,pieceHeight=Integer.parseInt(specification[3]);
            int[] box={anchor[0],40,anchor[1],anchor[0]+xSize-1,40+pieceHeight-1,anchor[1]+zSize-1};
            for(int dx=0;dx<=1;dx++)for(int dz=0;dz<=1;dz++) {
                int chunkX=(anchor[0]>>4)+dx,chunkZ=(anchor[1]>>4)+dz;
                Object tag=cls("ub").getConstructor().newInstance();
                cls("ub").getMethod("a",String.class,String.class).invoke(tag,"id","minecraft:"+specification[1].replace("_small",""));
                cls("ub").getMethod("a",String.class,int[].class).invoke(tag,"BB",box);
                for(var entry:Map.of("O",orientation,"GD",0,"Width",width,"Height",pieceHeight,"Depth",depth,"HPos",-1).entrySet())
                    cls("ub").getMethod("a",String.class,int.class).invoke(tag,entry.getKey(),entry.getValue());
                if(!pyramid&&!jungle)for(String field:new String[]{"Witch","Cat"})cls("ub").getMethod("a",String.class,boolean.class).invoke(tag,field,true);
                cls("ub").getMethod("a",String.class,String.class).invoke(tag,"EntryDoor",new String[]{"OPENING","WOOD_DOOR","GRATES","IRON_DOOR"}[(scenario+orientation)%4]);
                for(var entry:Map.of("Left",(scenario&1)==0,"Right",(scenario&2)==0,"leftLow",(scenario&1)==0,"leftHigh",(scenario&2)==0,"rightLow",(scenario&1)!=0,"rightHigh",(scenario&2)!=0,"Tall",pieceHeight>6,"Source",specification[1].equals("shstart"),"Chest",worldSeed==1234,"Mob",worldSeed==1234).entrySet())
                    cls("ub").getMethod("a",String.class,boolean.class).invoke(tag,entry.getKey(),entry.getValue());
                cls("ub").getMethod("a",String.class,int.class).invoke(tag,"Type",(scenario+(worldSeed==1234?1:0))%5);
                cls("ub").getMethod("a",String.class,int.class).invoke(tag,"Steps",3);
                Object piece=cls("emj$"+specification[0]).getConstructor(cls("ub")).newInstance(tag);
                Object worldRandom=cls("dyz").getConstructor(long.class).newInstance(worldSeed^0x123456789abcdefL);
                Map<String,Object> changed=new HashMap<>(),chunks=new HashMap<>(),entities=new HashMap<>();
                TreeSet<String> ticks=new TreeSet<>();
                final int mode=scenario;
                java.util.function.Function<int[],Object> read=p->{
                    String key=p[0]+" "+p[1]+" "+p[2];if(changed.containsKey(key))return changed.get(key);
                    int floor=mode==1?42+Math.floorMod(p[0]+2*p[2],5):mode==2?25:mode==3?-65:80;
                    if(p[1]<=floor)return blocks.get("stone");
                    if(mode==2&&p[1]<=60)return blocks.get("water");
                    
                    return blocks.get("air");
                };
                Object world=Proxy.newProxyInstance(cls("dds").getClassLoader(),new Class<?>[]{cls("dds")},(proxy,method,values)->{
                    String name=method.getName();
                    if(name.equals("I_"))return -64;if(name.equals("an"))return 320;if(name.equals("am"))return 384;
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
                        if(name.equals("a")&&values.length==3&&cls("epd").isInstance(values[1])){ticks.add(x+" "+y+" "+z);return null;}
                        if(name.equals("a_"))return read.apply(new int[]{x,y,z});
                        if(name.equals("b_"))return fluid.invoke(read.apply(new int[]{x,y,z}));
                        if(name.equals("c_"))return entities.get(x+" "+y+" "+z);
                        if(name.equals("a")&&values.length==2&&cls("dqj").isInstance(values[1]))return java.util.Optional.ofNullable(entities.get(x+" "+y+" "+z));
                        if(name.equals("a")&&values.length==3&&cls("dtc").isInstance(values[1])) {
                            String key=x+" "+y+" "+z;changed.put(key,values[1]);
                            Object block=cls("dtb$a").getMethod("b").invoke(values[1]);
                            if(cls("diq").isInstance(block)) {
                                if(!entities.containsKey(key))entities.put(key,cls("diq").getMethod("a",cls("jd"),cls("dtc")).invoke(block,values[0],values[1]));
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
                Object random=cls("dzx").getConstructor(cls("ayw")).newInstance(cls("dzz").getConstructor(long.class).newInstance(worldSeed));
                Object rawRandom=random;
                if(specification[1].equals("shpr")&&worldSeed==99)random=Proxy.newProxyInstance(cls("ayw").getClassLoader(),new Class<?>[]{cls("ayw")},(proxy,method,values)->{
                    Object result=method.invoke(rawRandom,values);return method.getName().equals("i")&&method.getParameterCount()==0?0.95f:result;
                });
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
                List<String> entityValues=new ArrayList<>();
                for(var entry:entities.entrySet()) {
                    Object entityTag=cls("dqh").getMethod("b",cls("jo$a")).invoke(entry.getValue(),registries);
                    entityValues.add(entry.getKey()+" "+IglooProbeSupport.canonicalNBT(entityTag));
                }
                String savedHash=digest(new ArrayList<>(List.of(IglooProbeSupport.canonicalNBT(saved))));
                System.out.printf("stronghold_block %s %d %d %d %d %d %d %d %d %s %d %s %d %s %d %s %s %d%n",specification[1],worldSeed,orientation,scenario,anchor[0],anchor[1],chunkX,chunkZ,edits.size(),digest(edits),post.size(),digest(post),entities.size(),digest(entityValues),ticks.size(),digest(new ArrayList<>(ticks)),savedHash,(Long)cls("ayw").getMethod("g").invoke(random));
            }
        }
    }
}
