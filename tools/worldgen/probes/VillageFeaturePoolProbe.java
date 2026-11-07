package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.*;
import java.util.function.Predicate;
import java.util.stream.Stream;

/** Actual FeaturePoolElement calls for every placed feature authored in villages. */
public final class VillageFeaturePoolProbe {
    static Class<?> cls(String name) throws Exception { return Class.forName(name); }
    static Map<String,Object> registry(Object lookup,String field) throws Exception {
        Object values=cls("jn$a").getMethod("b",cls("akq")).invoke(lookup,cls("lu").getField(field).get(null));
        Map<String,Object> result=new TreeMap<>();
        for(Object holder:((Stream<?>)cls("jo").getMethod("b").invoke(values)).toList()) {
            Object key=((Optional<?>)cls("jm").getMethod("e").invoke(holder)).orElseThrow();
            result.put(cls("akq").getMethod("a").invoke(key).toString(),cls("jm").getMethod("a").invoke(holder));
        }
        return result;
    }
    static void collect(com.google.gson.JsonObject element,Set<String> names) {
        if(element.get("element_type").getAsString().equals("minecraft:feature_pool_element"))names.add(element.get("feature").getAsString());
        if(element.has("elements"))for(var child:element.getAsJsonArray("elements"))collect(child.getAsJsonObject(),names);
    }
    public static void main(String[] args) throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);
        cls("ab").getMethod("a",cls("ae")).invoke(null,version);
        cls("akt").getMethod("a").invoke(null);
        TreeFeatureProbe.bindTreeTags();
        Object registries=cls("pd").getMethod("a").invoke(null);
        Object lookup=cls("jo$a").getMethod("b").invoke(registries);
        Map<String,Object> placed=registry(lookup,"aQ");
        Map<String,Object> biomes=registry(lookup,"aF");
        Object settingsLookup=cls("jn$a").getMethod("b",cls("akq")).invoke(lookup,cls("lu").getField("aO").get(null));
        Object settings=cls("jn").getMethod("b",cls("akq")).invoke(settingsLookup,cls("dzd").getField("c").get(null));
        Object biomeHolder=cls("jm").getMethod("a",Object.class).invoke(null,biomes.get("minecraft:plains"));
        Object biomeSource=cls("deh").getConstructor(cls("jm")).newInstance(biomeHolder);
        Object generator=cls("dzb").getConstructor(cls("dea"),cls("jm")).newInstance(biomeSource,settings);
        Set<String> names=new TreeSet<>();
        try(var jar=new java.util.jar.JarFile(new File(cls("ent").getProtectionDomain().getCodeSource().getLocation().toURI()))) {
            var entries=jar.entries();
            while(entries.hasMoreElements()) {
                String name=entries.nextElement().getName();
                if(!name.startsWith("data/minecraft/worldgen/template_pool/village/")||!name.endsWith(".json"))continue;
                var pool=com.google.gson.JsonParser.parseString(new String(jar.getInputStream(jar.getJarEntry(name)).readAllBytes(),StandardCharsets.UTF_8)).getAsJsonObject();
                for(var entry:pool.getAsJsonArray("elements"))collect(entry.getAsJsonObject().getAsJsonObject("element"),names);
            }
        }
        if(names.size()!=13)throw new IllegalStateException("expected 13 village placed features: "+names);
        Map<String,Object> blocks=new HashMap<>();
        Object blockRegistry=cls("lt").getField("e").get(null);
        for(String name:new String[]{"air","grass_block","dirt","stone","sand","water","oak_log","oak_leaves","powder_snow","snow_block"}) {
            Object block=cls("jz").getMethod("a",cls("akr")).invoke(blockRegistry,cls("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name));
            blocks.put(name,cls("dfy").getMethod("o").invoke(block));
        }
        blocks.put("flowing_water",cls("dte").getMethod("a",cls("duf"),Comparable.class).invoke(blocks.get("water"),cls("dts").getField("aP").get(null),1));
        Class<?> pos=cls("jd"),state=cls("dtc");
        Method x=pos.getMethod("u"),y=pos.getMethod("v"),z=pos.getMethod("w");
        Method fluid=cls("dtb$a").getMethod("u"),stateID=cls("dfy").getMethod("i",state);
        Object biomeRegistry=cls("ju").getConstructor(cls("akq"),com.mojang.serialization.Lifecycle.class).newInstance(cls("lu").getField("aF").get(null),com.mojang.serialization.Lifecycle.stable());
        Object biomeLookup=cls("jn$a").getMethod("b",cls("akq")).invoke(lookup,cls("lu").getField("aF").get(null));
        for(Object holder:((Stream<?>)cls("jo").getMethod("b").invoke(biomeLookup)).toList())cls("jz").getMethod("a",cls("jz"),cls("akq"),Object.class).invoke(null,biomeRegistry,cls("jm$c").getMethod("h").invoke(holder),cls("jm").getMethod("a").invoke(holder));
        Object worldHeight=cls("dcy").getMethod("e",int.class,int.class).invoke(null,-64,384);
        var chunkConstructor=cls("dvs").getConstructor(cls("dcd"),cls("dvv"),cls("dcy"),cls("jz"),cls("eab"));
        var elementConstructor=cls("ekt").getDeclaredConstructor(cls("jm"),cls("elb$a"));elementConstructor.setAccessible(true);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        System.setErr(new PrintStream(new FileOutputStream(FileDescriptor.err),true));
        for(String name:names)for(long seed:new long[]{0,1,-1,42,123456789,987654321})for(int mode=0;mode<10;mode++)for(int rotation=0;rotation<4;rotation++) {
            final int scenario=mode;
            int[] origin=mode==1?new int[]{-1,41,-16}:mode==2?new int[]{15,41,15}:mode==8?new int[]{8,319,8}:mode==9?new int[]{8,-64,8}:new int[]{8,41,8};
            Map<String,Object> changed=new HashMap<>(),postChunks=new HashMap<>();
            java.util.function.Function<int[],Object> read=p->{
                if(p[1]<-64||p[1]>=320)return blocks.get("air");
                String key=p[0]+" "+p[1]+" "+p[2];if(changed.containsKey(key))return changed.get(key);
                int height=40;
                if(scenario==7)height+=Math.floorDiv(p[0]-origin[0],4);
                if(p[1]==height)return blocks.get(scenario==3?"sand":scenario==4?"stone":"grass_block");
                if(p[1]<height)return blocks.get("dirt");
                if(scenario==5&&p[1]<=43)return blocks.get("water");
                if(scenario==6&&p[1]<=43)return blocks.get("flowing_water");
                return blocks.get("air");
            };
            Object world=Proxy.newProxyInstance(cls("dds").getClassLoader(),new Class<?>[]{cls("dds")},(proxy,method,values)->{
                String m=method.getName();
                if(m.equals("I_"))return -64;if(m.equals("J_"))return 384;if(m.equals("am"))return 320;if(m.equals("f_"))return true;
                if(m.equals("toString"))return "atlas-village-feature-world";
                if(m.equals("y")) {
                    int cx=(int)x.invoke(values[0])>>4,cz=(int)z.invoke(values[0])>>4;String key=cx+" "+cz;
                    if(!postChunks.containsKey(key))postChunks.put(key,chunkConstructor.newInstance(cls("dcd").getConstructor(int.class,int.class).newInstance(cx,cz),cls("dvv").getField("a").get(null),worldHeight,biomeRegistry,null));
                    return postChunks.get(key);
                }
                if(values!=null&&values.length>0&&pos.isInstance(values[0])) {
                    int[] p={(int)x.invoke(values[0]),(int)y.invoke(values[0]),(int)z.invoke(values[0])};Object s=read.apply(p);
                    if(m.equals("a_"))return s;if(m.equals("b_"))return fluid.invoke(s);
                    if(m.equals("a")&&values.length==2&&cls("dqj").isInstance(values[1]))return Optional.empty();
                    if(m.equals("a")&&values.length==2&&values[1] instanceof Predicate<?>)return ((Predicate<Object>)values[1]).test(s);
                    if(m.equals("b")&&values.length==2&&values[1] instanceof Predicate<?>)return ((Predicate<Object>)values[1]).test(fluid.invoke(s));
                    if(m.equals("a")&&values.length==3&&state.isInstance(values[1])) {
                        if(p[1]<-64||p[1]>=320)return false;changed.put(p[0]+" "+p[1]+" "+p[2],values[1]);return true;
                    }
                    if(method.getReturnType()==void.class)return null;
                }
                if(method.isDefault())return InvocationHandler.invokeDefault(proxy,method,values);
                throw new UnsupportedOperationException(method.toString());
            });
            Object holder=cls("jm").getMethod("a",Object.class).invoke(null,placed.get(name));
            Object element=elementConstructor.newInstance(holder,cls("elb$a").getEnumConstants()[0]);
            Object random=cls("dzx").getConstructor(cls("ayw")).newInstance(cls("dzz").getConstructor(long.class).newInstance(seed));
            Object start=pos.getConstructor(int.class,int.class,int.class).newInstance(origin[0],origin[1],origin[2]);
            Object clip=cls("ejj").getConstructor(int.class,int.class,int.class,int.class,int.class,int.class).newInstance(256,-64,256,271,319,271);
            Object pivot=pos.getConstructor(int.class,int.class,int.class).newInstance(999,999,999);
            boolean success=(boolean)cls("ekt").getMethod("a",cls("enu"),cls("dds"),cls("ddq"),cls("duz"),pos,pos,cls("dmm"),cls("ejj"),cls("ayw"),cls("end"),boolean.class).invoke(element,null,world,null,generator,start,pivot,cls("dmm").getEnumConstants()[rotation],clip,random,cls("end").getEnumConstants()[0],false);
            List<String> output=new ArrayList<>();int radius=0;
            for(var edit:changed.entrySet()) {
                output.add(edit.getKey()+" "+stateID.invoke(null,edit.getValue()));
                String[] p=edit.getKey().split(" ");radius=Math.max(radius,Math.max(Math.abs(Integer.parseInt(p[0])-origin[0]),Math.abs(Integer.parseInt(p[2])-origin[2])));
            }
            Set<String> marked=new HashSet<>();
            for(var entry:postChunks.entrySet()) {
                String[] coords=entry.getKey().split(" ");int cx=Integer.parseInt(coords[0]),cz=Integer.parseInt(coords[1]);
                Object[] sections=(Object[])cls("duy").getMethod("n").invoke(entry.getValue());
                for(int section=0;section<sections.length;section++)if(sections[section]!=null)for(Object packed:(Iterable<?>)sections[section]) {
                    int v=((Short)packed)&65535;marked.add((cx*16+(v&15))+" "+(-64+section*16+((v>>4)&15))+" "+(cz*16+((v>>8)&15)));
                }
            }
            Collections.sort(output);
            String hash=HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(String.join("\n",output).getBytes(StandardCharsets.UTF_8)));
            List<String> markRows=new ArrayList<>(marked);Collections.sort(markRows);
            String markHash=HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(String.join("\n",markRows).getBytes(StandardCharsets.UTF_8)));
            System.out.printf("village_feature %s %d %d %d %d %d %d %d %d %s %d %d %d %s%n",name,seed,mode,origin[0],origin[1],origin[2],success?1:0,changed.size(),cls("dzx").getMethod("g").invoke(random),hash,marked.size(),rotation,radius,markHash);
        }
    }
}
