package atlasprobe;

import java.io.*;
import java.util.*;
import java.util.function.Function;
import java.util.zip.GZIPOutputStream;

/** Exercises the original ProtoChunkTicks, not a replacement tick queue. */
public final class GenerationFluidTickProbe {
    static Class<?> cls(String n)throws Exception{return Class.forName(n);}
    static Object call(Object o,String c,String n,Class<?>[] t,Object...v)throws Exception{return cls(c).getMethod(n,t).invoke(o,v);}
    public static void main(String[]args)throws Exception {
        Object version=call(null,"t","a",new Class<?>[]{});call(null,"ab","a",new Class<?>[]{cls("ae")},version);call(null,"akt","a",new Class<?>[]{});
        Object[] fluids=new Object[5];
        String[] names={"empty","flowing_water","water","flowing_lava","lava"};
        for(int i=0;i<5;i++)fluids[i]=cls("epf").getField(String.valueOf((char)('a'+i))).get(null);
        Function<Object,String> name=o->{for(int i=0;i<5;i++)if(o==fluids[i])return "minecraft:"+names[i];throw new IllegalStateException("unknown fluid");};
        Object queue=cls("eyr").getConstructor().newInstance();
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        for(int i=0;i<30;i++) {
            int type=i%5,x=-17+(i/5)%2,y=-63,z=47,priority=i%7-3;
            long trigger=10000L+i*37,order=100-i;
            Object p=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(x,y,z);
            Object tick=cls("eyt").getConstructor(Object.class,cls("jd"),long.class,cls("eyx"),long.class).newInstance(fluids[type],p,trigger,call(null,"eyx","a",new Class<?>[]{int.class},priority),order);
            call(queue,"eyr","a",new Class<?>[]{cls("eyt")},tick);
            System.out.printf("generation_fluid_request %s %d %d %d %d %d %d %s%n",name.apply(fluids[type]),x,y,z,trigger,priority,order,call(queue,"eyr","a",new Class<?>[]{}));
        }
        for(long time:new long[]{0,99999}) {
            Object saved=call(queue,"eyr","b",new Class<?>[]{long.class,Function.class},time,name);
            for(Object tag:(List<?>)saved) {
                System.out.print("generation_fluid_saved "+time+" "+call(tag,"ub","l",new Class<?>[]{String.class},"i"));
                for(String key:new String[]{"x","y","z","t","p"})System.out.print(" "+call(tag,"ub","h",new Class<?>[]{String.class},key));
                System.out.println();
            }
        }
        // Exhaustive block-state -> fluid-type oracle, including aquatic plants,
        // bubble columns, waterlogged blocks and all 16 liquid levels.
        ByteArrayOutputStream bytes=new ByteArrayOutputStream();
        try(GZIPOutputStream gzip=new GZIPOutputStream(bytes)) {
            for(int id=0;id<26684;id++) {
                Object state=call(null,"dfy","a",new Class<?>[]{int.class},id);
                Object fluid=call(call(state,"dtc","u",new Class<?>[]{}),"epe","a",new Class<?>[]{});
                int type=-1;for(int i=0;i<5;i++)if(fluid==fluids[i])type=i;
                if(type<0)throw new IllegalStateException("unknown fluid at state "+id);
                gzip.write(type);
            }
        }
        System.out.println("generation_fluid_states "+Base64.getEncoder().encodeToString(bytes.toByteArray()));
    }
}
