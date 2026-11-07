package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.util.*;
import java.util.function.Supplier;
import java.util.function.Function;
import java.util.zip.GZIPOutputStream;

/** Bootstrap-visible sink for the bundler's isolated original-server classloader. */
public final class DecorationTraceSink {
    private static PrintWriter output;
    private static File snapshots;
    private static long sequence;
    private static final ThreadLocal<Object> activeChunk = new ThreadLocal<>();
    private static final ThreadLocal<Object[]> activeFeature = new ThreadLocal<>();
    private static final Map<Object, Map<Object, String>> featureIndices = new IdentityHashMap<>();

    public static synchronized void open(String path) throws IOException {
        output = new PrintWriter(new FileOutputStream(path), true);
        snapshots = new File(new File(path).getParentFile(), "feature-snapshots");
        snapshots.mkdirs();
    }

    public static synchronized void note(String text) { output.println(text); }

    private static boolean observedChunk(int x,int z){return x==-30&&z==-56||x==62&&z==-60||x==0&&z==10||x==-3&&z==44||x==4&&z==12||x==72&&z==-62;}
    private static boolean overworld(Object world)throws ReflectiveOperationException {
        Object dimension=world.getClass().getMethod("af").invoke(world);
        return dimension.getClass().getMethod("a").invoke(dimension).toString().equals("minecraft:overworld");
    }
    private static String registryName(Object world,String kind,Object value)throws ReflectiveOperationException {
        ClassLoader loader=world.getClass().getClassLoader();Object registry=Class.forName("lt",false,loader).getField(kind.equals("block")?"e":"c").get(null);
        return Class.forName("jz",false,loader).getMethod("b",Object.class).invoke(registry,value).toString();
    }
    public static synchronized void scheduledDispatch(String kind,Object world,Object pos,Object type){
        try {
            int x=(Integer)pos.getClass().getMethod("u").invoke(pos),y=(Integer)pos.getClass().getMethod("v").invoke(pos),z=(Integer)pos.getClass().getMethod("w").invoke(pos);
            boolean beachedTickRing=Math.abs((x>>4)-72)<=1&&Math.abs((z>>4)+62)<=1;
            if((!observedChunk(x>>4,z>>4)&&!beachedTickRing)||!overworld(world))return;
            long time=(Long)world.getClass().getMethod("Z").invoke(world);
            output.printf("scheduled_dispatch %d %s %d %d %d %s%n",time,kind,x,y,z,registryName(world,kind,type));output.flush();
        }catch(ReflectiveOperationException e){throw new IllegalStateException("Cannot observe native scheduled dispatch",e);}
    }
    private static String savedTicks(Object world,Object ticks,long time,String kind)throws ReflectiveOperationException,IOException {
        Function<Object,String> names=value->{try{return registryName(world,kind,value);}catch(ReflectiveOperationException e){throw new IllegalStateException(e);}};
        Object list=ticks.getClass().getMethod("a",long.class,Function.class).invoke(ticks,time,names);
        ClassLoader loader=world.getClass().getClassLoader();var bytes=new ByteArrayOutputStream();
        Class.forName("uo",false,loader).getMethod("a",Class.forName("uy",false,loader),DataOutput.class).invoke(null,list,new DataOutputStream(bytes));
        return Base64.getEncoder().encodeToString(bytes.toByteArray());
    }
    public static synchronized void beforeChunkSave(Object map,Object chunk){
        try {
            ClassLoader loader=chunk.getClass().getClassLoader();if(!Class.forName("dvi",false,loader).isInstance(chunk))return;
            Object p=chunk.getClass().getMethod("f").invoke(chunk);int x=p.getClass().getField("e").getInt(p),z=p.getClass().getField("f").getInt(p);
            if(!observedChunk(x,z))return;Object world=field(map,"aqb","s");if(!overworld(world))return;
            long time=(Long)world.getClass().getMethod("Z").invoke(world);boolean dirty=(Boolean)chunk.getClass().getMethod("i").invoke(chunk);
            if(x==72&&z==-62){
                Object levelTicks=world.getClass().getMethod("O").invoke(world),deadlines=field(levelTicks,"eyq","e");
                Field capacity=deadlines.getClass().getDeclaredField("n");capacity.setAccessible(true);
                output.printf("shipwreck_deadlines %d %d %s%n",time,capacity.getInt(deadlines),deadlines);
            }
            output.printf("live_ticks %d %d %d %s %s %s%n",time,x,z,dirty,savedTicks(world,field(chunk,"dvi","v"),time,"block"),savedTicks(world,field(chunk,"dvi","w"),time,"fluid"));output.flush();
        }catch(ReflectiveOperationException|IOException e){throw new IllegalStateException("Cannot serialize native live ticks",e);}
    }

    public static synchronized void magma(Object pos,Object face,Object neighbor,Object world){
        try {
            int x=(Integer)pos.getClass().getMethod("u").invoke(pos),y=(Integer)pos.getClass().getMethod("v").invoke(pos),z=(Integer)pos.getClass().getMethod("w").invoke(pos);
            if(x==-479&&y==48&&z==-896){output.printf("magma_shape %s %s %s%n",face,neighbor,world.getClass().getName());output.flush();}
        }catch(ReflectiveOperationException e){throw new IllegalStateException(e);}
    }

    public static synchronized void activation(String phase,Object chunk) {
        try {
            Object p=chunk.getClass().getMethod("f").invoke(chunk);
            int x=p.getClass().getField("e").getInt(p),z=p.getClass().getField("f").getInt(p);
            if((Math.abs(x+30)>2||Math.abs(z+56)>2)&&!observedChunk(x,z))return;
            Object ticks=field(chunk,"dvi","v");
            output.printf("activation %s %d %d%n",phase,x,z);
            if(x==-30&&z==-56){
                output.printf("pending %s %s%n",phase,field(ticks,"eyo","b"));
                output.printf("queued %s %s%n",phase,field(ticks,"eyo","a"));
            }
            if(x==4&&z==12||x==72&&z==-62){
                Object fluids=field(chunk,"dvi","w");
                output.printf("shipwreck_pending %s %d %d %s%n",phase,x,z,field(fluids,"eyo","b"));
                output.printf("shipwreck_queued %s %d %d %s%n",phase,x,z,field(fluids,"eyo","a"));
            }
            output.flush();
        }catch(ReflectiveOperationException e){throw new IllegalStateException("Cannot inspect original activation",e);}
    }

    private static Object field(Object value, String owner, String name) throws ReflectiveOperationException {
        Field field = Class.forName(owner, false, value.getClass().getClassLoader()).getDeclaredField(name);
        field.setAccessible(true);
        return field.get(value);
    }

    private static void featureState(String phase, Object[] context) throws ReflectiveOperationException {
        Object source = field(context[1], "dzx", "d");
        Object state = field(source, "dzz", "e");
        output.printf("feature_random %s %s %s %s%n", phase, context[0],
                      Long.toUnsignedString((Long)field(state,"dzy","b")),
                      Long.toUnsignedString((Long)field(state,"dzy","c")));
        output.flush();
        String key = (String)context[0];
        String[] fields = key.split(" ");
        if (fields[2].equals("6") && Integer.parseInt(fields[3])<=10) snapshot(phase,key,context[2]);
    }

    private static void snapshot(String phase, String key, Object world) throws ReflectiveOperationException {
        ClassLoader loader = world.getClass().getClassLoader();
        Class<?> posClass = Class.forName("jd", false, loader), stateClass = Class.forName("dtc", false, loader);
        Class<?> mutableClass = Class.forName("jd$a", false, loader);
        Object pos = mutableClass.getConstructor().newInstance();
        Method move = mutableClass.getMethod("d",int.class,int.class,int.class);
        Method get = Class.forName("dcc", false, loader).getMethod("a_",posClass);
        Method id = Class.forName("dfy", false, loader).getMethod("i",stateClass);
        Map<Object,Integer> ids = new IdentityHashMap<>();
        File path = new File(snapshots,phase+"_"+key.replace(' ','_')+".gz");
        try (DataOutputStream stream = new DataOutputStream(new BufferedOutputStream(new GZIPOutputStream(new FileOutputStream(path))))) {
            for (int y=-64;y<320;y++) for(int z=704;z<720;z++) for(int x=-48;x< -32;x++) {
                move.invoke(pos,x,y,z);
                Object state = get.invoke(world,pos);
                Integer value = ids.get(state);
                if(value==null) { value=(Integer)id.invoke(null,state); ids.put(state,value); }
                stream.writeShort(value);
            }
        } catch(IOException error) { throw new UncheckedIOException(error); }
    }

    public static synchronized void featureBegin(Object feature, Object generator, Object random, Object world) {
        try {
            Object chunk = activeChunk.get();
            if (chunk == null) return;
            Object p = chunk.getClass().getMethod("f").invoke(chunk);
            int x = p.getClass().getField("e").getInt(p), z = p.getClass().getField("f").getInt(p);
            if (x < -4 || x > -2 || z < 43 || z > 45) return;
            Map<Object,String> indices = featureIndices.get(generator);
            if (indices == null) {
                indices = new IdentityHashMap<>();
                List<?> steps = (List<?>)((Supplier<?>)field(generator,"duz","c")).get();
                for (int step=0;step<steps.size();step++) {
                    Object entry = steps.get(step);
                    List<?> features = (List<?>)entry.getClass().getMethod("a").invoke(entry);
                    for (int index=0;index<features.size();index++) indices.put(features.get(index),step+" "+index);
                }
                featureIndices.put(generator,indices);
            }
            String index = indices.get(feature);
            if (index == null) throw new IllegalStateException("Unknown original feature index");
            Object[] context = {x+" "+z+" "+index, random, world};
            activeFeature.set(context);
            featureState("begin",context);
        } catch (ReflectiveOperationException exception) { throw new IllegalStateException(exception); }
    }

    public static synchronized void featureEnd() {
        Object[] context = activeFeature.get();
        if (context == null) return;
        activeFeature.remove();
        try { featureState("end",context); }
        catch (ReflectiveOperationException exception) { throw new IllegalStateException(exception); }
    }

    public static synchronized void configuredOrigin(Object p) {
        Object[] context=activeFeature.get();
        if(context==null) return;
        String key=(String)context[0];
        if(!key.equals("-3 44 6 0")) return;
        try {
            int x=(Integer)p.getClass().getMethod("u").invoke(p),y=(Integer)p.getClass().getMethod("v").invoke(p),z=(Integer)p.getClass().getMethod("w").invoke(p);
            Object source=field(context[1],"dzx","d"),state=field(source,"dzz","e");
            output.printf("configured_origin %s %d %d %d %s %s%n",key,x,y,z,
                          Long.toUnsignedString((Long)field(state,"dzy","b")),Long.toUnsignedString((Long)field(state,"dzy","c")));
            output.flush();
        } catch(ReflectiveOperationException exception) { throw new IllegalStateException(exception); }
    }

    public static synchronized void event(String phase, Object chunk) {
        try {
            if (phase.equals("begin")) activeChunk.set(chunk);
            else { chunk = activeChunk.get(); activeChunk.remove(); }
            Object position = chunk.getClass().getMethod("f").invoke(chunk);
            int x = position.getClass().getField("e").getInt(position);
            int z = position.getClass().getField("f").getInt(position);
            output.printf("decoration %d %s %d %d %s%n", sequence++, phase, x, z,
                          Thread.currentThread().getName());
            output.flush();
        } catch (ReflectiveOperationException exception) {
            throw new IllegalStateException("Cannot observe original decoration", exception);
        }
    }
}
