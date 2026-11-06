package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Stream;

/** Calls the original FossilFeature and StructureTemplate implementation.
 * Unsafe is only used to provide the server-to-template-manager references;
 * all templates, processors and configured features are loaded by the JAR.
 */
public final class FossilFeatureProbe {
    static Class<?> cls(String name) throws Exception { return Class.forName(name); }
    static Object call(Object target, String name, Class<?>[] types, Object... args) throws Exception {
        return target.getClass().getMethod(name, types).invoke(target, args);
    }
    static Object key(String name) throws Exception { return cls("akr").getMethod("a", String.class).invoke(null, name); }
    static void field(Object obj, Class<?> owner, String name, Object value) throws Exception {
        Field field = owner.getDeclaredField(name); field.setAccessible(true); field.set(obj, value);
    }

    static Object serverLevel(Object registry) throws Exception {
        Field f = sun.misc.Unsafe.class.getDeclaredField("theUnsafe"); f.setAccessible(true);
        sun.misc.Unsafe unsafe = (sun.misc.Unsafe)f.get(null);
        Object manager = unsafe.allocateInstance(cls("enu"));
        Map<Object, Optional<Object>> templates = new ConcurrentHashMap<>();
        Object lookup = cls("jz").getMethod("q").invoke(registry);
        for (String kind : new String[]{"spine", "skull"}) for (int i=1; i<=4; i++) for (String suffix : new String[]{"", "_coal"}) {
            String name = "fossil/"+kind+"_"+i+suffix;
            Object template = cls("ent").getConstructor().newInstance();
            try (InputStream input = FossilFeatureProbe.class.getClassLoader().getResourceAsStream("data/minecraft/structure/"+name+".nbt")) {
                if (input == null) throw new IllegalStateException("missing template "+name);
                Object nbt = cls("uo").getMethod("a", InputStream.class, cls("uk"))
                    .invoke(null, input, cls("uk").getMethod("a").invoke(null));
                cls("ent").getMethod("a", cls("jn"), cls("ub")).invoke(template, lookup, nbt);
            }
            templates.put(key("minecraft:"+name), Optional.of(template));
        }
        field(manager, cls("enu"), "f", templates);
        Object server = unsafe.allocateInstance(cls("apn"));
        field(server, cls("net.minecraft.server.MinecraftServer"), "aA", manager);
        Object level = unsafe.allocateInstance(cls("aqu"));
        field(level, cls("aqu"), "J", server);
        return level;
    }

    static int floor(int x, int z, int mode) {
        if (mode==1) return 38+Math.floorMod(x+z,5);
        if (mode==7) return -65;
        if (mode==8) return 318;
        if (mode==9) return -60;
        return 40;
    }
    static String base(int x, int y, int z, int mode, int[] origin) {
        if (y < -64 || y >= 320 || y > floor(x,z,mode)) return "air";
        if (mode==4) return "bedrock";
        if (mode==3 && Math.floorMod(x+2*y+3*z,5)==0) return "bedrock";
        if (mode==5 && y < 35 && Math.floorMod(x+z,3)==0) return "water";
        if (mode==6 && y < 35 && Math.floorMod(x+z,3)==0) return "lava";
        if (mode==10 && y < 35 && x >= origin[0]) return "air";
        if (mode==11 && y < 35 && (x >= origin[0] || z >= origin[2])) return "air";
        if (mode==12 && y < 35 && Math.floorMod(x+z,3)==0) return "flowing_water";
        if (mode==13 && y < 35 && Math.floorMod(x+z,3)==0) return "waterlogged_slab";
        if (mode==14) return "deepslate";
        if (mode==15 && Math.floorMod(x+2*y+3*z,5)==0) return "end_portal_frame";
        return "stone";
    }
    static int[] origin(int mode) {
        if (mode==1) return new int[]{-1,64,-16};
        if (mode==2) return new int[]{15,64,15};
        if (mode==8) return new int[]{8,320,8};
        if (mode==9) return new int[]{8,-64,8};
        if (mode==14) return new int[]{29999968,-8,29999968};
        if (mode==15) return new int[]{-29999968,64,-29999968};
        return new int[]{8,64,8};
    }

    public static void main(String[] args) throws Exception {
        Object version = cls("t").getMethod("a").invoke(null);
        cls("ab").getMethod("a",cls("ae")).invoke(null,version);
        cls("akt").getMethod("a").invoke(null);
        TreeFeatureProbe.bindTreeTags();
        Object registry = cls("lt").getField("e").get(null);
        Map<Object,List<Object>> tags = new HashMap<>();
        for (Object tag : ((Stream<?>)cls("jz").getMethod("k").invoke(registry)).toList()) tags.put(tag,List.of());
        List<Object> protectedBlocks = new ArrayList<>();
        for (String name : TreeFeatureProbe.tagBlocks("features_cannot_replace")) {
            Object block = cls("jz").getMethod("a",cls("akr")).invoke(registry,key(name));
            protectedBlocks.add(cls("jz").getMethod("e",Object.class).invoke(registry,block));
        }
        tags.put(cls("awe").getField("bP").get(null),protectedBlocks);
        cls("jz").getMethod("a",Map.class).invoke(registry,tags);
        Object registries = cls("pd").getMethod("a").invoke(null);
        Object lookup = cls("jo$a").getMethod("b").invoke(registries);
        Object featureLookup = cls("jn$a").getMethod("b",cls("akq")).invoke(lookup,cls("lu").getField("aI").get(null));
        Map<String,Object> configured = new HashMap<>();
        for (Object holder : ((Stream<?>)cls("jo").getMethod("b").invoke(featureLookup)).toList()) {
            Object k = ((Optional<?>)cls("jm").getMethod("e").invoke(holder)).orElseThrow();
            configured.put(cls("akq").getMethod("a").invoke(k).toString(),cls("jm").getMethod("a").invoke(holder));
        }
        Object level = serverLevel(registry);
        Object dimensions = cls("jn$a").getMethod("b",cls("akq")).invoke(lookup,cls("lu").getField("aK").get(null));
        Object dimensionHolder = cls("jn").getMethod("b",cls("akq")).invoke(dimensions,cls("dww").getField("a").get(null));
        Object dimension = cls("jm").getMethod("a").invoke(dimensionHolder);
        Class<?> pos = cls("jd"), state = cls("dtc");
        Method x=pos.getMethod("u"), y=pos.getMethod("v"), z=pos.getMethod("w");
        Method stateID=cls("dfy").getMethod("i",state), fluid=cls("dtb$a").getMethod("u");
        Map<String,Object> states=new HashMap<>();
        for(String name:new String[]{"air","stone","bedrock","water","lava","deepslate","end_portal_frame","oak_slab"}) {
            Object block=cls("jz").getMethod("a",cls("akr")).invoke(registry,key("minecraft:"+name));
            states.put(name,cls("dfy").getMethod("o").invoke(block));
        }
        states.put("flowing_water",cls("dte").getMethod("a",cls("duf"),Comparable.class)
            .invoke(states.get("water"),cls("dts").getField("aP").get(null),Integer.valueOf(1)));
        states.put("waterlogged_slab",cls("dte").getMethod("a",cls("duf"),Comparable.class)
            .invoke(states.get("oak_slab"),cls("dts").getField("C").get(null),Boolean.TRUE));
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        for(String name:new String[]{"fossil_coal","fossil_diamonds"}) for(long seed=0;seed<128;seed++) for(int mode=0;mode<16;mode++) {
            int scenario=mode; int[] start=origin(mode);
            Map<String,Object> changed=new HashMap<>();
            Object world=Proxy.newProxyInstance(cls("dds").getClassLoader(),new Class<?>[]{cls("dds")},(proxy,method,values)->{
                String m=method.getName();
                if(m.equals("I_")) return -64;
                if(m.equals("J_")) return 384;
                if(m.equals("am")) return 320;
                if(m.equals("f_") && values.length==1 && pos.isInstance(values[0])) return true; // ensureCanWrite
                if(m.equals("E")) return level;
                if(m.equals("D_")) return dimension;
                if(m.equals("toString")) return "atlas-fossil-probe";
                if(m.equals("a") && values.length==3 && cls("dyy$a").isInstance(values[0])) return Math.max(-64,Math.min(320,floor((int)values[1],(int)values[2],scenario)+1));
                if(values!=null && values.length>0 && pos.isInstance(values[0])) {
                    int px=(int)x.invoke(values[0]),py=(int)y.invoke(values[0]),pz=(int)z.invoke(values[0]);
                    String k=px+" "+py+" "+pz;
                    Object s=changed.getOrDefault(k,states.get(base(px,py,pz,scenario,start)));
                    if(m.equals("a_")) return s;
                    if(m.equals("b_")) return fluid.invoke(s);
                    if(m.equals("a") && values.length==3 && state.isInstance(values[1])) {
                        if(py < -64 || py >= 320) return false;
                        changed.put(k,values[1]);return true;
                    }
                    if(m.equals("b") && values.length==2 && cls("dfy").isInstance(values[1])) return null; // blockUpdated notification
                    if(m.equals("a") && values.length==3 && cls("epd").isInstance(values[1])) return null; // fluid ticks do not execute during placement
                }
                throw new UnsupportedOperationException(method.toString());
            });
            Object random=cls("dzx").getConstructor(cls("ayw")).newInstance(cls("dzz").getConstructor(long.class).newInstance(seed));
            Object p=pos.getConstructor(int.class,int.class,int.class).newInstance(start[0],start[1],start[2]);
            boolean success=(boolean)cls("ebq").getMethod("a",cls("dds"),cls("duz"),cls("ayw"),pos)
                .invoke(configured.get("minecraft:"+name),world,null,random,p);
            List<String> output=new ArrayList<>();
            for(var change:changed.entrySet()) output.add(change.getKey()+" "+stateID.invoke(null,change.getValue()));
            output.sort(String::compareTo);
            if (args.length > 0 && name.equals("fossil_coal") && seed == 1 && mode == 0) output.forEach(System.out::println);
            String hash=HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(String.join("\n",output).getBytes(StandardCharsets.UTF_8)));
            System.out.printf("fossil_case minecraft:%s %d %d %d %d %d %d %d %d %s%n",name,seed,mode,start[0],start[1],start[2],success?1:0,changed.size(),cls("dzx").getMethod("g").invoke(random),hash);
        }
    }
}
