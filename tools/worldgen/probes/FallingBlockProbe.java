package atlasprobe;

import com.google.gson.Gson;
import java.io.*;
import java.lang.reflect.*;
import java.util.*;
import java.util.stream.Stream;

public final class FallingBlockProbe {
    static Class<?> c(String n)throws Exception{return Class.forName(n);}
    static final class Loader extends ClassLoader {Class<?> define(byte[] data){return defineClass("FallingBlockLevel",data,0,data.length);}}
    public static void main(String[]args)throws Exception {
        Object version=c("t").getMethod("a").invoke(null);c("ab").getMethod("a",c("ae")).invoke(null,version);c("akt").getMethod("a").invoke(null);RuinedPortalProbeSupport.bindTags();
        Class<?> level;try(InputStream in=FallingBlockProbe.class.getResourceAsStream("/FallingBlockLevel.class")){level=new Loader().define(in.readAllBytes());}
        Field f=c("sun.misc.Unsafe").getDeclaredField("theUnsafe");f.setAccessible(true);Object unsafe=f.get(null),world=c("sun.misc.Unsafe").getMethod("allocateInstance",Class.class).invoke(unsafe,level);
        Object registries=IglooProbeSupport.registryAccess(),lookup=c("jo$a").getMethod("b").invoke(registries);
        level.getField("access").set(world,registries);
        level.getField("damage").set(world,c("brl").getConstructor(c("ka")).newInstance(registries));
        Object chunks=c("sun.misc.Unsafe").getMethod("allocateInstance",Class.class).invoke(unsafe,c("aqs")),map=c("sun.misc.Unsafe").getMethod("allocateInstance",Class.class).invoke(unsafe,c("aqb"));
        Field entityMap=c("aqb").getDeclaredField("K");entityMap.setAccessible(true);entityMap.set(map,new it.unimi.dsi.fastutil.ints.Int2ObjectOpenHashMap<>());
        Field chunkMap=c("aqs").getDeclaredField("a");chunkMap.setAccessible(true);chunkMap.set(chunks,map);level.getField("chunkSource").set(world,chunks);
        Object dimensions=c("jn$a").getMethod("b",c("akq")).invoke(lookup,c("lu").getField("aK").get(null));
        for(Object h:((Stream<?>)c("jo").getMethod("b").invoke(dimensions)).toList()) {Object key=((Optional<?>)c("jm").getMethod("e").invoke(h)).orElseThrow();if(c("akq").getMethod("a").invoke(key).toString().equals("minecraft:overworld"))level.getField("dimension").set(world,c("jm").getMethod("a").invoke(h));}
        Method setup=level.getMethod("configure",int.class,int.class,int.class,long.class),run=level.getMethod("run",String.class,int.class),tags=level.getMethod("entityTags");Field events=level.getField("events");Gson gson=new Gson();
        System.setOut(new PrintStream(new BufferedOutputStream(new FileOutputStream(FileDescriptor.out),65536),false));System.setErr(new PrintStream(new FileOutputStream(FileDescriptor.err),true));
        boolean brush=args.length>1&&args[1].equals("brushable");level.getField("brushProbe").setBoolean(world,brush);
        if(args.length>1&&args[1].equals("kelp")) {
            for(int id:(int[])level.getMethod("kelpStates").invoke(null))for(String belowName:new String[]{"air","stone","water","magma_block","kelp","kelp_plant","sand","soul_sand","oak_fence","stone_slab","ice","oak_leaves"})for(long seed:new long[]{0,1,-1,1234,987654321})for(boolean drops:new boolean[]{false,true}) {
                int below=(int)level.getMethod("namedState",String.class).invoke(null,belowName);long entitySeed=seed^0x123456789abcdef0L;
                var row=new LinkedHashMap<String,Object>();row.put("state",id);row.put("below",below);row.put("seed",seed);row.put("entitySeed",entitySeed);row.put("drops",drops);
                try {
                    row.put("entities",level.getMethod("runKelp",int.class,int.class,long.class,long.class,boolean.class).invoke(world,id,below,seed,entitySeed,drops));
                    row.put("destroyed",level.getField("kelpDestroyed").getBoolean(world));row.put("events",events.get(world));row.put("itemNext",level.getMethod("kelpItemNext").invoke(world));
                    row.put("next",Long.toUnsignedString((long)c("ayw").getMethod("g").invoke(level.getField("random").get(world))));
                }catch(InvocationTargetException e){System.out.flush();throw new IllegalStateException(gson.toJson(row),e.getCause());}
                System.out.println("falling_block "+gson.toJson(row));
            }
            if(!"true".equals(System.getProperty("atlas.entitySeedApplied")))throw new IllegalStateException("Entity seed instrumentation missing");
            System.out.flush();return;
        }
        if(brush&&args[0].equals("callbacks")) {
            for(int id:(int[])level.getMethod("brushStates").invoke(null))for(int below:new int[]{0,1,80,86,90,2089,12543})for(int y:new int[]{-65,-64,64,319,320})for(int count=-1;count<10;count++)for(long now:new long[]{99,100,104}) {
                var row=new LinkedHashMap<String,Object>();row.put("state",id);row.put("below",below);row.put("y",y);row.put("count",count);row.put("now",now);
                try {
                    level.getMethod("configureBrush",int.class,int.class,int.class,int.class,long.class).invoke(world,id,below,y,count,now);
                    row.put("next",Long.toUnsignedString((long)run.invoke(world,"tick",0)));
                    row.put("events",events.get(world));row.put("entities",tags.invoke(world));row.put("progress",level.getMethod("brushProgress").invoke(world));
                }catch(InvocationTargetException e){System.out.flush();throw new IllegalStateException(gson.toJson(row),e.getCause());}
                System.out.println("falling_block "+gson.toJson(row));
            }
            System.out.flush();return;
        }
        if(args.length>0&&args[0].equals("drops")) {
            for(String block:brush?new String[]{"suspicious_sand","suspicious_gravel"}:new String[]{"sand","red_sand","gravel"})for(long seed:new long[]{0,1,-1,1234,987654321})for(int scene=0;scene<10;scene++)for(boolean drops:new boolean[]{false,true})for(boolean dropItem:new boolean[]{false,true})for(boolean cancel:new boolean[]{false,true}) {
                int id=(int)level.getMethod("namedState",String.class).invoke(null,block);
                // Slab top at 60.5 collides while the landing cell is occupied.
                String floor=scene<2?"stone_slab":"air";
                int support=(int)level.getMethod("namedState",String.class).invoke(null,floor);
                int age=switch(scene){case 2,4,6,8,9->100;case 3->99;case 5->600;case 7->599;default->0;};
                double y=switch(scene){case 0->60.5;case 1->60.75;case 2,3->-64;case 4->321;case 8->321.5;case 9->-63;default->64;};
                long entitySeed=seed^0x123456789abcdef0L;
                Map<String,Object> row=new LinkedHashMap<>();row.put("state",id);row.put("support",support);row.put("seed",seed);row.put("entitySeed",entitySeed);row.put("age",age);row.put("y",y);row.put("drops",drops);row.put("dropItem",dropItem);row.put("cancel",cancel);
                try {
                    level.getMethod("configureDrop",int.class,int.class,long.class,long.class,boolean.class,boolean.class,boolean.class,int.class,double.class).invoke(world,id,support,seed,entitySeed,drops,dropItem,cancel,age,y);
                    row.put("entities",level.getMethod("dropTick").invoke(world));row.put("removed",level.getMethod("removed").invoke(world));row.put("events",new ArrayList<>((List<?>)events.get(world)));
                    row.put("itemNext",level.getMethod("itemRandomNext").invoke(world));
                    row.put("next",Long.toUnsignedString((long)c("ayw").getMethod("g").invoke(level.getField("random").get(world))));
                } catch(InvocationTargetException e){System.out.flush();throw new IllegalStateException(gson.toJson(row),e.getCause());}
                System.out.println("falling_drops "+gson.toJson(row));
            }
            if(!"true".equals(System.getProperty("atlas.entitySeedApplied")))throw new IllegalStateException("Entity seed instrumentation was not applied");
            System.out.flush();return;
        }
        if(args.length>0&&args[0].equals("motion")) {
            int[] motionStates=brush?(int[])level.getMethod("brushStates").invoke(null):new int[]{(int)level.getMethod("namedState",String.class).invoke(null,"sand"),(int)level.getMethod("namedState",String.class).invoke(null,"red_sand"),(int)level.getMethod("namedState",String.class).invoke(null,"gravel")};
            for(int id:motionStates)for(String floor:new String[]{"stone","oak_fence","water","air"}) {
                int support=(int)level.getMethod("namedState",String.class).invoke(null,floor);
                level.getMethod("configureMotion",int.class,int.class).invoke(world,id,support);
                for(int tick=0;tick<=30;tick++) {
                    Map<String,Object> row=new LinkedHashMap<>();row.put("state",id);row.put("support",support);row.put("tick",tick);
                    try{row.put("entities",tick==0?tags.invoke(world):level.getMethod("motionTick").invoke(world));row.put("removed",level.getMethod("removed").invoke(world));row.put("events",new ArrayList<>((List<?>)events.get(world)));}
                    catch(InvocationTargetException e){System.out.flush();throw new IllegalStateException(gson.toJson(row),e.getCause());}
                    System.out.println("falling_motion "+gson.toJson(row));((List<?>)events.get(world)).clear();
                }
            }
            System.out.flush();return;
        }
        for(int id:(int[])level.getMethod("states").invoke(null))for(int below:new int[]{0,1,80,86,90,2089,12543})for(int y:new int[]{-65,-64,64,319,320})for(String action:new String[]{"tick","place","shape"}) {
            setup.invoke(world,id,below,y,1234L);Map<String,Object> row=new LinkedHashMap<>();row.put("state",id);row.put("below",below);row.put("y",y);row.put("action",action);
            try{row.put("next",Long.toUnsignedString((long)run.invoke(world,action,0)));row.put("events",events.get(world));row.put("entities",tags.invoke(world));}
            catch(InvocationTargetException e){System.out.flush();throw new IllegalStateException(gson.toJson(row),e.getCause());}
            System.out.println("falling_block "+gson.toJson(row));
        }
        System.out.flush();
    }
}
