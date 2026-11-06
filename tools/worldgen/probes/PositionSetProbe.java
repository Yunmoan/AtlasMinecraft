package atlasprobe;
import java.lang.reflect.*;
import java.util.*;
import java.security.MessageDigest;
import java.nio.charset.StandardCharsets;

/** Actual HashSet<BlockPos> iteration, resize, tree-bin insertion and iterator removal. */
public final class PositionSetProbe {
    static Class<?> pos;
    static Set<Object> set;
    static int mode;
    static Object position(int id)throws Exception {
        int z=id-120,y=mode==2?id%23-11:0;
        // Collide in table buckets while keeping full hashes distinct.
        // BlockPos inherits Comparable<Vec3i>; HashMap does not recognize it
        // as Comparable<BlockPos>. Full-hash ties use JVM identity hashes and
        // cannot supply reproducible cross-process iteration vectors.
        int hash=mode==0?id*64:mode==1?id*128:mode==2?id*256:id*997;
        return pos.getConstructor(int.class,int.class,int.class).newInstance(hash-31*y-961*z,y,z);
    }
    static String coords(Object p)throws Exception { return pos.getMethod("u").invoke(p)+" "+pos.getMethod("v").invoke(p)+" "+pos.getMethod("w").invoke(p); }
    static void record(String op,Object p)throws Exception {
        List<String> values=new ArrayList<>();for(Object value:set)values.add(coords(value));
        String hash=HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(String.join("\n",values).getBytes(StandardCharsets.UTF_8)));
        System.out.printf("position_set %d %s %s %d %s%n",mode,op,coords(p),set.size(),hash);
    }
    static void add(int id)throws Exception { Object p=position(id);set.add(p);record("add",p); }
    static void pop()throws Exception { var iterator=set.iterator();Object p=iterator.next();iterator.remove();record("pop",p); }
    public static void main(String[] args)throws Exception {
        pos=Class.forName("jd");
        for(mode=0;mode<4;mode++) {
            set=new HashSet<>();
            for(int i=0;i<80;i++)add(i);
            for(int i=80;i<220;i++){if(i%3==0)pop();else add(i);}
            while(!set.isEmpty())pop();
            for(int i=240;i<260;i++)add(i);
            while(!set.isEmpty())pop();
        }
    }
}
