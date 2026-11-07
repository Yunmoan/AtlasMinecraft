package atlasprobe;

import com.google.gson.Gson;
import java.io.*;
import java.lang.reflect.*;
import java.util.*;
import java.util.function.*;

/** Runs the original LevelChunkTicks and LevelTicks, including their heaps and
 * fastutil deadline map. Every request and callback mutation is recorded. */
public final class ScheduledTickProbe {
    static Class<?> cls(String n)throws Exception{return Class.forName(n);}
    static Object call(Object o,String c,String n,Class<?>[] t,Object...v)throws Exception{return cls(c).getMethod(n,t).invoke(o,v);}
    static Object field(Object o,String c,String n)throws Exception{Field f=cls(c).getDeclaredField(n);f.setAccessible(true);return f.get(o);}
    static Map<String,Object> obj(Object...v){Map<String,Object> m=new LinkedHashMap<>();for(int i=0;i<v.length;i+=2)m.put((String)v[i],v[i+1]);return m;}
    static final String[] TYPES={"empty","flowing_water","water","flowing_lava","lava"};
    static Object pos(int[] p)throws Exception{return cls("jd").getConstructor(int.class,int.class,int.class).newInstance(p[0],p[1],p[2]);}
    static int[] coords(Object p)throws Exception{return new int[]{(int)call(p,"jd","u",new Class<?>[]{}),(int)call(p,"jd","v",new Class<?>[]{}),(int)call(p,"jd","w",new Class<?>[]{})};}
    static Object priority(int p)throws Exception{return call(null,"eyx","a",new Class<?>[]{int.class},p);}
    static Object scheduled(Map<String,Object> t)throws Exception{return cls("eyt").getConstructor(Object.class,cls("jd"),long.class,cls("eyx"),long.class).newInstance(t.get("type"),pos((int[])t.get("pos")),((Number)t.get("trigger")).longValue(),priority((int)t.get("priority")),((Number)t.get("order")).longValue());}
    static Object saved(Map<String,Object> t)throws Exception{return cls("eys").getConstructor(Object.class,cls("jd"),int.class,cls("eyx")).newInstance(t.get("type"),pos((int[])t.get("pos")),((Number)t.get("delay")).intValue(),priority((int)t.get("priority")));}
    static Map<String,Object> scheduledView(Object t)throws Exception{return obj("type",call(t,"eyt","a",new Class<?>[]{}),"pos",coords(call(t,"eyt","b",new Class<?>[]{})),"trigger",call(t,"eyt","c",new Class<?>[]{}),"priority",call(call(t,"eyt","d",new Class<?>[]{}),"eyx","a",new Class<?>[]{}),"order",call(t,"eyt","e",new Class<?>[]{}));}
    static Object chunkPos(int x,int z)throws Exception{return cls("dcd").getConstructor(int.class,int.class).newInstance(x,z);}
    static long key(int x,int z){return Integer.toUnsignedLong(x)|(Integer.toUnsignedLong(z)<<32);}
    static List<Map<String,Object>> save(Object chunk,long time)throws Exception{
        List<Map<String,Object>> result=new ArrayList<>();
        Function<Object,String> name=Object::toString;
        for(Object tag:(List<?>)call(chunk,"eyo","a",new Class<?>[]{long.class,Function.class},time,name)) {
            int[] p=new int[3];String[] keys={"x","y","z"};for(int i=0;i<3;i++)p[i]=(int)call(tag,"ub","h",new Class<?>[]{String.class},keys[i]);
            result.add(obj("type",call(tag,"ub","l",new Class<?>[]{String.class},"i"),"pos",p,"delay",call(tag,"ub","h",new Class<?>[]{String.class},"t"),"priority",call(tag,"ub","h",new Class<?>[]{String.class},"p")));
        }return result;
    }
    static final class Case {
        final Object level; final Set<Long> allowed=new HashSet<>(); final LinkedHashMap<Long,Object> chunks=new LinkedHashMap<>();
        final List<Map<String,Object>> operations=new ArrayList<>(); final Map<String,Object> row;
        long order=1000;
        Case(int seed,int count,boolean ties)throws Exception {
            Object profiler=cls("bnc").getField("a").get(null);
            level=cls("eyq").getConstructor(LongPredicate.class,Supplier.class).newInstance((LongPredicate)allowed::contains,(Supplier<Object>)()->profiler);
            row=obj("seed",seed,"count",count,"ties",ties,"operations",operations);
        }
        void load(int x,int z,List<Map<String,Object>> pending)throws Exception {
            List<Object> ticks=new ArrayList<>();for(var t:pending)ticks.add(saved(t));
            Object c=cls("eyo").getConstructor(List.class).newInstance(ticks);chunks.put(key(x,z),c);allowed.add(key(x,z));
            call(level,"eyq","a",new Class<?>[]{cls("dcd"),cls("eyo")},chunkPos(x,z),c);
            operations.add(obj("op","load","x",x,"z",z,"saved",pending));
        }
        void unpack(int x,int z,long time)throws Exception {
            call(chunks.get(key(x,z)),"eyo","a",new Class<?>[]{long.class},time);operations.add(obj("op","unpack","x",x,"z",z,"time",time));
        }
        void request(Map<String,Object> t)throws Exception {
            call(level,"eyq","a",new Class<?>[]{cls("eyt")},scheduled(t));
            operations.add(obj("op","schedule","tick",t));
        }
        void snapshot(long time)throws Exception {
            List<Object> result=new ArrayList<>();for(var e:chunks.entrySet()){
                var saved=save(e.getValue(),time);List<Object> has=new ArrayList<>();
                for(var t:saved)has.add(call(level,"eyq","a",new Class<?>[]{cls("jd"),Object.class},pos((int[])t.get("pos")),t.get("type")));
                result.add(obj("x",(int)(long)e.getKey(),"z",(int)(e.getKey()>>32),"saved",saved,"has",has));
            }
            operations.add(obj("op","snapshot","time",time,"chunks",result,"count",call(level,"eyq","a",new Class<?>[]{})));
        }
        void tick(long time,int limit,int mode,boolean reschedule)throws Exception {
            allowed.clear();for(long k:chunks.keySet())if(mode==0 || ((int)k&1)==0)allowed.add(k);
            List<Object> events=new ArrayList<>();int[] count={0};
            BiConsumer<Object,Object> consumer=(p,type)->{try{
                List<?> executed=(List<?>)field(level,"eyq","h");Object t=executed.get(executed.size()-1);
                Object query=pos(new int[]{0,64,0});
                Map<String,Object> event=obj("tick",scheduledView(t),"has",call(level,"eyq","a",new Class<?>[]{cls("jd"),Object.class},p,type),"will",call(level,"eyq","b",new Class<?>[]{cls("jd"),Object.class},p,type),"queryHas",call(level,"eyq","a",new Class<?>[]{cls("jd"),Object.class},query,TYPES[2]),"queryWill",call(level,"eyq","b",new Class<?>[]{cls("jd"),Object.class},query,TYPES[2]));
                List<Object> actions=new ArrayList<>();event.put("actions",actions);
                if(reschedule && count[0]++<3){
                    Map<String,Object> again=obj("type",type,"pos",coords(p),"trigger",time,"priority",0,"order",order++);actions.add(again);call(level,"eyq","a",new Class<?>[]{cls("eyt")},scheduled(again));
                    Map<String,Object> duplicate=obj("type",type,"pos",coords(p),"trigger",time-50,"priority",-3,"order",order++);actions.add(duplicate);call(level,"eyq","a",new Class<?>[]{cls("eyt")},scheduled(duplicate));
                }
                event.put("hasAfter",call(level,"eyq","a",new Class<?>[]{cls("jd"),Object.class},p,type));events.add(event);
            }catch(Exception e){throw new RuntimeException(e);}};
            call(level,"eyq","a",new Class<?>[]{long.class,int.class,BiConsumer.class},time,limit,consumer);
            operations.add(obj("op","tick","time",time,"limit",limit,"mode",mode,"events",events));
        }
        void unload(int x,int z)throws Exception{call(level,"eyq","a",new Class<?>[]{cls("dcd")},chunkPos(x,z));chunks.remove(key(x,z));operations.add(obj("op","unload","x",x,"z",z));}
    }
    public static void main(String[]args)throws Exception {
        Object version=call(null,"t","a",new Class<?>[]{});call(null,"ab","a",new Class<?>[]{cls("ae")},version);call(null,"akt","a",new Class<?>[]{});
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));Gson gson=new Gson();
        for(int seed=0;seed<12;seed++)for(int size:new int[]{1,2,7,24,25,48,100})for(boolean ties:new boolean[]{false,true}) {
            Random random=new Random(seed);Case c=new Case(seed,size,ties);List<int[]> positions=new ArrayList<>();
            for(int i=0;i<size;i++)positions.add(i==0?new int[]{0,0}:new int[]{i%2==0?i:-i,(i*17+seed)%43-21});Collections.shuffle(positions,random);
            for(int[] cp:positions){
                int x=cp[0],z=cp[1];List<Map<String,Object>> pending=new ArrayList<>();
                for(int j=0;j<8;j++)pending.add(obj("type",TYPES[(j+2)%5],"pos",new int[]{x*16+j%4,64+j/4,z*16},"delay",ties?0:random.nextInt(41)-20,"priority",ties?0:random.nextInt(7)-3));
                // SavedTick.loadTickList preserves duplicate saved entries;
                // unlike newly scheduled requests, unpacking does not dedup.
                if(seed%3==0)pending.add(new LinkedHashMap<>(pending.get(0)));
                if(seed%4==0){pending.get(7).put("delay",Integer.MIN_VALUE);pending.get(7).put("priority",-9);}
                c.load(x,z,pending);
                // Dedup against still-packed entries, then add another type at
                // the same position. Unpack must not erase that newer tick.
                var first=pending.get(0);c.request(obj("type",first.get("type"),"pos",first.get("pos"),"trigger",0L,"priority",-3,"order",c.order++));
                c.request(obj("type",TYPES[0],"pos",first.get("pos"),"trigger",100L,"priority",0,"order",c.order++));
                if(random.nextBoolean())c.unpack(x,z,100);
            }
            c.snapshot(93);for(int[] cp:positions)c.unpack(cp[0],cp[1],100);c.snapshot(100);c.snapshot((1L<<33)+100);
            c.tick(100,0,0,false);c.tick(100,3,1,true);c.snapshot(100);c.tick(100,13,0,true);c.snapshot(100);
            // Remove many containers, forcing shrink of the original map;
            // reload their saved delays at a later game time in reverse order.
            List<Map<String,Object>> detached=new ArrayList<>();int remove=size*4/5;
            for(int i=0;i<remove;i++){int[] cp=positions.get(i);detached.add(obj("x",cp[0],"z",cp[1],"saved",save(c.chunks.get(key(cp[0],cp[1])),100)));c.unload(cp[0],cp[1]);}
            c.tick(110,11,0,true);c.snapshot(110);Collections.reverse(detached);
            for(var entry:detached){int x=(int)entry.get("x"),z=(int)entry.get("z");@SuppressWarnings("unchecked") List<Map<String,Object>> pending=(List<Map<String,Object>>)entry.get("saved");c.load(x,z,pending);c.unpack(x,z,150);}
            c.tick(150,27,0,true);c.snapshot(150);c.tick(170,65536,1,false);c.snapshot(170);c.tick(200,65536,0,false);c.snapshot(200);c.tick(200,65536,0,false);c.snapshot(200);
            System.out.println("scheduled_tick "+gson.toJson(c.row));
        }
    }
}
