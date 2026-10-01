package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.lang.reflect.Method;
import java.lang.reflect.Proxy;
import java.util.Comparator;
import java.util.HashMap;
import java.util.Map;
import java.util.Optional;

public final class BlueIceProbe {
    public static void main(String[] args) throws Exception {
        Object version = Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a", Class.forName("ae")).invoke(null, version);
        Class.forName("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out), true));

        Class<?> positionClass = Class.forName("jd");
        Class<?> blockClass = Class.forName("dfy");
        Method blockState = blockClass.getMethod("o");
        Object waterBlock = Class.forName("dga").getField("G").get(null);
        Object iceBlock = Class.forName("dga").getField("dO").get(null);
        Object packedBlock = Class.forName("dga").getField("iC").get(null);
        Object blueBlock = Class.forName("dga").getField("mW").get(null);
        Object registry = Class.forName("lt").getField("e").get(null);
        Method registryName = Class.forName("jz").getMethod("b", Object.class);
        if (!registryName.invoke(registry, waterBlock).toString().equals("minecraft:water")
            || !registryName.invoke(registry, iceBlock).toString().equals("minecraft:ice")
            || !registryName.invoke(registry, packedBlock).toString().equals("minecraft:packed_ice")
            || !registryName.invoke(registry, blueBlock).toString().equals("minecraft:blue_ice"))
            throw new IllegalStateException("blue ice block mapping changed");
        Object water = blockState.invoke(waterBlock);
        Object packed = blockState.invoke(packedBlock);
        Object air = blockState.invoke(Class.forName("dga").getField("a").get(null));
        Method xCoordinate = positionClass.getMethod("u");
        Method yCoordinate = positionClass.getMethod("v");
        Method zCoordinate = positionClass.getMethod("w");
        Method place = Class.forName("ece").getMethod("a", Class.forName("ecg"));
        Method nextLong = Class.forName("dzx").getMethod("g");
        Object feature = Class.forName("ece").getField("D").get(null);
        Object config = Class.forName("eep").getField("b").get(null);
        long[] seeds = {0L, 1L, -1L, 123456789L};
        int[][] cases = {{8, 40, 8, 1}, {15, 40, 15, 1}, {-8, 40, -8, 1},
            {8, 40, 8, 0}};
        for (long seed : seeds) {
            for (int[] entry : cases) {
                boolean supported = entry[3] != 0;
                Map<String, Object> changed = new HashMap<>();
                Object world = Proxy.newProxyInstance(Class.forName("dds").getClassLoader(),
                    new Class<?>[] {Class.forName("dds")}, (proxy, method, values) -> {
                        String name = method.getName();
                        if (name.equals("I_")) return -64;
                        if (name.equals("J_")) return 384;
                        if (name.equals("z_")) return 320;
                        if (name.equals("toString")) return "atlas-blue-ice-probe-world";
                        if (values != null && values.length > 0 && positionClass.isInstance(values[0])) {
                            int x = (int) xCoordinate.invoke(values[0]);
                            int y = (int) yCoordinate.invoke(values[0]);
                            int z = (int) zCoordinate.invoke(values[0]);
                            String key = x + " " + y + " " + z;
                            if (name.equals("a_")) {
                                if (changed.containsKey(key)) return changed.get(key);
                                if (supported && x == entry[0] + 1 && y == entry[1]
                                    && z == entry[2]) return packed;
                                return y >= 30 && y <= 61 ? water : air;
                            }
                            if (name.equals("a") && values.length == 3) {
                                changed.put(key, values[1]);
                                return true;
                            }
                        }
                        throw new UnsupportedOperationException(method.toString());
                    });
                Object random = Class.forName("dzx").getConstructor(Class.forName("ayw"))
                    .newInstance(Class.forName("dzz").getConstructor(long.class).newInstance(seed));
                Object origin = positionClass.getConstructor(int.class, int.class, int.class)
                    .newInstance(entry[0], entry[1], entry[2]);
                Object context = Class.forName("ecg").getConstructor(Optional.class,
                    Class.forName("dds"), Class.forName("duz"), Class.forName("ayw"),
                    positionClass, Class.forName("eei"))
                    .newInstance(Optional.empty(), world, null, random, origin, config);
                boolean placed = (boolean) place.invoke(feature, context);
                System.out.printf("blue_count %d %d %d %d %d %d %d %d%n", seed,
                    entry[0], entry[1], entry[2], entry[3], placed ? 1 : 0,
                    changed.size(), nextLong.invoke(random));
                changed.keySet().stream().sorted(Comparator.comparingInt((String value) ->
                    Integer.parseInt(value.split(" ")[1])).thenComparingInt(value ->
                    Integer.parseInt(value.split(" ")[2])).thenComparingInt(value ->
                    Integer.parseInt(value.split(" ")[0]))).forEach(position ->
                    System.out.printf("blue_block %d %d %d %d %d %s%n", seed,
                        entry[0], entry[1], entry[2], entry[3], position));
            }
        }
    }
}
