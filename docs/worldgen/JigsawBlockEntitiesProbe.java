package atlasprobe;

import com.google.gson.JsonElement;
import com.google.gson.JsonParser;
import java.io.ByteArrayOutputStream;
import java.io.DataOutputStream;
import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.Base64;
import java.util.List;
import java.util.Map;
import java.util.TreeMap;
import java.util.TreeSet;
import java.util.jar.JarFile;

/** Derives the original load/save normalization of built-in template block
 * entity payloads. Coordinates and per-placement loot seeds remain runtime. */
public final class JigsawBlockEntitiesProbe {
    static Class<?> cls(String name)throws Exception{return Class.forName(name);}
    static void templates(JsonElement value,TreeSet<String> names) {
        if(value.isJsonArray()){for(var child:value.getAsJsonArray())templates(child,names);return;}
        if(!value.isJsonObject())return;
        var object=value.getAsJsonObject();
        if(object.has("location"))names.add(object.get("location").getAsString().replace("minecraft:",""));
        for(var entry:object.entrySet())templates(entry.getValue(),names);
    }
    public static void main(String[] args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);
        RuinedPortalProbeSupport.bindTags();Object registries=IglooProbeSupport.registryAccess();TreeSet<String> names=new TreeSet<>();
        try(JarFile jar=new JarFile(new java.io.File(cls("ent").getProtectionDomain().getCodeSource().getLocation().toURI()))) {
            var entries=jar.entries();while(entries.hasMoreElements()) {
                var entry=entries.nextElement();if(entry.getName().startsWith("data/minecraft/worldgen/template_pool/")&&entry.getName().endsWith(".json"))templates(JsonParser.parseString(new String(jar.getInputStream(entry).readAllBytes(),StandardCharsets.UTF_8)),names);
            }
            names.removeIf(name->jar.getEntry("data/minecraft/structure/"+name+".nbt")==null);
        }
        Object manager=IglooProbeSupport.manager(names.toArray(String[]::new));
        var palettes=cls("ent").getDeclaredField("l");palettes.setAccessible(true);
        var blockRegistry=cls("lt").getField("e").get(null);
        TreeSet<String> entityBlocks=new TreeSet<>();
        for(Object block:(Iterable<?>)blockRegistry) {
            Object state=cls("dfy").getMethod("o").invoke(block);
            if((Boolean)cls("dtb$a").getMethod("t").invoke(state))entityBlocks.add(cls("jz").getMethod("b",Object.class).invoke(blockRegistry,block).toString());
        }
        Map<String,String> rows=new TreeMap<>();
        for(String name:names) {
            Object template=cls("enu").getMethod("a",cls("akr")).invoke(manager,cls("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name));
            for(Object palette:(List<?>)palettes.get(template))for(Object info:(List<?>)cls("ent$a").getMethod("a").invoke(palette)) {
                Object state=cls("ent$c").getMethod("b").invoke(info),input=cls("ent$c").getMethod("c").invoke(info),block=cls("dtb$a").getMethod("b").invoke(state);
                String id=cls("jz").getMethod("b",Object.class).invoke(blockRegistry,block).toString();
                if(input==null||id.equals("minecraft:jigsaw")||id.equals("minecraft:structure_block")||!(Boolean)cls("dtb$a").getMethod("t").invoke(state))continue;
                String inputHash=JigsawBlocksProbe.hash(new ArrayList<>(List.of(IglooProbeSupport.canonicalNBT(input))),false),key=id+" "+inputHash;
                Object copy=cls("ub").getMethod("i").invoke(input),p=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(0,0,0),entity=cls("diq").getMethod("a",cls("jd"),cls("dtc")).invoke(block,p,state);
                boolean randomized=cls("bqz").isInstance(entity);
                if(randomized)cls("ub").getMethod("a",String.class,long.class).invoke(copy,"LootTableSeed",0L);
                cls("dqh").getMethod("c",cls("ub"),cls("jo$a")).invoke(entity,copy,registries);
                Object saved=cls("dqh").getMethod("b",cls("jo$a")).invoke(entity,registries);
                for(String field:new String[]{"x","y","z"})cls("ub").getMethod("r",String.class).invoke(saved,field);
                ByteArrayOutputStream bytes=new ByteArrayOutputStream();cls("uo").getMethod("a",cls("uy"),java.io.DataOutput.class).invoke(null,saved,new DataOutputStream(bytes));
                String row=key+" "+(randomized?1:0)+" "+Base64.getEncoder().encodeToString(bytes.toByteArray());
                String prior=rows.putIfAbsent(key,row);
                if(prior!=null&&!prior.equals(row))throw new IllegalStateException("block state affects saved template NBT: "+key);
            }
        }
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));System.setErr(new PrintStream(new FileOutputStream(FileDescriptor.err),true));
        System.out.println("version 3955");for(String name:entityBlocks)System.out.println("entity_block "+name);for(String row:rows.values())System.out.println(row);
        System.err.println("original template block entity normalization cases="+rows.size());
    }
}
