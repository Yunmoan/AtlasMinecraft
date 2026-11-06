package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.nio.file.*;
import java.util.*;
import io.netty.buffer.*;
import com.mojang.brigadier.*;
import com.mojang.brigadier.tree.*;

/** Development-only reference: IDs come from runtime registration, not sorted names. */
public final class PlayerEntityProbe {
    public static void main(String[] args) throws Exception {
        Object version = Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a", Class.forName("ae")).invoke(null, version);
        Class.forName("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out), true));
        Object registry = Class.forName("lt").getField("f").get(null);
        Method getID = Class.forName("jz").getMethod("a", Object.class);
        Method getKey = Class.forName("jz").getMethod("b", Object.class);
        for (Object entity : (Iterable<?>) registry) {
            int id = (int) getID.invoke(registry, entity);
            String name = getKey.invoke(registry, entity).toString();
            if (id == 82 || name.equals("minecraft:player")) {
                System.out.println(id + " " + name);
            }
        }
        Class<?> blockPos = Class.forName("jd");
        Method spawnable = Class.forName("dcw").getMethod("l", blockPos);
        int[][] bounds = {{0, 1024, 0, 1}, {-30000000, -20000000, -30000000, 1}, {29999999, 19999999, 29999999, 1},
            {30000000, 0, 0, 0}, {-30000001, 0, 0, 0}, {0, 20000000, 0, 0}, {0, -20000001, 0, 0}, {1000000000, 64, 1000000000, 0}};
        for (int[] bound : bounds) {
            Object pos = blockPos.getConstructor(int.class, int.class, int.class).newInstance(bound[0], bound[1], bound[2]);
            if ((boolean) spawnable.invoke(null, pos) != (bound[3] != 0)) throw new IllegalStateException("unexpected vanilla teleport bounds: " + Arrays.toString(bound));
        }
        System.out.println("vanilla teleport bounds checked for " + bounds.length + " cases; billion-coordinate destination rejected");
        if (args.length == 0) return;
        Path fixtures = Path.of(args[0]);
        Object access = Class.forName("ka").getMethod("a", Class.forName("jz")).invoke(null, Class.forName("lt").getField("aA").get(null));
        for (String name : new String[]{"spawn", "commands", "chat", "tps", "join", "left"}) {
            ByteBuf raw = Unpooled.wrappedBuffer(Files.readAllBytes(fixtures.resolve(name + ".bin")));
            try {
                Object buffer = Class.forName("wk").getConstructor(ByteBuf.class, Class.forName("ka")).newInstance(raw, access);
                String className = name.equals("spawn") ? "abv" : name.equals("commands") ? "acn" : "afs";
                Object packet = Class.forName("yy").getMethod("decode", Object.class).invoke(Class.forName(className).getField("a").get(null), buffer);
                if (raw.isReadable()) throw new IllegalStateException(name + " has trailing bytes");
                if (name.equals("spawn")) {
                    Object entity = Class.forName("abv").getMethod("f").invoke(packet);
                    String key = getKey.invoke(registry, entity).toString();
                    if (!key.equals("minecraft:player")) throw new IllegalStateException("spawn decoded as " + key);
                    System.out.println("spawn packet decoded as " + key);
                } else if (!name.equals("commands")) {
                    Object component = Class.forName("afs").getMethod("b").invoke(packet);
                    String plain = (String) Class.forName("wz").getMethod("getString").invoke(component);
                    System.out.println(name + " component decoded: " + plain);
                    if (name.equals("chat") && !plain.equals("绿色 Atlas 彩色🚀")) throw new IllegalStateException("chat text changed");
					List<?> flat = (List<?>) Class.forName("wz").getMethod("h").invoke(component);
					StringBuilder styles = new StringBuilder();
					for (Object part : flat) styles.append(Class.forName("wz").getMethod("a").invoke(part)).append('\n');
					if (name.equals("chat") && !styles.toString().toLowerCase(Locale.ROOT).contains("12abcd")) throw new IllegalStateException("RGB style lost");
                } else {
					Object lookup = Class.forName("pd").getMethod("a").invoke(null);
					Object buildContext = Class.forName("ep").getMethod("a", Class.forName("jo$a"), Class.forName("cpl")).invoke(null, lookup, Class.forName("cpn").getField("f").get(null));
					@SuppressWarnings("unchecked")
					RootCommandNode<Object> root = (RootCommandNode<Object>) Class.forName("acn").getMethod("a", Class.forName("ep")).invoke(packet, buildContext);
					CommandDispatcher<Object> dispatcher = new CommandDispatcher<>(root);
					String[] valid = {"help", "help tp", "list", "seed", "op CodecProbe", "deop CodecProbe", "tps", "mspt", "status", "sysinfo", "time query", "time query daytime", "time query gametime", "time query day", "time set day", "time set 1d", "time add 20s", "weather clear", "weather rain", "weather thunder", "tp ~ ~ ~", "tp CodecProbe", "tp CodecProbe ~ ~1 ~", "teleport CodecProbe ^ ^ ^3", "minecraft:tp CodecProbe", "setblock ~ ~-1 ~ minecraft:oak_log[axis=x]", "setblock 0 -60 1 stone keep", "setblock 0 -60 1 1", "save", "save-all flush"};
					for (String command : valid) {
						ParseResults<Object> parsed = dispatcher.parse(command, new Object());
						if (parsed.getReader().canRead() || parsed.getContext().getCommand() == null) throw new IllegalStateException("vanilla rejected command: " + command + " " + parsed.getExceptions());
					}
					for (String command : new String[]{"tp", "time", "weather snow", "op", "list extra", "seed extra", "tp ^ ~ ^"}) {
						ParseResults<Object> parsed = dispatcher.parse(command, new Object());
						if (!parsed.getReader().canRead() && parsed.getContext().getCommand() != null) throw new IllegalStateException("vanilla accepted invalid command: " + command);
					}
					System.out.println("commands packet decoded; vanilla Brigadier accepted " + valid.length + " valid commands and rejected 7 invalid commands");
                }
            } finally { raw.release(); }
        }
    }
}
