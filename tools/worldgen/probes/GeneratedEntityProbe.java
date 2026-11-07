package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.util.*;
import java.util.function.Supplier;
import java.util.stream.Stream;

/** Captures actual Entity/Mob save NBT and the base Mob.finalizeSpawn stream. */
public final class GeneratedEntityProbe {
    static void field(Object receiver,String cls,String name,Object value)throws Exception {
        Field f=Class.forName(cls).getDeclaredField(name);f.setAccessible(true);f.set(receiver,value);
    }
    static void print(String label,String id,long seed,Object tag)throws Exception {
        var bytes=io.netty.buffer.Unpooled.buffer();
        Object buffer=Class.forName("vw").getConstructor(io.netty.buffer.ByteBuf.class).newInstance(bytes);
        Class.forName("vw").getMethod("a",Class.forName("uy")).invoke(buffer,tag);
        byte[] data=new byte[bytes.readableBytes()];bytes.readBytes(data);bytes.release();
        System.out.printf("%s %s %d %s%n",label,id,seed,HexFormat.of().formatHex(data));
    }
    static void metadata(String label,String id,long seed,Object entity,Object access)throws Exception {
        Object data=Class.forName("bsr").getMethod("ar").invoke(entity);
        List<?> values=(List<?>)Class.forName("aka").getMethod("c").invoke(data);if(values==null)values=List.of();
        Object packet=Class.forName("aez").getConstructor(int.class,List.class).newInstance(77,values);
        var bytes=io.netty.buffer.Unpooled.buffer();Object buffer=Class.forName("wk").getConstructor(io.netty.buffer.ByteBuf.class,Class.forName("ka")).newInstance(bytes,access);
        Class.forName("yx").getMethod("encode",Object.class,Object.class).invoke(Class.forName("aez").getField("a").get(null),buffer,packet);
        byte[] output=new byte[bytes.readableBytes()];bytes.readBytes(output);bytes.release();
        System.out.printf("%s %s %d %s%n",label,id,seed,HexFormat.of().formatHex(output));
    }
    public static void main(String[] args)throws Exception {
        Object version=Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a",Class.forName("ae")).invoke(null,version);
        Class.forName("akt").getMethod("a").invoke(null);
        Object registries=Class.forName("pd").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        System.setErr(new PrintStream(new FileOutputStream(FileDescriptor.err),true));
        Object lookup=Class.forName("jo$a").getMethod("b").invoke(registries);
        Object dimLookup=Class.forName("jn$a").getMethod("b",Class.forName("akq")).invoke(lookup,Class.forName("lu").getField("aK").get(null));
        Object dimension=null;
        for(Object holder:((Stream<?>)Class.forName("jo").getMethod("b").invoke(dimLookup)).toList()) {
            Object key=((Optional<?>)Class.forName("jm").getMethod("e").invoke(holder)).orElseThrow();
            if(Class.forName("akq").getMethod("a").invoke(key).toString().equals("minecraft:overworld"))dimension=holder;
        }
        Field unsafeField=sun.misc.Unsafe.class.getDeclaredField("theUnsafe");unsafeField.setAccessible(true);
        var unsafe=(sun.misc.Unsafe)unsafeField.get(null);
        Object server=unsafe.allocateInstance(Class.forName("apn")),level=unsafe.allocateInstance(Class.forName("aqu"));
        Object worldData=Proxy.newProxyInstance(Class.forName("erl").getClassLoader(),new Class<?>[]{Class.forName("erl")},(proxy,method,values)->{if(method.getName().equals("K"))return Class.forName("cpn").getField("g").get(null);throw new UnsupportedOperationException(method.toString());});
        field(server,"net.minecraft.server.MinecraftServer","j",worldData);
        field(level,"aqu","J",server);
        field(level,"dcw","D",dimension);
        field(level,"dcw","I",Class.forName("ka").getMethod("a",Class.forName("jz")).invoke(null,Class.forName("lt").getField("a").get(null)));
        field(level,"dcw","E",(Supplier<Object>)()->{try{return Class.forName("bnc").getField("a").get(null);}catch(Exception e){throw new RuntimeException(e);}});
        Object levelData=Proxy.newProxyInstance(Class.forName("erm").getClassLoader(),new Class<?>[]{Class.forName("erm")},(proxy,method,values)->{
            if(method.getReturnType()==long.class)return 0L;
            if(method.getReturnType()==boolean.class)return false;
            if(method.getReturnType()==int.class)return 0;
            throw new UnsupportedOperationException(method.toString());
        });
        field(level,"dcw","A",levelData);
        Object pieces=Class.forName("ekg").getConstructor(List.class).newInstance(List.of(unsafe.allocateInstance(Class.forName("eml"))));
        Object hut=Class.forName("ejz").getConstructor(Class.forName("ejr"),Class.forName("dcd"),int.class,Class.forName("ekg")).newInstance(null,Class.forName("dcd").getConstructor(int.class,int.class).newInstance(0,0),0,pieces);
        class HelperLoader extends ClassLoader {
            HelperLoader(){super(GeneratedEntityProbe.class.getClassLoader());}
            Class<?> define(byte[] bytes){return defineClass("AtlasEntityStructureManager",bytes,0,bytes.length);}
        }
        Class<?> structureManager=new HelperLoader().define(java.nio.file.Files.readAllBytes(java.nio.file.Path.of("artifacts/worldgen-probes/AtlasEntityStructureManager.class")));
        field(level,"aqu","ad",structureManager.getConstructor(Class.forName("ejz")).newInstance(hut));
        Object catRegistry=Class.forName("lt").getField("ai").get(null);
        Map<Object,List<Object>> catTags=new HashMap<>();
        for(String name:new String[]{"default_spawns","full_moon_spawns"}) {
            String text;try(var input=GeneratedEntityProbe.class.getClassLoader().getResourceAsStream("data/minecraft/tags/cat_variant/"+name+".json")){text=new String(input.readAllBytes(),java.nio.charset.StandardCharsets.UTF_8);}
            var matches=java.util.regex.Pattern.compile("\"minecraft:[^\"]+\"").matcher(text);List<Object> values=new ArrayList<>();
            if(name.equals("full_moon_spawns"))values.addAll(catTags.get(Class.forName("awf").getField("a").get(null)));
            while(matches.find()){String resource=matches.group().replace("\"","");if(resource.equals("minecraft:default_spawns"))continue;Object key=Class.forName("akr").getMethod("a",String.class).invoke(null,resource);Object variant=Class.forName("jz").getMethod("a",Class.forName("akr")).invoke(catRegistry,key);values.add(Class.forName("jz").getMethod("e",Object.class).invoke(catRegistry,variant));}
            catTags.put(Class.forName("awf").getField(name.equals("default_spawns")?"a":"b").get(null),values);
        }
        Class.forName("jz").getMethod("a",Map.class).invoke(catRegistry,catTags);
        Object metadataAccess=Class.forName("ka$c").getConstructor(List.class).newInstance(List.of(catRegistry,Class.forName("lt").getField("w").get(null),Class.forName("lt").getField("x").get(null)));
        for(String[] entry:new String[][]{{"minecraft:cat_variant","ai"},{"minecraft:villager_type","w"},{"minecraft:villager_profession","x"}}) {
            Object values=Class.forName("lt").getField(entry[1]).get(null);
            for(Object value:(Iterable<?>)values)System.out.printf("entity_registry %s %s %s%n",entry[0],Class.forName("jz").getMethod("b",Object.class).invoke(values,value),Class.forName("jz").getMethod("a",Object.class).invoke(values,value));
        }
        Class<?> entity=Class.forName("bsr"),world=Class.forName("dcw"),type=Class.forName("bsx"),tag=Class.forName("ub");
        Object registry=Class.forName("lt").getField("f").get(null);
        for(String id:new String[]{"minecraft:end_crystal","minecraft:witch","minecraft:cat","minecraft:villager","minecraft:zombie_villager"}) {
            Object key=Class.forName("akr").getMethod("a",String.class).invoke(null,id);
            Object entityType=Class.forName("jz").getMethod("a",Class.forName("akr")).invoke(registry,key);
            String cls=id.endsWith("witch")?"ckt":id.endsWith("cat")?"cfh":id.endsWith("zombie_villager")?"ckx":id.endsWith("villager")?"cmk":"chz";
            for(long seed:new long[]{0,1,-1,42,1234,987654321})for(boolean fullMoon:new boolean[]{false,true}) {
                if(fullMoon&&!cls.equals("cfh"))continue;
                Object value=Class.forName(cls).getConstructor(type,world).newInstance(entityType,level);
                entity.getMethod("b",double.class,double.class,double.class,float.class,float.class).invoke(value,-15.5,70.0,16.5,0f,0f);
                entity.getMethod("a_",UUID.class).invoke(value,new UUID(0,0));
                if(!cls.equals("chz")) {
                    Class.forName("btp").getMethod("fR").invoke(value);
                    if(cls.equals("cmk")||cls.equals("ckx")) {
                        Object villagerData=Class.forName("cml").getConstructor(Class.forName("cmp"),Class.forName("cmn"),int.class).newInstance(Class.forName("cmp").getField("c").get(null),Class.forName("cmn").getField("b").get(null),1);
                        Class.forName(cls).getMethod("a",Class.forName("cml")).invoke(value,villagerData);
                    }
                    if(cls.equals("cfh")) {
                        Object allBlack=Class.forName("jz").getMethod("g",Class.forName("akq")).invoke(catRegistry,Class.forName("cfi").getField("l").get(null));
                        Class.forName("cfh").getMethod("i",Class.forName("jm")).invoke(value,allBlack);
                    }
                }
                Object saved=tag.getConstructor().newInstance();
                if(!(boolean)entity.getMethod("e",tag).invoke(value,saved))throw new IllegalStateException("entity did not save");
                if(!fullMoon){print("entity_default",id,seed,saved);metadata("entity_default_metadata",id,seed,value,metadataAccess);}
                if(seed==0 && !fullMoon && !cls.equals("chz"))for(int scenario=0;scenario<8;scenario++) {
                    Object changed=Class.forName(cls).getConstructor(type,world).newInstance(entityType,level);
                    entity.getMethod("g",tag).invoke(changed,saved);
                    Object altered=tag.getConstructor().newInstance();entity.getMethod("e",tag).invoke(changed,altered);
                    tag.getMethod("a",String.class,float.class).invoke(altered,"Health",scenario==0?1f:3.5f);
                    tag.getMethod("a",String.class,boolean.class).invoke(altered,"NoAI",(scenario&1)!=0);
                    tag.getMethod("a",String.class,boolean.class).invoke(altered,"LeftHanded",(scenario&2)!=0);
                    if(cls.equals("cfh")||cls.equals("cmk"))tag.getMethod("a",String.class,int.class).invoke(altered,"Age",(scenario&4)!=0?-24000:0);
                    if(cls.equals("ckx"))tag.getMethod("a",String.class,boolean.class).invoke(altered,"IsBaby",(scenario&4)!=0);
                    if(cls.equals("cfh")) {
                        tag.getMethod("a",String.class,byte.class).invoke(altered,"CollarColor",(byte)scenario);
                        tag.getMethod("a",String.class,String.class).invoke(altered,"variant",scenario==0?"minecraft:tabby":"minecraft:all_black");
                    }
                    if(cls.equals("cmk")||cls.equals("ckx")) {
                        Object villagerData=Class.forName("cml").getConstructor(Class.forName("cmp"),Class.forName("cmn"),int.class).newInstance(Class.forName("cmp").getField("e").get(null),Class.forName("jz").getMethod("a",Class.forName("akr")).invoke(Class.forName("lt").getField("x").get(null),Class.forName("akr").getMethod("a",String.class).invoke(null,"minecraft:farmer")),2);
                        Class.forName(cls).getMethod("a",Class.forName("cml")).invoke(changed,villagerData);
                        Object villageTag=Class.forName("cml").getField("c").get(null);
                        Object nbt=((com.mojang.serialization.Codec<Object>)villageTag).encodeStart((com.mojang.serialization.DynamicOps<Object>)Class.forName("up").getField("a").get(null),villagerData).getOrThrow();
                        tag.getMethod("a",String.class,Class.forName("uy")).invoke(altered,"VillagerData",nbt);
                    }
                    entity.getMethod("g",tag).invoke(changed,altered);
                    Object actual=tag.getConstructor().newInstance();entity.getMethod("e",tag).invoke(changed,actual);
                    print("entity_changed",id,scenario,actual);metadata("entity_changed_metadata",id,scenario,changed,metadataAccess);
                }
                if(cls.equals("ckt")||cls.equals("cfh")) {
                    Object random=Class.forName("dzz").getConstructor(long.class).newInstance(seed);
                    Object spawnWorld=Proxy.newProxyInstance(Class.forName("ddl").getClassLoader(),new Class<?>[]{Class.forName("ddl")},(proxy,method,values)->{if(method.getName().equals("E_"))return random;if(method.getName().equals("aq"))return fullMoon?1f:0f;if(method.getName().equals("E"))return level;throw new UnsupportedOperationException(method.toString());});
                    Class.forName(cls).getMethod("a",Class.forName("ddl"),Class.forName("bqp"),Class.forName("btr"),Class.forName("buh")).invoke(value,spawnWorld,null,Class.forName("btr").getField("d").get(null),null);
                    saved=tag.getConstructor().newInstance();entity.getMethod("e",tag).invoke(value,saved);
                    print(fullMoon?"entity_finalized_fullmoon":"entity_finalized",id,seed,saved);
                    metadata(fullMoon?"entity_finalized_fullmoon_metadata":"entity_finalized_metadata",id,seed,value,metadataAccess);
                    System.out.printf("entity_next %s %d %d%n",id,seed,Class.forName("dzz").getMethod("g").invoke(random));
                }
            }
        }
    }
}
