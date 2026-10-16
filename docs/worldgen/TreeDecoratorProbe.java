package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.lang.reflect.Method;
import java.lang.reflect.Proxy;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.*;
import java.util.function.BiConsumer;
import java.util.function.Predicate;
import java.util.stream.Stream;

/** Runs all six actual Java 1.21.1 tree decorators, including nest occupants.
 * Collections' independent RNG is reset per case; it is not the world RNG. */
public final class TreeDecoratorProbe {
    static String hash(List<String> output) throws Exception {
        output.sort(String::compareTo);
        return HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(String.join("\n", output).getBytes(StandardCharsets.UTF_8)));
    }

    public static void main(String[] args) throws Exception {
        Object version = Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a", Class.forName("ae")).invoke(null, version);
        Class.forName("akt").getMethod("a").invoke(null);
        TreeFeatureProbe.bindTreeTags();
        Object registries = Class.forName("pd").getMethod("a").invoke(null);
        Object lookup = Class.forName("jo$a").getMethod("b").invoke(registries);
        Object featureLookup = Class.forName("jn$a").getMethod("b", Class.forName("akq")).invoke(lookup, Class.forName("lu").getField("aI").get(null));
        Map<String, Object> configs = new HashMap<>();
        for (Object holder : ((Stream<?>) Class.forName("jo").getMethod("b").invoke(featureLookup)).toList()) {
            Object key = ((Optional<?>) Class.forName("jm").getMethod("e").invoke(holder)).orElseThrow();
            Object configured = Class.forName("jm").getMethod("a").invoke(holder);
            configs.put(Class.forName("akq").getMethod("a").invoke(key).toString(), Class.forName("ebq").getMethod("c").invoke(configured));
        }
        Object attached = ((List<?>) Class.forName("efe").getField("i").get(configs.get("minecraft:mangrove"))).get(1);
        var providerField = attached.getClass().getDeclaredField("e"); providerField.setAccessible(true);
        Object attachedProvider = providerField.get(attached);
        Object down = Class.forName("ji").getField("a").get(null);
        Object[] decorators = {
            Class.forName("ehc").getField("b").get(null),
            Class.forName("egz").getConstructor(float.class).newInstance(1f),
            Class.forName("egy").getConstructor(float.class).newInstance(1f),
            Class.forName("egx").getConstructor(float.class).newInstance(1f),
            ((List<?>) Class.forName("efe").getField("i").get(configs.get("minecraft:mega_spruce"))).get(0),
            Class.forName("egw").getConstructor(float.class,int.class,int.class,Class.forName("egj"),int.class,List.class).newInstance(1f,1,0,attachedProvider,2,List.of(down))
        };
        var shuffleRandom = Collections.class.getDeclaredField("r"); shuffleRandom.setAccessible(true);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out), true));
        Class<?> pos = Class.forName("jd"), stateClass = Class.forName("dtc"), block = Class.forName("dfy");
        Method x = pos.getMethod("u"), y = pos.getMethod("v"), z = pos.getMethod("w");
        Method id = block.getMethod("i", stateClass);
        Object registry = Class.forName("lt").getField("e").get(null);
        Map<String,Object> states = new HashMap<>();
        for (String name : new String[]{"air","grass_block","stone","dirt","oak_log","oak_leaves","vine","mangrove_roots"}) {
            Object b = Class.forName("jz").getMethod("a", Class.forName("akr")).invoke(registry,Class.forName("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name));
            states.put(name,block.getMethod("o").invoke(b));
        }
        Method occupants = Class.forName("dqe").getDeclaredMethod("l"); occupants.setAccessible(true);
        for (int decorator = 0; decorator < decorators.length; decorator++) for (long seed : new long[]{0,1,-1,1234,123456789,42,987654321}) for (int scenario = 0; scenario < 10; scenario++) {
            int ox = scenario == 4 ? -16 : scenario == 7 ? 29999968 : 8, oz = scenario == 4 ? -1 : scenario == 7 ? -29999968 : 8;
            Set<Object> logs = new HashSet<>(), leaves = new HashSet<>(), roots = new HashSet<>();
            Map<String,Object> initial = new HashMap<>(), changed = new HashMap<>(), entities = new HashMap<>();
            for(int h=40;h<=47;h++) {
                Object p=pos.getConstructor(int.class,int.class,int.class).newInstance(ox,h,oz); logs.add(p); initial.put(ox+" "+h+" "+oz,states.get(h==40?"dirt":"oak_log"));
            }
            for(int h=45;h<=48;h++) for(int dx=-2;dx<=2;dx++) for(int dz=-2;dz<=2;dz++) {
                if(scenario==3 || dx==0 && dz==0 && h<48 || Math.abs(dx)==2 && Math.abs(dz)==2) continue;
                Object p=pos.getConstructor(int.class,int.class,int.class).newInstance(ox+dx,h,oz+dz); leaves.add(p); initial.put((ox+dx)+" "+h+" "+(oz+dz),states.get("oak_leaves"));
            }
            if(scenario==8 || scenario==9) for(int dx=3;dx<=4;dx++) {
                int h=scenario==8?38:40;
                Object p=pos.getConstructor(int.class,int.class,int.class).newInstance(ox+dx,h,oz);roots.add(p);initial.put((ox+dx)+" "+h+" "+oz,states.get("mangrove_roots"));
            }
            int mode = scenario;
            Object world = Proxy.newProxyInstance(Class.forName("dds").getClassLoader(),new Class<?>[]{Class.forName("dds")},(proxy,method,values)->{
                String m=method.getName();
                if(m.equals("toString")) return "tree-decorator-probe";
                if(values!=null && values.length>0 && pos.isInstance(values[0])) {
                    int px=(int)x.invoke(values[0]),py=(int)y.invoke(values[0]),pz=(int)z.invoke(values[0]);
                    String key=px+" "+py+" "+pz;
                    Object base=py==40?states.get(mode==1?"stone":"grass_block"):states.get("air");
                    if(mode==2 && py>=41 && py<=46 && Math.abs(px-ox)+Math.abs(pz-oz)==1) base=states.get("stone");
                    if(mode==5 && py==40 && (px+pz)%3==0) base=states.get("air");
                    if(mode==5 && py==38) base=states.get("dirt");
                    if(mode==6 && py==43 && Math.abs(px-ox)==3 && Math.abs(pz-oz)==3) base=states.get("vine");
                    Object state=changed.getOrDefault(key,initial.getOrDefault(key,base));
                    if(m.equals("a_")) return state;
                    if(m.equals("a") && values.length==2 && values[1] instanceof Predicate<?>) return ((Predicate<Object>)values[1]).test(state);
                    if(m.equals("a") && values.length==2 && Class.forName("dqj").isInstance(values[1])) return Optional.ofNullable(entities.get(key));
                    if(m.equals("a") && values.length==3 && stateClass.isInstance(values[1])) {
                        changed.put(key,values[1]);
                        if(Class.forName("dtb$a").getMethod("b").invoke(values[1]).getClass()==Class.forName("dfs")) entities.put(key,Class.forName("dqe").getConstructor(pos,stateClass).newInstance(values[0],values[1]));
                        return true;
                    }
                }
                throw new UnsupportedOperationException(method.toString());
            });
            Object random=Class.forName("dzx").getConstructor(Class.forName("ayw")).newInstance(Class.forName("dzz").getConstructor(long.class).newInstance(seed));
            shuffleRandom.set(null,new Random(seed ^ 0x123456789abcdef0L));
            BiConsumer<Object,Object> setter=(p,s)->{
                try { Class.forName("dds").getMethod("a",pos,stateClass,int.class).invoke(world,p,s,19); } catch(Exception e) { throw new RuntimeException(e); }
            };
            Object context=Class.forName("eha$a").getConstructor(Class.forName("ddc"),BiConsumer.class,Class.forName("ayw"),Set.class,Set.class,Set.class).newInstance(world,setter,random,logs,leaves,roots);
            Class.forName("eha").getMethod("a",Class.forName("eha$a")).invoke(decorators[decorator],context);
            List<String> output=new ArrayList<>(), beeOutput=new ArrayList<>();
            for(var entry:changed.entrySet()) output.add(entry.getKey()+" "+id.invoke(null,entry.getValue()));
            for(var entry:entities.entrySet()) {
                List<String> beeData=new ArrayList<>();
                for(Object occupant:(List<?>)occupants.invoke(entry.getValue())) {
                    Object entityData=Class.forName("dqe$c").getMethod("a").invoke(occupant);
                    Object tag=Class.forName("cxh").getMethod("c").invoke(entityData);
                    String entityID=(String)Class.forName("ub").getMethod("l",String.class).invoke(tag,"id");
                    beeData.add(entityID+":"+Class.forName("dqe$c").getMethod("b").invoke(occupant)+":"+Class.forName("dqe$c").getMethod("c").invoke(occupant));
                }
                beeOutput.add(entry.getKey()+" "+String.join(",",beeData));
            }
            System.out.printf("decorator_case %d %d %d %d %d %d %d %s %d %s%n",decorator,seed,scenario,ox,oz,changed.size(),Class.forName("dzx").getMethod("g").invoke(random),hash(output),entities.size(),hash(beeOutput));
        }
    }
}
