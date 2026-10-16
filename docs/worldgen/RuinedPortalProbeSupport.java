package atlasprobe;

import java.lang.reflect.InvocationHandler;
import java.lang.reflect.Proxy;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.function.Function;

final class RuinedPortalProbeSupport {
    static Class<?> cls(String n)throws Exception{return Class.forName(n);}
    static String[] names() {
        List<String> names=new ArrayList<>();
        for(int i=1;i<=10;i++)names.add("ruined_portal/portal_"+i);
        for(int i=1;i<=3;i++)names.add("ruined_portal/giant_portal_"+i);
        return names.toArray(String[]::new);
    }
    static void bindTags()throws Exception {
        var createPack=cls("ats").getDeclaredMethod("b");createPack.setAccessible(true);
        Object pack=createPack.invoke(null);
        Object resources=cls("atx").getConstructor(cls("ass"),List.class).newInstance(cls("ass").getField("b").get(null),List.of(pack));
        for(String[] entry:new String[][]{{"e","block","f"},{"c","fluid","D"}}) {
            Object registry=cls("lt").getField(entry[0]).get(null),registryKey=cls("lu").getField(entry[2]).get(null);
            Function<Object,Optional<?>> lookup=key->{try{return (Optional<?>)cls("jz").getMethod("c",cls("akr")).invoke(registry,key);}catch(Exception error){throw new IllegalStateException(error);}};
            Object loader=cls("awv").getConstructor(Function.class,String.class).newInstance(lookup,"tags/"+entry[1]);
            Map<?,?> loaded=(Map<?,?>)cls("awv").getMethod("b",cls("aue")).invoke(loader,resources);
            Map<Object,Object> tags=new HashMap<>();
            for(var tag:loaded.entrySet())tags.put(cls("awu").getMethod("a",cls("akq"),cls("akr")).invoke(null,registryKey,tag.getKey()),new ArrayList<>((java.util.Collection<?>)tag.getValue()));
            cls("jz").getMethod("a",Map.class).invoke(registry,tags);
        }
    }
    static Object properties(int variant)throws Exception {
        // Warm land, cold buried, ocean, overgrown/vines, blackstone, full air.
        return cls("emf$a").getConstructor(boolean.class,float.class,boolean.class,boolean.class,boolean.class,boolean.class).newInstance(variant==1,variant==0?0.0f:variant==2?1.0f:0.8f,variant==5,variant==3,variant==3,variant==4);
    }
    static Object block(String name)throws Exception {
        return cls("dfy").getMethod("o").invoke(cls("jz").getMethod("a",cls("akr")).invoke(cls("lt").getField("e").get(null),cls("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name)));
    }
    static Object world(int mode,Map<String,Object> edits,Map<String,Object> entities,Object registries)throws Exception {
        Object dimensionRegistry=cls("ka").getMethod("d",cls("akq")).invoke(registries,cls("lu").getField("aK").get(null));
        Object dimensionType=cls("jz").getMethod("f",cls("akq")).invoke(dimensionRegistry,cls("dww").getField("a").get(null));
        Map<String,Object> states=new HashMap<>();
        for(String name:new String[]{"air","stone","grass_block","water","lava","bedrock","obsidian"})states.put(name,block(name));
        var x=cls("jd").getMethod("u");var y=cls("jd").getMethod("v");var z=cls("jd").getMethod("w");
        Function<int[],Object> read=p->{String key=p[0]+" "+p[1]+" "+p[2];if(edits.containsKey(key))return edits.get(key);int floor=mode==1?60+Math.floorMod(p[0]+2*p[2],5):63;if(mode==4&&Math.floorMod(p[0]+p[1]+p[2],7)==0)return states.get("bedrock");if(mode==5&&Math.floorMod(p[0]+p[1]+p[2],3)==0)return states.get("obsidian");if(p[1]<=floor)return states.get(p[1]==floor?"grass_block":"stone");if(mode==2&&p[1]<=69)return states.get("water");if(mode==3&&p[1]<=69)return states.get("lava");return states.get("air");};
        return Proxy.newProxyInstance(cls("dds").getClassLoader(),new Class<?>[]{cls("dds")},(proxy,method,args)->{
            String name=method.getName();if(name.equals("I_"))return -64;if(name.equals("an"))return 384;if(name.equals("am"))return 320;if(name.equals("H_"))return registries;if(name.equals("J"))return cls("cpn").getField("g").get(null);if(name.equals("x_"))return false;if(name.equals("D_"))return dimensionType;
            if(args!=null&&args.length==3&&cls("dyy$a").isInstance(args[0])) {
                String kind=((Enum<?>)args[0]).name();
                for(int py=319;py>=-64;py--){Object state=read.apply(new int[]{(Integer)args[1],py,(Integer)args[2]});if(kind.startsWith("OCEAN_FLOOR")?((java.util.function.Predicate<Object>)cls("dyy$a").getMethod("e").invoke(args[0])).test(state):!(Boolean)cls("dtb$a").getMethod("i").invoke(state))return py+1;}return -64;
            }
            if(args!=null&&args.length>0&&cls("jd").isInstance(args[0])) {
                int px=(Integer)x.invoke(args[0]),py=(Integer)y.invoke(args[0]),pz=(Integer)z.invoke(args[0]);String key=px+" "+py+" "+pz;
                if(name.equals("a_"))return read.apply(new int[]{px,py,pz});
                if(name.equals("b_"))return cls("dtb$a").getMethod("u").invoke(read.apply(new int[]{px,py,pz}));
                if(name.equals("c_"))return entities.get(key);
                if(name.equals("a")&&args.length==3&&cls("dtc").isInstance(args[1])) {
                    Object old=read.apply(new int[]{px,py,pz}),oldBlock=cls("dtb$a").getMethod("b").invoke(old),block=cls("dtb$a").getMethod("b").invoke(args[1]);edits.put(key,args[1]);
                    if(cls("diq").isInstance(block)){if(oldBlock!=block||!entities.containsKey(key))entities.put(key,cls("diq").getMethod("a",cls("jd"),cls("dtc")).invoke(block,args[0],args[1]));}else entities.remove(key);return true;
                }
            }
            if(method.getReturnType()==void.class)return null;
            if(method.isDefault())return InvocationHandler.invokeDefault(proxy,method,args);
            throw new UnsupportedOperationException(method.toString());
        });
    }
}
