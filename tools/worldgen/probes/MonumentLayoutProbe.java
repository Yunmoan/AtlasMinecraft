package atlasprobe;

import java.io.*;
import java.lang.reflect.Field;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.*;

/** Reads the actual MonumentBuilding graph, fitted rooms and saved pieces.
 * No replacement layout algorithm runs in this oracle. */
public final class MonumentLayoutProbe {
    static Class<?> cls(String name) throws Exception { return Class.forName(name); }
    static Object field(Object value, String owner, String name) throws Exception {
        Field f = cls(owner).getDeclaredField(name); f.setAccessible(true); return f.get(value);
    }
    static String hash(String text) throws Exception {
        return HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(text.getBytes(StandardCharsets.UTF_8)));
    }
    static String describe(Object building) throws Exception {
        StringBuilder text = new StringBuilder();
        List<?> children = (List<?>) field(building, "emb$h", "I");
        List<Object> pieces = new ArrayList<>(); pieces.add(building); pieces.addAll(children);
        Map<Integer,Object> rooms = new TreeMap<>();
        ArrayDeque<Object> pending = new ArrayDeque<>();
        pending.add(field(building, "emb$h", "G"));
        while (!pending.isEmpty()) {
            Object room = pending.remove(); int id = (Integer) field(room, "emb$v", "a");
            if (rooms.putIfAbsent(id, room) != null) continue;
            for (Object neighbor : (Object[]) field(room, "emb$v", "b")) if (neighbor != null) pending.add(neighbor);
        }
        for (var entry : rooms.entrySet()) {
            Object room = entry.getValue();
            text.append("room ").append(entry.getKey());
            Object[] links = (Object[]) field(room, "emb$v", "b");
            boolean[] open = (boolean[]) field(room, "emb$v", "c");
            for (int d=0; d<6; d++) text.append(' ').append(links[d]==null ? -1 : field(links[d], "emb$v", "a")).append(':').append(open[d] ? 1 : 0);
            text.append(' ').append((Boolean)field(room,"emb$v","d") ? 1 : 0).append(' ').append((Boolean)field(room,"emb$v","e") ? 1 : 0).append('\n');
        }
        for (Object piece : pieces) {
            Object room = field(piece, "emb$r", "B");
            int variant = -1;
            if (piece.getClass()==cls("emb$s") || piece.getClass()==cls("emb$u")) variant=(Integer)field(piece,piece.getClass().getName(),"a");
            Object tag = cls("ejv").getMethod("a",cls("ekh")).invoke(piece,(Object)null);
            text.append("piece ").append(room==null ? -1 : field(room,"emb$v","a")).append(' ').append(variant).append(' ').append(IglooProbeSupport.canonicalNBT(tag)).append('\n');
        }
        return text.toString();
    }
    public static void main(String[] args) throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);
        cls("ab").getMethod("a",cls("ae")).invoke(null,version); cls("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        for (long seed : new long[]{0,1,-1,1234,Long.MIN_VALUE,Long.MAX_VALUE,4511451270791457643L})
        for (int[] pos : new int[][]{{0,0},{-29,-29},{243,-173},{-30000000,29999900}})
        for (String direction : new String[]{"north","east","south","west"}) {
            Object random=cls("dyz").getConstructor(long.class).newInstance(seed);
            Object facing=cls("ji").getMethod("a",String.class).invoke(null,direction);
            Object building=cls("emb$h").getConstructor(cls("ayw"),int.class,int.class,cls("ji")).newInstance(random,pos[0],pos[1],facing);
            String description=describe(building);
            long next=(Long)cls("ayw").getMethod("g").invoke(random);
            System.out.printf("monument_layout %d %d %d %s %d %s %d%n",seed,pos[0],pos[1],direction,((List<?>)field(building,"emb$h","I")).size(),hash(description),next);
            if (seed==0 && pos[0]==0 && direction.equals("north")) System.err.print(description);
        }
    }
}
