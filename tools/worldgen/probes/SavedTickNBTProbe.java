package atlasprobe;

import com.google.gson.Gson;
import java.io.*;
import java.lang.reflect.*;
import java.util.*;
import java.util.function.*;

/** Original ChunkSerializer fluid lookup + LevelChunkTicks NBT load/save. */
public final class SavedTickNBTProbe {
    static String tickKey="fluid_ticks";
    static String bytes(Object root)throws Exception {
        ByteArrayOutputStream out=new ByteArrayOutputStream();
        ScheduledTickProbe.call(null,"uo","a",new Class<?>[]{ScheduledTickProbe.cls("ub"),DataOutput.class},root,new DataOutputStream(out));
        return Base64.getEncoder().encodeToString(out.toByteArray());
    }
    static Object root(Object ticks)throws Exception {
        Object root=ScheduledTickProbe.cls("ub").getConstructor().newInstance();
        ScheduledTickProbe.call(root,"ub","a",new Class<?>[]{String.class,ScheduledTickProbe.cls("uy")},tickKey,ticks);return root;
    }
    public static void main(String[]args)throws Exception {
        Object version=ScheduledTickProbe.call(null,"t","a",new Class<?>[]{});ScheduledTickProbe.call(null,"ab","a",new Class<?>[]{ScheduledTickProbe.cls("ae")},version);ScheduledTickProbe.call(null,"akt","a",new Class<?>[]{});
        boolean blocks=args.length>0 && args[0].equals("blocks");if(blocks)tickKey="block_ticks";
        Method lookup=ScheduledTickProbe.cls("dwi").getDeclaredMethod(blocks?"b":"c",String.class);lookup.setAccessible(true);
        Function<String,Optional<Object>> resolve=name->{try{@SuppressWarnings("unchecked") Optional<Object> value=(Optional<Object>)lookup.invoke(null,name);return value;}catch(Exception e){throw new RuntimeException(e);}};
        Map<Object,String> names=new IdentityHashMap<>();for(int i=0;i<5;i++)names.put(ScheduledTickProbe.cls("epf").getField(String.valueOf((char)('a'+i))).get(null),"minecraft:"+ScheduledTickProbe.TYPES[i]);
        Object registry=ScheduledTickProbe.cls("lt").getField("e").get(null);
        Function<Object,String> name=blocks?b->{try{return registry.getClass().getMethod("b",Object.class).invoke(registry,b).toString();}catch(Exception e){throw new RuntimeException(e);}}:names::get;
        String[] inputs={
            "{}", "{fluid_ticks:[]}", "{fluid_ticks:[1,2,3]}", "{fluid_ticks:'wrong'}",
            "{fluid_ticks:[{}]}", "{fluid_ticks:[{i:'water'}]}",
            "{fluid_ticks:[{i:'minecraft:water',x:-1,y:-53,z:0,t:30,p:0},{i:'minecraft:water',x:-1,y:-53,z:0,t:5,p:-3}]}",
            "{fluid_ticks:[{i:'minecraft:lava',x:16,y:319,z:-17,t:-9,p:9}]}",
            "{fluid_ticks:[{i:'flowing_lava',x:-17,y:-64,z:16,t:-2147483648,p:-9}]}",
            "{fluid_ticks:[{i:'minecraft:empty',x:0,y:400,z:0,t:2147483647,p:3}]}",
            "{fluid_ticks:[{i:'flowing_water',x:15,y:65,z:15,t:4294967297L,p:4294967294L}]}",
            "{fluid_ticks:[{i:'water',x:-0.1d,y:64.9f,z:-16.01f,t:-4.1d,p:2.9f}]}",
            "{fluid_ticks:[{i:'water',x:0b,y:64s,z:0L,t:-4s,p:-2b}]}",
            "{fluid_ticks:[{i:'water',x:'bad',y:[],z:{},t:'40',p:'-3'}]}",
            "{fluid_ticks:[{i:'other:water'},{i:'minecraft:missing'},{i:'WATER'},{i:'minecraft:water:bad'},{i:''},{i:3},{i:':water'}]}",
            "{fluid_ticks:[{i:'water',t:1e100d,p:-1e100d},{i:'lava',t:-1e100f,p:1e100f}]}",
            "{fluid_ticks:[{i:'water',x:-2147483649d,z:2147483648d,t:3.4028235e38f,p:-3.4028235e38f}]}",
            "{fluid_ticks:[{i:'water',x:-2147483648,z:2147483647,t:30,p:1},{i:'lava',x:2147483647,z:-2147483648,t:10,p:-1}]}"
        };
        if(blocks)for(int i=0;i<inputs.length;i++)inputs[i]=inputs[i].replace("fluid_ticks","block_ticks").replace("flowing_water","mangrove_leaves").replace("flowing_lava","gravel").replace("water","oak_leaves").replace("lava","sand").replace("empty","air").replace("WATER","OAK_LEAVES");
        Gson gson=new Gson();System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        for(int[] cp:new int[][]{{0,0},{-1,0},{1,-2},{-2,1},{-1,-2},{-134217728,134217727},{134217727,-134217728}})for(int i=0;i<inputs.length;i++){
            Object input=ScheduledTickProbe.call(null,"uz","a",new Class<?>[]{String.class},inputs[i]);
            Object list=ScheduledTickProbe.call(input,"ub","c",new Class<?>[]{String.class,int.class},tickKey,10);
            Object queue=ScheduledTickProbe.call(null,"eyo","a",new Class<?>[]{ScheduledTickProbe.cls("uh"),Function.class,ScheduledTickProbe.cls("dcd")},list,resolve,ScheduledTickProbe.chunkPos(cp[0],cp[1]));
            Object saved=ScheduledTickProbe.call(queue,"eyo","a",new Class<?>[]{long.class,Function.class},0L,name);
            ScheduledTickProbe.call(queue,"eyo","a",new Class<?>[]{long.class},100L);
            Object active=ScheduledTickProbe.call(queue,"eyo","a",new Class<?>[]{long.class,Function.class},104L,name);
            System.out.println("saved_tick_nbt "+gson.toJson(ScheduledTickProbe.obj("case",i,"x",cp[0],"z",cp[1],"input",bytes(input),"saved",bytes(root(saved)),"active",bytes(root(active)))));
        }
    }
}
