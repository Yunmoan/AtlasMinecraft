package atlasprobe;

import com.google.gson.*;
import io.netty.buffer.*;
import java.io.*;
import java.lang.reflect.*;
import java.nio.file.*;
import java.util.*;

/** Development-only wire fixtures emitted by the official 1.21.1 codec. */
public final class ItemStackProbe {
    static Class<?> stack, compound, component;
    static Object lookup, access, codec;
    static JsonArray fixtures = new JsonArray();
    static Set<String> seen = new HashSet<>();

    static Object parse(String snbt) throws Exception {
        Object tag = Class.forName("uz").getMethod("a", String.class).invoke(null, snbt);
        Object item = stack.getMethod("a", Class.forName("jo$a"), compound).invoke(null, lookup, tag);
        if ((boolean)stack.getMethod("e").invoke(item)) throw new IllegalStateException("invalid sample " + snbt);
        return item;
    }

    static void emit(String name, Object item) throws Exception {
        ByteBuf raw = Unpooled.buffer();
        try {
            Object buffer = Class.forName("wk").getConstructor(ByteBuf.class, Class.forName("ka")).newInstance(raw, access);
            Class.forName("yz").getMethod("encode", Object.class, Object.class).invoke(codec, buffer, item);
            byte[] bytes = new byte[raw.readableBytes()];
            raw.getBytes(0, bytes);
            JsonObject entry = new JsonObject();
            entry.addProperty("name", name);
            entry.addProperty("hex", HexFormat.of().formatHex(bytes));
            fixtures.add(entry);
        } finally { raw.release(); }
    }

    public static void main(String[] args) throws Exception {
        Object version=Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a",Class.forName("ae")).invoke(null,version);
        Class.forName("akt").getMethod("a").invoke(null);
        lookup=Class.forName("pd").getMethod("a").invoke(null);
        access=Class.forName("ka").getMethod("a",Class.forName("jz")).invoke(null,Class.forName("lt").getField("aA").get(null));
        stack=Class.forName("cuq"); compound=Class.forName("ub"); component=Class.forName("kp");
        codec=stack.getField("h").get(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Object registry=Class.forName("lt").getField("g").get(null);
        Object componentRegistry=Class.forName("lt").getField("aq").get(null);
        for(Object itemType:(Iterable<?>)registry) {
            Object source=stack.getConstructor(Class.forName("dcv")).newInstance(itemType);
            Object components=stack.getMethod("a").invoke(source);
            for(Object typed:(Iterable<?>)components) {
                Object type=Class.forName("ks").getMethod("a").invoke(typed);
                String name=Class.forName("jz").getMethod("b",Object.class).invoke(componentRegistry,type).toString();
                if(!seen.add(name)) continue;
                Object target=parse("{id:'minecraft:stone',count:1}");
                Object value=Class.forName("ks").getMethod("b").invoke(typed);
                stack.getMethod("b",component,Object.class).invoke(target,type,value);
                try { emit("default_"+name,target); }
                catch(InvocationTargetException error) {
                    // This access contains built-in registries; dynamic holder
                    // IDs must be tested with the server configuration registry.
                    System.out.println("skip dynamic sample "+name+": "+error.getCause());
                }
            }
        }
        String[] samples = {
            "{id:'minecraft:potion',count:1,components:{'minecraft:potion_contents':{potion:'minecraft:strong_healing'}}}",
            "{id:'minecraft:potion',count:1,components:{'minecraft:potion_contents':{custom_color:1193046,custom_effects:[{id:'minecraft:speed',amplifier:2,duration:600,ambient:true,show_particles:false,show_icon:true,hidden_effect:{id:'minecraft:speed',duration:1200}}]}}}",
            "{id:'minecraft:apple',count:3,components:{'minecraft:food':{nutrition:4,saturation:0.3f,can_always_eat:true,eat_seconds:0.8f,effects:[{effect:{id:'minecraft:speed',duration:200},probability:0.5f}]}}}",
            "{id:'minecraft:apple',count:1,components:{'minecraft:food':{nutrition:4,saturation:0.3f,using_converts_to:{id:'minecraft:bowl',count:1},effects:[{effect:{id:'minecraft:regeneration',amplifier:1,duration:100},probability:1.0f}]}}}",
            "{id:'minecraft:stone',count:1,components:{'minecraft:max_stack_size':37,'minecraft:custom_model_data':12,'minecraft:enchantment_glint_override':true,'minecraft:unbreakable':{show_in_tooltip:false},'minecraft:custom_name':'\"Wire probe\"'}}",
            "{id:'minecraft:stone',count:1,components:{'minecraft:can_place_on':{predicates:[{blocks:'minecraft:stone',state:{axis:'x'}}],show_in_tooltip:false}}}",
            "{id:'minecraft:stone',count:1,components:{'minecraft:profile':{name:'Atlas',properties:[{name:'textures',value:'probe',signature:'signature'}]}}}",
            "{id:'minecraft:stone',count:1,components:{'minecraft:container':[{slot:2,item:{id:'minecraft:potion',count:1,components:{'minecraft:potion_contents':'minecraft:healing'}}}]}}"
        };
        for(int i=0;i<samples.length;i++) emit("custom_"+i,parse(samples[i]));
        Path output=Path.of(args[0]);
        Files.createDirectories(output.getParent());
        Files.writeString(output,new GsonBuilder().setPrettyPrinting().create().toJson(fixtures)+"\n");
        JsonObject registries=new JsonObject();
        for(String[] spec:new String[][]{{"mob_effect","d"},{"potion","h"}}) {
            Object reg=Class.forName("lt").getField(spec[1]).get(null);
            TreeMap<Integer,String> names=new TreeMap<>();
            for(Object value:(Iterable<?>)reg) {
                int id=(int)Class.forName("jz").getMethod("a",Object.class).invoke(reg,value);
                names.put(id,Class.forName("jz").getMethod("b",Object.class).invoke(reg,value).toString());
            }
            registries.add("minecraft:"+spec[0],new Gson().toJsonTree(names.values()));
        }
        Files.writeString(Path.of(args[1]),new Gson().toJson(registries)+"\n");
        System.out.println("emitted "+fixtures.size()+" official item wire samples");
    }
}
