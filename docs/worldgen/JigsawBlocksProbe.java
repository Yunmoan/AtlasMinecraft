package atlasprobe;

import com.google.gson.JsonParser;
import com.google.gson.JsonObject;
import com.mojang.serialization.Codec;
import com.mojang.serialization.DynamicOps;
import com.mojang.serialization.JsonOps;
import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.HexFormat;
import java.util.List;
import java.util.Map;
import java.util.TreeMap;
import java.util.jar.JarFile;
import java.lang.reflect.Proxy;
import com.google.gson.JsonElement;

/** Actual original Bastion SinglePoolElement/StructureTemplate processor and
 * placement calls, with defined terrain, clipping, full block NBT and RNG. */
public final class JigsawBlocksProbe {
    static Class<?> cls(String name)throws Exception{return Class.forName(name);}
    static String hash(List<String> rows,boolean sort)throws Exception {
        if(sort)rows.sort(String::compareTo);
        return HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(String.join("\n",rows).getBytes(StandardCharsets.UTF_8)));
    }
    static void collectElements(JsonObject element,Map<String,JsonObject> elements,boolean overworld)throws Exception {
        String type=element.get("element_type").getAsString();
        if(type.equals("minecraft:empty_pool_element")||type.equals("minecraft:feature_pool_element"))return;
        String name;
        if(overworld) {
            if(type.equals("minecraft:list_pool_element"))for(JsonElement child:element.getAsJsonArray("elements"))collectElements(child.getAsJsonObject(),elements,true);
            name=(element.has("location")?element.get("location").getAsString():"minecraft:ancient_city/list")+":"+hash(new ArrayList<>(List.of(JigsawProcessorsProbe.canonicalJSON(element).toString())),false).substring(0,16);
        }else{name=element.get("location").getAsString();}
        elements.put(name,element);
    }
    static void collectLocations(JsonObject element,List<String> names) {
        if(element.has("location"))names.add(element.get("location").getAsString().replace("minecraft:",""));
        if(element.has("elements"))for(JsonElement child:element.getAsJsonArray("elements"))collectLocations(child.getAsJsonObject(),names);
    }
    static Object world(int mode,Map<String,Object> edits,Map<String,Object> entities,Object registries,boolean overworld)throws Exception {
        Object base=RuinedPortalProbeSupport.world(mode,edits,entities,registries);
        if(!overworld)return base;
        Object level=JigsawProcessorsProbe.seedLevel(1234);
        return Proxy.newProxyInstance(cls("dds").getClassLoader(),new Class<?>[]{cls("dds")},(p,m,v)->{if(m.getName().equals("E"))return level;return Proxy.getInvocationHandler(base).invoke(base,m,v);});
    }
    public static void main(String[] args)throws Exception {run(false);}
    public static void run(boolean overworld)throws Exception {run(overworld,false);}
    public static void run(boolean overworld,boolean outpost)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);
        RuinedPortalProbeSupport.bindTags();Object registries=IglooProbeSupport.registryAccess();
        Map<String,JsonObject> elements=new TreeMap<>();
        try(JarFile jar=new JarFile(new java.io.File(cls("ent").getProtectionDomain().getCodeSource().getLocation().toURI()))) {
            var entries=jar.entries();while(entries.hasMoreElements()) {
                var entry=entries.nextElement();String name=entry.getName();
                String prefix="data/minecraft/worldgen/template_pool/";
                if(!name.startsWith(prefix)||!name.endsWith(".json"))continue;
                String family=name.substring(prefix.length());
                if(outpost?!family.startsWith("pillager_outpost/"):overworld?!(family.startsWith("ancient_city/")||family.startsWith("trail_ruins/")||family.startsWith("trial_chambers/")):!family.startsWith("bastion/"))continue;
                var pool=JsonParser.parseString(new String(jar.getInputStream(entry).readAllBytes(),StandardCharsets.UTF_8)).getAsJsonObject();
                for(var weighted:pool.getAsJsonArray("elements")) {
                    var element=weighted.getAsJsonObject().getAsJsonObject("element");
                    collectElements(element,elements,overworld);
                }
            }
        }
        List<String> locations=new ArrayList<>();for(JsonObject element:elements.values())collectLocations(element,locations);
        Object manager=IglooProbeSupport.manager(locations.stream().distinct().toArray(String[]::new));
        DynamicOps ops=(DynamicOps)cls("ka").getMethod("a",DynamicOps.class).invoke(registries,JsonOps.INSTANCE);
        Codec codec=(Codec)cls("ekz").getField("f").get(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        System.setErr(new PrintStream(new FileOutputStream(FileDescriptor.err),true));
        var templates=cls("ent").getDeclaredField("m");templates.setAccessible(true);
        int actorCount=0;
        for(String name:locations.stream().distinct().toList()) {
            Object template=cls("enu").getMethod("a",cls("akr")).invoke(manager,cls("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name));
            actorCount+=((List<?>)templates.get(template)).size();
        }
        System.err.println("Jigsaw elements="+elements.size()+" authored actors="+actorCount);
        if(overworld&&!outpost&&actorCount!=0)throw new IllegalStateException("unexpected actors in three Overworld Jigsaw families");
        var makeSettings=cls("eky").getDeclaredMethod("a",cls("dmm"),cls("ejj"),cls("end"),boolean.class);makeSettings.setAccessible(true);
        var palettes=cls("ent").getDeclaredField("l");palettes.setAccessible(true);
        for(var entry:elements.entrySet()) {
            String name=entry.getKey();Object element=codec.parse(ops,entry.getValue()).getOrThrow();
            if(overworld)System.out.printf("jigsaw_element %s %s%n",name,java.util.Base64.getEncoder().encodeToString(entry.getValue().toString().getBytes(StandardCharsets.UTF_8)));
            Object template=entry.getValue().has("location")?cls("enu").getMethod("a",cls("akr")).invoke(manager,cls("akr").getMethod("a",String.class).invoke(null,entry.getValue().get("location").getAsString())):null;
            for(int rotation=0;rotation<4;rotation++)for(int[] anchor:new int[][]{{15,64,15},{-16,32,-16}})for(int mode:new int[]{0,2,3}) {
                Object rotate=cls("dmm").getEnumConstants()[rotation],origin=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(anchor[0],anchor[1],anchor[2]);
                Object pivot=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(anchor[0]+5,48,anchor[2]-7);
                Object box=cls("ekz").getMethod("a",cls("enu"),cls("jd"),cls("dmm")).invoke(element,manager,origin,rotate);
                Object center=cls("ejj").getMethod("g").invoke(box);
                int centerX=(Integer)cls("jd").getMethod("u").invoke(center)>>4,centerZ=(Integer)cls("jd").getMethod("w").invoke(center)>>4;
                // Processors see all template blocks before clipping and use
                // their own position random; piece random is only loot seeds.
                if(template!=null&&!((List<?>)palettes.get(template)).isEmpty()) {
                Object settings=makeSettings.invoke(element,rotate,box,cls("end").getEnumConstants()[1],false);
                Object palette=cls("enp").getMethod("a",List.class,cls("jd")).invoke(settings,palettes.get(template),origin);List<?> blocks=(List<?>)cls("ent$a").getMethod("a").invoke(palette);
                List<?> processed=(List<?>)cls("ent").getMethod("a",cls("ddl"),cls("jd"),cls("jd"),cls("enp"),List.class).invoke(null,world(mode,new HashMap<>(),new HashMap<>(),registries,overworld),origin,pivot,settings,blocks);
                List<String> processingRows=new ArrayList<>();
                for(Object info:processed) {
                    Object p=cls("ent$c").getMethod("a").invoke(info),state=cls("ent$c").getMethod("b").invoke(info),data=cls("ent$c").getMethod("c").invoke(info);
                    processingRows.add(cls("jd").getMethod("u").invoke(p)+" "+cls("jd").getMethod("v").invoke(p)+" "+cls("jd").getMethod("w").invoke(p)+" "+cls("dfy").getMethod("i",cls("dtc")).invoke(null,state)+" "+(data==null?"null":IglooProbeSupport.canonicalNBT(data)));
                }
                System.out.printf("jigsaw_processors %s %d %d %d %d %d %d %s%n",name,rotation,mode,anchor[0],anchor[1],anchor[2],processingRows.size(),hash(processingRows,false));
                }
                for(int outside=0;outside<2;outside++) {
                    int cx=centerX+outside,cz=centerZ;
                    Object clip=cls("ejj").getConstructor(int.class,int.class,int.class,int.class,int.class,int.class).newInstance(cx*16,-64,cz*16,cx*16+15,319,cz*16+15);
                    Map<String,Object> edits=new HashMap<>(),entities=new HashMap<>();Object world=world(mode,edits,entities,registries,overworld);
                    long seed=1234+rotation+mode+anchor[0];Object random=cls("dzx").getConstructor(cls("ayw")).newInstance(cls("dzz").getConstructor(long.class).newInstance(seed));
                    // Compare the complete original block stage separately
                    // from live entity instantiation/finalization.
                    if(template!=null) {
                        Object blockSettings=makeSettings.invoke(element,rotate,clip,cls("end").getEnumConstants()[1],false);
                        cls("enp").getMethod("a",boolean.class).invoke(blockSettings,true);
                        cls("ent").getMethod("a",cls("ddl"),cls("jd"),cls("jd"),cls("enp"),cls("ayw"),int.class).invoke(template,world,origin,pivot,blockSettings,random,18);
                    }else{
                        cls("ekz").getMethod("a",cls("enu"),cls("dds"),cls("ddq"),cls("duz"),cls("jd"),cls("jd"),cls("dmm"),cls("ejj"),cls("ayw"),cls("end"),boolean.class).invoke(element,manager,world,null,null,origin,pivot,rotate,clip,random,cls("end").getEnumConstants()[1],false);
                    }
                    List<String> rows=new ArrayList<>(),entityRows=new ArrayList<>();
                    for(var edit:edits.entrySet())rows.add(edit.getKey()+" "+cls("dfy").getMethod("i",cls("dtc")).invoke(null,edit.getValue()));
                    for(var block:entities.entrySet())entityRows.add(block.getKey()+" "+IglooProbeSupport.canonicalNBT(cls("dqh").getMethod("b",cls("jo$a")).invoke(block.getValue(),registries)));
                    System.out.printf("jigsaw_blocks %s %d %d %d %d %d %d %d %d %d %s %d %s %s%n",name,seed,rotation,mode,anchor[0],anchor[1],anchor[2],cx,cz,rows.size(),hash(rows,true),entities.size(),hash(entityRows,true),cls("ayw").getMethod("g").invoke(random));
                }
            }
        }
    }
}
