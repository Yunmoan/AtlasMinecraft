package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.lang.reflect.Method;
import java.util.HashMap;
import java.util.Map;
import java.util.stream.Stream;

public final class CarverStartProbe {
    public static void main(String[] args) throws Exception {
        Object version = Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a", Class.forName("ae")).invoke(null, version);
        Class.forName("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out), true));
        Object registries = Class.forName("pd").getMethod("a").invoke(null);
        Object lookup = Class.forName("jo$a").getMethod("b").invoke(registries);
        Object registry = Class.forName("jn$a").getMethod("b", Class.forName("akq"))
            .invoke(lookup, Class.forName("lu").getField("aH").get(null));
        Map<String, Object> carvers = new HashMap<>();
        for (Object holder : ((Stream<?>) Class.forName("jo").getMethod("b").invoke(registry)).toList()) {
            Object key = Class.forName("jm$c").getMethod("h").invoke(holder);
            String name = Class.forName("akq").getMethod("a").invoke(key).toString();
            carvers.put(name, Class.forName("jm").getMethod("a").invoke(holder));
        }
        String[] names = {"minecraft:cave", "minecraft:cave_extra_underground", "minecraft:canyon"};
        Method seedCarver = Class.forName("dzx").getMethod("c", long.class, int.class, int.class);
        Method isStart = Class.forName("ebc").getMethod("a", Class.forName("ayw"));
        Method nextLong = Class.forName("dzx").getMethod("g");
        int[][] targets = {{0, 0}, {17, -9}, {-41, 23}};
        long[] seeds = {0, 1, -9223372036854775807L};
        for (long seed : seeds) {
            for (int[] target : targets) {
                for (int dx = -8; dx <= 8; dx++) {
                    for (int dz = -8; dz <= 8; dz++) {
                        int sourceX = target[0] + dx, sourceZ = target[1] + dz;
                        for (int index = 0; index < names.length; index++) {
                            Object random = Class.forName("dzx").getConstructor(Class.forName("ayw"))
                                .newInstance(Class.forName("dyz").getConstructor(long.class).newInstance(0L));
                            seedCarver.invoke(random, seed + index, sourceX, sourceZ);
                            boolean started = (Boolean) isStart.invoke(carvers.get(names[index]), random);
                            if (started) {
                                System.out.printf("start %d %d %d %d %d %d %d %d%n", seed, target[0], target[1],
                                    sourceX, sourceZ, index, 1, (Long) nextLong.invoke(random));
                            }
                        }
                    }
                }
            }
        }
    }
}
