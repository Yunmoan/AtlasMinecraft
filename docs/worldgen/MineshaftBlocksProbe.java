package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.util.*;
import java.util.function.Function;

/** Original complete piece block calls, with independent pending minecart
 * requests. Full minecart runtime defaults/finalization are a separate stage. */
public final class MineshaftBlocksProbe {
    static Class<?> cls(String n)throws Exception{return Class.forName(n);}
    static Object block(String n)throws Exception{return RuinedPortalProbeSupport.block(n);}
    static Object piece(String spec,int mesa,int orientation,int ax,int az)throws Exception {
        int width=3,height=3,depth=20;String type="a",id="mscorridor";
        if(spec.equals("rail"))depth=15;
        if(spec.equals("spider")||spec.equals("spawned"))depth=10;
        if(spec.contains("crossing")){width=5;depth=5;height=spec.startsWith("tall")?7:3;type="b";id="mscrossing";}
        if(spec.equals("stairs")){height=8;depth=9;type="e";id="msstairs";}
        if(spec.equals("room")){width=11;height=9;depth=13;type="d";id="msroom";}
        int sx=orientation%2==0?width:depth,sz=orientation%2==0?depth:width;
        if(spec.equals("room")){sx=width;sz=depth;}
        Object tag=cls("ub").getConstructor().newInstance();cls("ub").getMethod("a",String.class,String.class).invoke(tag,"id","minecraft:"+id);
        cls("ub").getMethod("a",String.class,int[].class).invoke(tag,"BB",new int[]{ax,64,az,ax+sx-1,64+height-1,az+sz-1});
        for(var entry:Map.of("O",spec.contains("crossing")||spec.equals("room")?-1:orientation,"GD",0,"MST",mesa,"D",orientation,"Num",depth/5).entrySet())cls("ub").getMethod("a",String.class,int.class).invoke(tag,entry.getKey(),entry.getValue());
        for(var entry:Map.of("hr",spec.equals("rail"),"sc",spec.equals("spider")||spec.equals("spawned"),"hps",spec.equals("spawned"),"tf",spec.equals("tallcrossing")).entrySet())cls("ub").getMethod("a",String.class,boolean.class).invoke(tag,entry.getKey(),entry.getValue());
        if(spec.equals("room")){
            Object entrances=cls("uh").getConstructor().newInstance();
            for(int[]bb:new int[][]{{ax+2,65,az,ax+4,67,az+1},{ax+sx-2,66,az+5,ax+sx-1,68,az+7}}){Object temp=cls("ub").getConstructor().newInstance();cls("ub").getMethod("a",String.class,int[].class).invoke(temp,"v",bb);((List<Object>)entrances).add(cls("ub").getMethod("c",String.class).invoke(temp,"v"));}
            cls("ub").getMethod("a",String.class,cls("uy")).invoke(tag,"Entrances",entrances);
        }
        return cls("elv$"+type).getConstructor(cls("ub")).newInstance(tag);
    }
    static Object terrain(int mode,int mesa,int x,int y,int z,Map<String,Object>states){
        if(mode==4)return states.get("air");
        if(mode==2&&y==65)return states.get("water");
        if(mode==7&&Math.floorMod(x+y+z,7)==0)return states.get(new String[]{mesa==0?"oak_planks":"dark_oak_planks",mesa==0?"oak_log":"dark_oak_log",mesa==0?"oak_fence":"dark_oak_fence","chain"}[Math.floorMod(x+z,4)]);
        if(mode==0||mode==2||mode==3||mode==7)return y<=100?states.get("stone"):states.get("air");
        int floor=mode==1?50:40,roof=mode==1?72:mode==5?90:mode==6?100:mode==8?73:71;
        if(y<=floor)return states.get("stone");
        if(mode==5&&y<=50)return states.get("water");
        if(mode==6&&y<=50)return states.get("lava");
        if(y>=roof)return states.get(y==roof&&mode==8?"sand":y==roof&&mode==9?"oak_leaves":"stone");
        return states.get("air");
    }
    public static void main(String[]args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);RuinedPortalProbeSupport.bindTags();
        Object registries=IglooProbeSupport.registryAccess();RuinedPortalStartsProbe.bindBiomeTags(registries);Object biomeRegistry=cls("ka").getMethod("d",cls("akq")).invoke(registries,cls("lu").getField("aF").get(null));
        Map<String,Object>states=new HashMap<>();for(String n:new String[]{"air","stone","water","lava","sand","oak_leaves","oak_planks","oak_log","oak_fence","dark_oak_planks","dark_oak_log","dark_oak_fence","chain"})states.put(n,block(n));
        Object level=OceanRuinBlocksProbe.level(1234,registries),heights=cls("dcy").getMethod("e",int.class,int.class).invoke(null,-64,384);
        Object plains=((Optional<?>)cls("jz").getMethod("c",cls("akr")).invoke(biomeRegistry,cls("akr").getMethod("a",String.class).invoke(null,"minecraft:plains"))).orElseThrow(),deepDark=((Optional<?>)cls("jz").getMethod("c",cls("akr")).invoke(biomeRegistry,cls("akr").getMethod("a",String.class).invoke(null,"minecraft:deep_dark"))).orElseThrow();
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        for(String spec:new String[]{"corridor","rail","spider","spawned","crossing","tallcrossing","stairs","room"})for(long seed:new long[]{0,1234})for(int mesa=0;mesa<2;mesa++)for(int orientation=0;orientation<4;orientation++)for(int mode=0;mode<10;mode++)for(int[]anchor:new int[][]{{0,0},{15,15},{-16,-16}})for(int clipMode=0;clipMode<3;clipMode++) {
            Object piece=piece(spec,mesa,orientation,anchor[0],anchor[1]);final int scenario=mode,mineType=mesa;
            Map<String,Object>edits=new HashMap<>(),entities=new HashMap<>(),chunks=new HashMap<>();TreeSet<String>ticks=new TreeSet<>();List<String>actors=new ArrayList<>();
            Function<int[],Object>read=p->{String key=p[0]+" "+p[1]+" "+p[2];return edits.containsKey(key)?edits.get(key):terrain(scenario,mineType,p[0],p[1],p[2],states);};
            Object delegate=RuinedPortalProbeSupport.world(0,edits,entities,registries);
            Object world=Proxy.newProxyInstance(cls("dds").getClassLoader(),new Class<?>[]{cls("dds")},(proxy,method,v)->{
                String name=method.getName();if(name.equals("E")&&method.getParameterCount()==0)return level;if(name.equals("t"))return scenario==3?deepDark:plains;
                if(v!=null&&v.length>=3&&cls("jd").isInstance(v[0])&&cls("epd").isInstance(v[1]))ticks.add(cls("jd").getMethod("u").invoke(v[0])+" "+cls("jd").getMethod("v").invoke(v[0])+" "+cls("jd").getMethod("w").invoke(v[0]));
                if(name.equals("b")&&v!=null&&v.length==1&&cls("bsr").isInstance(v[0])){
                    Object saved=cls("ub").getConstructor().newInstance();cls("bsr").getMethod("e",cls("ub")).invoke(v[0],saved);Object request=cls("ub").getConstructor().newInstance();
                    for(String key:new String[]{"id","Pos","LootTable","LootTableSeed"})cls("ub").getMethod("a",String.class,cls("uy")).invoke(request,key,cls("ub").getMethod("c",String.class).invoke(saved,key));actors.add(IglooProbeSupport.canonicalNBT(request));return method.getReturnType()==boolean.class?true:null;
                }
                if(v!=null&&v.length==3&&cls("dyy$a").isInstance(v[0])){java.util.function.Predicate<Object>predicate=(java.util.function.Predicate<Object>)cls("dyy$a").getMethod("e").invoke(v[0]);for(int y=319;y>=-64;y--)if(predicate.test(read.apply(new int[]{(Integer)v[1],y,(Integer)v[2]})))return y+1;return -64;}
                if(v!=null&&v.length>0&&cls("jd").isInstance(v[0])){
                    int x=(Integer)cls("jd").getMethod("u").invoke(v[0]),y=(Integer)cls("jd").getMethod("v").invoke(v[0]),z=(Integer)cls("jd").getMethod("w").invoke(v[0]);
                    if(name.equals("a_"))return read.apply(new int[]{x,y,z});if(name.equals("b_"))return cls("dtb$a").getMethod("u").invoke(read.apply(new int[]{x,y,z}));
                    if(name.equals("y")){String key=(x>>4)+" "+(z>>4);if(!chunks.containsKey(key))chunks.put(key,cls("dvs").getConstructor(cls("dcd"),cls("dvv"),cls("dcy"),cls("jz"),cls("eab")).newInstance(cls("dcd").getConstructor(int.class,int.class).newInstance(x>>4,z>>4),cls("dvv").getField("a").get(null),heights,biomeRegistry,null));return chunks.get(key);}
                }
                try{return method.invoke(delegate,v);}catch(InvocationTargetException error){throw error.getCause();}
            });
            Object random=cls("dzx").getConstructor(cls("ayw")).newInstance(cls("dzz").getConstructor(long.class).newInstance(seed)),box=cls("ejv").getMethod("f").invoke(piece);int minX=(Integer)cls("ejj").getMethod("h").invoke(box)>>4,maxX=(Integer)cls("ejj").getMethod("k").invoke(box)>>4,minZ=(Integer)cls("ejj").getMethod("j").invoke(box)>>4,maxZ=(Integer)cls("ejj").getMethod("m").invoke(box)>>4;
            if(clipMode<2){minX=maxX=(anchor[0]>>4)+(clipMode==1?3:0);minZ=maxZ=anchor[1]>>4;}
            for(int x=minX;x<=maxX;x++)for(int z=minZ;z<=maxZ;z++)cls("ejv").getMethod("a",cls("dds"),cls("ddq"),cls("duz"),cls("ayw"),cls("ejj"),cls("dcd"),cls("jd")).invoke(piece,world,null,null,random,cls("ejj").getConstructor(int.class,int.class,int.class,int.class,int.class,int.class).newInstance(x*16,-64,z*16,x*16+15,319,z*16+15),cls("dcd").getConstructor(int.class,int.class).newInstance(x,z),cls("jd").getConstructor(int.class,int.class,int.class).newInstance(anchor[0],64,anchor[1]));
            List<String>blocks=new ArrayList<>(),blockEntities=new ArrayList<>(),post=new ArrayList<>();for(var entry:edits.entrySet())blocks.add(entry.getKey()+" "+cls("dfy").getMethod("i",cls("dtc")).invoke(null,entry.getValue()));for(var entry:entities.entrySet())blockEntities.add(entry.getKey()+" "+IglooProbeSupport.canonicalNBT(cls("dqh").getMethod("b",cls("jo$a")).invoke(entry.getValue(),registries)));
            for(var entry:chunks.entrySet()){String[]p=entry.getKey().split(" ");int cx=Integer.parseInt(p[0]),cz=Integer.parseInt(p[1]);Object[]sections=(Object[])cls("duy").getMethod("n").invoke(entry.getValue());for(int section=0;section<sections.length;section++)if(sections[section]!=null)for(Object value:(Iterable<?>)sections[section]){int v=((Short)value)&65535;post.add((cx*16+(v&15))+" "+(-64+section*16+((v>>4)&15))+" "+(cz*16+((v>>8)&15)));}}
            post=new ArrayList<>(new TreeSet<>(post));Object saved=cls("ejv").getMethod("a",cls("ekh")).invoke(piece,(Object)null);
            String debugKey=spec+" "+seed+" "+mesa+" "+orientation+" "+mode+" "+anchor[0]+" "+anchor[1]+" "+clipMode;
            if("1".equals(System.getenv("ATLAS_MINESHAFT_DEBUG"))||debugKey.equals(System.getenv("ATLAS_MINESHAFT_DEBUG"))){blocks.sort(String::compareTo);java.nio.file.Files.writeString(java.nio.file.Path.of(System.getenv("ATLAS_MINESHAFT_DEBUG_OUTPUT")),String.join("\n",blocks));System.err.println(saved+" post="+post+" actors="+actors);throw new IllegalStateException("explicit mineshaft debug stop");}
            System.out.printf("mineblocks %s %d %d %d %d %d %d %d %d %s %d %s %d %s %d %s %d %s %s %d%n",spec,seed,mesa,orientation,mode,anchor[0],anchor[1],clipMode,blocks.size(),RuinedPortalBlocksProbe.digest(blocks),post.size(),RuinedPortalBlocksProbe.digest(post),entities.size(),RuinedPortalBlocksProbe.digest(blockEntities),ticks.size(),RuinedPortalBlocksProbe.digest(new ArrayList<>(ticks)),actors.size(),RuinedPortalBlocksProbe.digest(actors),RuinedPortalStartsProbe.digest(IglooProbeSupport.canonicalNBT(saved)),cls("ayw").getMethod("g").invoke(random));
        }
    }
}
