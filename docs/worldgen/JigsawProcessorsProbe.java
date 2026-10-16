package atlasprobe;

import com.google.gson.JsonElement;
import com.google.gson.JsonObject;
import com.google.gson.JsonParser;
import com.mojang.serialization.Codec;
import com.mojang.serialization.DynamicOps;
import com.mojang.serialization.JsonOps;
import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.lang.reflect.Proxy;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.TreeMap;
import java.util.jar.JarFile;

/** Exercises all original 40 configured processor lists against varied full
 * input palettes, NBT, transforms, protected terrain and capped world seeds. */
public final class JigsawProcessorsProbe {
    static Class<?> cls(String name)throws Exception{return Class.forName(name);}
    static JsonElement canonicalJSON(JsonElement value) {
        if(!value.isJsonObject())return value;
        var result=new JsonObject();var sorted=new TreeMap<String,JsonElement>();
        for(var entry:value.getAsJsonObject().entrySet())sorted.put(entry.getKey(),entry.getValue());
        for(var entry:sorted.entrySet())result.add(entry.getKey(),canonicalJSON(entry.getValue()));
        return result;
    }
    static void collectStates(JsonElement value,Map<String,JsonObject> states) {
        if(value.isJsonArray()){for(var child:value.getAsJsonArray())collectStates(child,states);return;}
        if(!value.isJsonObject())return;
        var object=value.getAsJsonObject();
        if(object.has("Name")){var canonical=canonicalJSON(object).getAsJsonObject();states.put(canonical.toString(),canonical);}
        if(object.has("block")){var state=new JsonObject();state.addProperty("Name",object.get("block").getAsString());states.put(state.toString(),state);}
        for(var entry:object.entrySet())collectStates(entry.getValue(),states);
    }
    static Object seedLevel(long seed)throws Exception {
        var f=sun.misc.Unsafe.class.getDeclaredField("theUnsafe");f.setAccessible(true);var unsafe=(sun.misc.Unsafe)f.get(null);
        Object server=unsafe.allocateInstance(cls("apn")),level=unsafe.allocateInstance(cls("aqu"));
        Object settings=cls("dzw").getConstructor(long.class,boolean.class,boolean.class).newInstance(seed,true,false);
        Object worldData=Proxy.newProxyInstance(cls("erl").getClassLoader(),new Class<?>[]{cls("erl")},(p,m,v)->{if(m.getName().equals("y"))return settings;throw new UnsupportedOperationException(m.toString());});
        var serverData=cls("net.minecraft.server.MinecraftServer").getDeclaredField("j");serverData.setAccessible(true);serverData.set(server,worldData);
        var levelServer=cls("aqu").getDeclaredField("J");levelServer.setAccessible(true);levelServer.set(level,server);
        return level;
    }
    public static void main(String[] args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);
        RuinedPortalProbeSupport.bindTags();Object registries=IglooProbeSupport.registryAccess();
        DynamicOps ops=(DynamicOps)cls("ka").getMethod("a",DynamicOps.class).invoke(registries,JsonOps.INSTANCE);
        Map<String,JsonObject> processors=new TreeMap<>(),states=new TreeMap<>();
        try(JarFile jar=new JarFile(new java.io.File(cls("ent").getProtectionDomain().getCodeSource().getLocation().toURI()))) {
            var entries=jar.entries();while(entries.hasMoreElements()) {
                var entry=entries.nextElement();String name=entry.getName(),prefix="data/minecraft/worldgen/processor_list/";
                if(!name.startsWith(prefix)||!name.endsWith(".json"))continue;
                var data=JsonParser.parseString(new String(jar.getInputStream(entry).readAllBytes(),StandardCharsets.UTF_8)).getAsJsonObject();
                processors.put("minecraft:"+name.substring(prefix.length(),name.length()-5),data);collectStates(data,states);
            }
        }
        for(String name:new String[]{"gravel","white_wool","red_wool","white_carpet","red_carpet","stone","bedrock","obsidian","air"}) {
            var state=new JsonObject();state.addProperty("Name","minecraft:"+name);states.put(state.toString(),state);
        }
        Codec stateCodec=(Codec)cls("dtc").getField("b").get(null),elementCodec=(Codec)cls("ekz").getField("f").get(null);
        List<Object> palette=new ArrayList<>();int index=0;
        for(var encoded:states.values())for(int repeat=0;repeat<4;repeat++) {
            Object state=stateCodec.parse(ops,encoded).getOrThrow(),p=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(index%12,(index/144)*3,(index/12)%12),data=null;
            if(index%7==0){data=cls("ub").getConstructor().newInstance();cls("ub").getMethod("a",String.class,int.class).invoke(data,"atlas_marker",index);}
            palette.add(cls("ent$c").getConstructor(cls("jd"),cls("dtc"),cls("ub")).newInstance(p,state,data));index++;
        }
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));System.setErr(new PrintStream(new FileOutputStream(FileDescriptor.err),true));
        System.err.println("original processor lists="+processors.size()+" palette blocks="+palette.size());
        var settingsMethod=cls("eky").getDeclaredMethod("a",cls("dmm"),cls("ejj"),cls("end"),boolean.class);settingsMethod.setAccessible(true);
        Object clip=cls("ejj").getConstructor(int.class,int.class,int.class,int.class,int.class,int.class).newInstance(-1000,-64,-1000,1000,319,1000);
        for(String name:processors.keySet())for(int rotation=0;rotation<4;rotation++)for(int[] anchor:new int[][]{{15,64,15},{-16,32,-16},{100,130,-120}})for(int mode:new int[]{0,4,5})for(long seed:new long[]{0,1234,Long.MAX_VALUE}) {
            var encoded=new JsonObject();encoded.addProperty("element_type","minecraft:single_pool_element");encoded.addProperty("location","minecraft:bastion/blocks/air");encoded.addProperty("projection","rigid");encoded.addProperty("processors",name);
            Object element=elementCodec.parse(ops,encoded).getOrThrow(),settings=settingsMethod.invoke(element,cls("dmm").getEnumConstants()[rotation],clip,cls("end").getEnumConstants()[1],false);
            Object origin=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(anchor[0],anchor[1],anchor[2]),pivot=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(anchor[0]+5,48,anchor[2]-7);
            Object base=RuinedPortalProbeSupport.world(mode,new HashMap<>(),new HashMap<>(),registries),level=seedLevel(seed);
            Object world=Proxy.newProxyInstance(cls("dds").getClassLoader(),new Class<?>[]{cls("dds")},(p,m,v)->{if(m.getName().equals("E"))return level;return Proxy.getInvocationHandler(base).invoke(base,m,v);});
            List<?> processed=(List<?>)cls("ent").getMethod("a",cls("ddl"),cls("jd"),cls("jd"),cls("enp"),List.class).invoke(null,world,origin,pivot,settings,palette);
            List<String> rows=new ArrayList<>();
            for(Object info:processed) {
                Object p=cls("ent$c").getMethod("a").invoke(info),state=cls("ent$c").getMethod("b").invoke(info),data=cls("ent$c").getMethod("c").invoke(info);
                rows.add(cls("jd").getMethod("u").invoke(p)+" "+cls("jd").getMethod("v").invoke(p)+" "+cls("jd").getMethod("w").invoke(p)+" "+cls("dfy").getMethod("i",cls("dtc")).invoke(null,state)+" "+(data==null?"null":IglooProbeSupport.canonicalNBT(data)));
            }
            System.out.printf("jigsaw_rules %s %d %d %d %d %d %d %d %s%n",name,seed,rotation,mode,anchor[0],anchor[1],anchor[2],rows.size(),JigsawBlocksProbe.hash(rows,false));
        }
    }
}
