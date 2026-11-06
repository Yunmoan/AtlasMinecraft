package atlasprobe;

import java.io.InputStream;
import java.io.ByteArrayOutputStream;
import java.io.DataOutputStream;
import java.nio.charset.StandardCharsets;
import java.lang.reflect.Field;
import java.util.HashMap;
import java.util.Optional;
import java.util.List;
import java.util.stream.Stream;
import java.util.ArrayList;
import java.util.Collections;
import java.util.Base64;

/** Uses the original template loader; supplies an original manager preloaded
 * with the three resource templates without constructing a server/save. */
final class IglooProbeSupport {
    static Class<?> cls(String n)throws Exception{return Class.forName(n);}
    static Object manager()throws Exception {
        return manager(new String[]{"igloo/top","igloo/middle","igloo/bottom"});
    }
    static Object manager(String[] names)throws Exception {
        Class<?> unsafeClass=cls("sun.misc.Unsafe");Field unsafeField=unsafeClass.getDeclaredField("theUnsafe");unsafeField.setAccessible(true);
        Object unsafe=unsafeField.get(null);
        Object manager=unsafeClass.getMethod("allocateInstance",Class.class).invoke(unsafe,cls("enu"));
        var repository=cls("enu").getDeclaredField("f");repository.setAccessible(true);
        HashMap<Object,Optional<Object>> templates=new HashMap<>();repository.set(manager,templates);
        Object blockLookup=cls("jz").getMethod("q").invoke(cls("lt").getField("e").get(null));
        for(String name:names) {
            Object template=cls("ent").getConstructor().newInstance();
            try(InputStream input=cls("ent").getResourceAsStream("/data/minecraft/structure/"+name+".nbt")) {
                if(input!=null) {
                    Object tag=cls("uo").getMethod("a",InputStream.class,cls("uk")).invoke(null,input,cls("uk").getMethod("a").invoke(null));
                    cls("ent").getMethod("a",cls("jn"),cls("ub")).invoke(template,blockLookup,tag);
                }
                Object key=cls("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name);
                templates.put(key,Optional.of(template));
            }
        }
        return manager;
    }

    static Object registryAccess()throws Exception {
        var createPack=cls("ats").getDeclaredMethod("b");createPack.setAccessible(true);
        Object pack=createPack.invoke(null);
        Object resources=cls("atx").getConstructor(cls("ass"),List.class).newInstance(cls("ass").getField("b").get(null),List.of(pack));
        Object builtin=cls("ka").getMethod("a",cls("jz")).invoke(null,cls("lt").getField("aA").get(null));
        Object loaded=cls("akm").getMethod("a",cls("aue"),cls("ka"),List.class).invoke(null,resources,builtin,cls("akm").getField("a").get(null));
        return cls("ka$c").getConstructor(Stream.class).newInstance(Stream.concat((Stream<?>)cls("ka").getMethod("c").invoke(builtin),(Stream<?>)cls("ka").getMethod("c").invoke(loaded)));
    }

    static String canonicalNBT(Object tag)throws Exception {
        if(cls("ub").isInstance(tag)) {
            List<String> keys=new ArrayList<>((java.util.Set<String>)cls("ub").getMethod("e").invoke(tag));Collections.sort(keys);
            StringBuilder text=new StringBuilder("C{");
            for(String key:keys)text.append(Base64.getEncoder().encodeToString(key.getBytes(StandardCharsets.UTF_8))).append(':').append(canonicalNBT(cls("ub").getMethod("c",String.class).invoke(tag,key))).append(';');
            return text.append('}').toString();
        }
        if(cls("uh").isInstance(tag)) {
            StringBuilder text=new StringBuilder("L"+cls("uh").getMethod("f").invoke(tag)+"[");
            for(Object value:(Iterable<?>)tag)text.append(canonicalNBT(value)).append(';');
            return text.append(']').toString();
        }
        ByteArrayOutputStream bytes=new ByteArrayOutputStream();
        // writeAnyTag is the 1.21.1 network format (tag type plus payload).
        // writeUnnamedTag also writes an empty root name and is the disk format.
        cls("uo").getMethod("a",cls("uy"),java.io.DataOutput.class).invoke(null,tag,new DataOutputStream(bytes));
        return Base64.getEncoder().encodeToString(bytes.toByteArray());
    }
}
