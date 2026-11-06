package atlasprobe;

import java.io.*;
import java.nio.file.*;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.*;

/** Calls original template assembly, then saves every piece in order. */
public final class MansionLayoutProbe {
    static Class<?> cls(String name)throws Exception{return Class.forName(name);}
    public static void main(String[] args)throws Exception{
        Object version=cls("t").getMethod("a").invoke(null);
        cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);
        String[] names;
        try(var files=Files.list(Path.of(args[0]))) {
            names=files.map(p->p.getFileName().toString()).filter(n->n.endsWith(".nbt")).sorted().map(n->"woodland_mansion/"+n.substring(0,n.length()-4)).toArray(String[]::new);
        }
        if(names.length!=73)throw new IllegalStateException("expected 73 original templates");
        Object manager=IglooProbeSupport.manager(names);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        for(long seed:new long[]{0,1,-1,1234,Long.MIN_VALUE,Long.MAX_VALUE,4511451270791457643L})
        for(int[] pos:new int[][]{{0,64,0},{-173,79,243},{29999000,60,-29999000}})
        for(Object rotation:cls("dmm").getEnumConstants()) {
            Object random=cls("dyz").getConstructor(long.class).newInstance(seed);
            Object origin=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(pos[0],pos[1],pos[2]);
            List<Object> pieces=new ArrayList<>();
            cls("emn").getMethod("a",cls("enu"),cls("jd"),cls("dmm"),List.class,cls("ayw")).invoke(null,manager,origin,rotation,pieces,random);
            StringBuilder saved=new StringBuilder();
            for(Object piece:pieces)saved.append(IglooProbeSupport.canonicalNBT(cls("ejv").getMethod("a",cls("ekh")).invoke(piece,(Object)null))).append('\n');
            String hash=HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(saved.toString().getBytes(StandardCharsets.UTF_8)));
            System.out.printf("mansion_layout %d %d %d %d %s %d %s %d%n",seed,pos[0],pos[1],pos[2],((Enum<?>)rotation).name(),pieces.size(),hash,cls("ayw").getMethod("g").invoke(random));
            if(seed==0 && pos[0]==0 && ((Enum<?>)rotation).ordinal()==0)System.err.print(saved);
        }
    }
}
