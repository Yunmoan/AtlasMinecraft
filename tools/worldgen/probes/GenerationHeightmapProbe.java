package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.util.*;
import java.util.stream.Stream;
import com.mojang.serialization.Lifecycle;

/** Original ProtoChunk changes heightmap maintenance sets at CARVERS. */
public final class GenerationHeightmapProbe {
    static Class<?> cls(String name)throws Exception{return Class.forName(name);}
    static Object call(Object receiver,String owner,String name,Class<?>[] types,Object...args)throws Exception {
        return cls(owner).getMethod(name,types).invoke(receiver,args);
    }
    public static void main(String[]args)throws Exception {
        Object version=call(null,"t","a",new Class<?>[]{});call(null,"ab","a",new Class<?>[]{cls("ae")},version);call(null,"akt","a",new Class<?>[]{});
        Object biomeRegistry=cls("ju").getConstructor(cls("akq"),Lifecycle.class).newInstance(cls("lu").getField("aF").get(null),Lifecycle.stable());
        Object registries=call(null,"pd","a",new Class<?>[]{}),lookup=call(registries,"jo$a","b",new Class<?>[]{});
        Object biomes=call(lookup,"jn$a","b",new Class<?>[]{cls("akq")},cls("lu").getField("aF").get(null));
        for(Object holder:((Stream<?>)call(biomes,"jo","b",new Class<?>[]{})).toList())
            call(null,"jz","a",new Class<?>[]{cls("jz"),cls("akq"),Object.class},biomeRegistry,call(holder,"jm$c","h",new Class<?>[]{}),call(holder,"jm","a",new Class<?>[]{}));
        Object heights=call(null,"dcy","e",new Class<?>[]{int.class,int.class},-64,384);
        Object chunk=cls("dvs").getConstructor(cls("dcd"),cls("dvv"),cls("dcy"),cls("jz"),cls("eab")).newInstance(
            cls("dcd").getConstructor(int.class,int.class).newInstance(-1,2),cls("dvv").getField("a").get(null),heights,biomeRegistry,null);
        Map<String,Object> states=new HashMap<>();Object registry=cls("lt").getField("e").get(null);
        for(String name:new String[]{"stone","water","air","oak_log","oak_leaves","short_grass"}) {
            Object block=call(registry,"jz","a",new Class<?>[]{cls("akr")},call(null,"akr","a",new Class<?>[]{String.class},"minecraft:"+name));
            states.put(name,call(block,"dfy","o",new Class<?>[]{}));
        }
        for(int x=0;x<7;x++) {
            put(chunk,x,60+x,states.get("stone"));
            put(chunk,x,63+x,states.get("water"));
        }
        call(chunk,"dvs","a",new Class<?>[]{cls("dvz")},cls("dvz").getField("i").get(null));
        @SuppressWarnings({"unchecked","rawtypes"}) EnumSet maps=EnumSet.noneOf((Class)cls("dyy$a"));
        for(String name:new String[]{"d","b","e","f"})maps.add(cls("dyy$a").getField(name).get(null));
        call(null,"dyy","a",new Class<?>[]{cls("duy"),Set.class},chunk,maps);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        for(int phase=0;phase<7;phase++) {
            for(int x=0;x<8;x++) {
                if(phase==1)put(chunk,x,75+x,states.get("oak_log"));
                if(phase==2)put(chunk,x,76+x,states.get("oak_leaves"));
                if(phase==3)put(chunk,x,77+x,states.get("water"));
                if(phase==4)put(chunk,x,78+x,states.get("short_grass"));
                if(phase==5)for(int y=75+x;y<=78+x;y++)put(chunk,x,y,states.get("air"));
                if(phase==6)for(int y=60+x;y<=63+x;y++)put(chunk,x,y,states.get("air"));
                System.out.printf("generation_heightmap %d %d",phase,x);
                for(String name:new String[]{"c","a","d","b"}) {
                    int height=(Integer)call(chunk,"duy","a",new Class<?>[]{cls("dyy$a"),int.class,int.class},cls("dyy$a").getField(name).get(null),x,0);
                    System.out.print(" "+(height+1));
                }
                System.out.println();
            }
        }
    }
    static void put(Object chunk,int x,int y,Object state)throws Exception {
        Object p=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(x-16,y,32);
        call(chunk,"dvs","a",new Class<?>[]{cls("jd"),cls("dtc"),boolean.class},p,state,false);
    }
}
