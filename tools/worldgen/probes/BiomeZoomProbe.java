package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.lang.reflect.Proxy;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Stream;

public final class BiomeZoomProbe {
    public static void main(String[] args) throws Exception {
        Object version = Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a", Class.forName("ae")).invoke(null, version);
        Class.forName("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out), true));
        Object registries = Class.forName("pd").getMethod("a").invoke(null);
        Object lookup = Class.forName("jo$a").getMethod("b").invoke(registries);
        Object biomeRegistry = Class.forName("jn$a").getMethod("b", Class.forName("akq"))
            .invoke(lookup, Class.forName("lu").getField("aF").get(null));
        List<?> holders = ((Stream<?>) Class.forName("jo").getMethod("b").invoke(biomeRegistry)).toList();
        long[] seeds = {0L, 1L, -1L, 7122209284156038298L};
        int[][] points = {{0, 64, 0}, {1, 64, 1}, {2, 64, 2}, {3, 64, 3},
            {4, 64, 4}, {5, 64, 5}, {8, 64, 8}, {15, -64, 15}, {-1, -63, -1},
            {-2, 0, -2}, {-3, 0, -3}, {-4, 0, -4}, {173, -32, -421},
            {-10000, 120, 24567}, {29999999, 319, -29999999}};
        Class<?> positionClass = Class.forName("jd");
        Class<?> getterClass = Class.forName("ddy$a");
        for (long seed : seeds) {
            long zoomSeed = (Long) Class.forName("ddy").getMethod("a", long.class).invoke(null, seed);
            for (int[] point : points) {
                int quartX = Math.floorDiv(point[0] - 2, 4);
                int quartY = Math.floorDiv(point[1] - 2, 4);
                int quartZ = Math.floorDiv(point[2] - 2, 4);
                Map<Object, int[]> positions = new HashMap<>();
                Map<String, Object> choices = new HashMap<>();
                for (int index = 0; index < 8; index++) {
                    int qx = quartX + ((index & 4) == 0 ? 0 : 1);
                    int qy = quartY + ((index & 2) == 0 ? 0 : 1);
                    int qz = quartZ + ((index & 1) == 0 ? 0 : 1);
                    Object holder = holders.get(index);
                    positions.put(holder, new int[] {qx, qy, qz});
                    choices.put(qx + "," + qy + "," + qz, holder);
                }
                Object getter = Proxy.newProxyInstance(getterClass.getClassLoader(), new Class<?>[] {getterClass},
                    (proxy, method, arguments) -> choices.get(arguments[0] + "," + arguments[1] + "," + arguments[2]));
                Object manager = Class.forName("ddy").getConstructor(getterClass, long.class).newInstance(getter, zoomSeed);
                Object position = positionClass.getConstructor(int.class, int.class, int.class)
                    .newInstance(point[0], point[1], point[2]);
                Object selected = Class.forName("ddy").getMethod("a", positionClass).invoke(manager, position);
                int[] expected = positions.get(selected);
                if (expected == null) throw new IllegalStateException("no quart position selected");
                System.out.printf("zoom %d %d %d %d %d %d %d%n", seed,
                    point[0], point[1], point[2], expected[0], expected[1], expected[2]);
            }
        }
    }
}
