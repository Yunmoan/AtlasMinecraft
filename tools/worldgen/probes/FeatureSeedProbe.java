package atlasprobe;

import java.lang.reflect.Method;

public final class FeatureSeedProbe {
    public static void main(String[] args) throws Exception {
        Class<?> randomClass = Class.forName("dzx");
        Object source = Class.forName("dzz").getConstructor(long.class).newInstance(0L);
        Object random = randomClass.getConstructor(Class.forName("ayw")).newInstance(source);
        Method decorationSeed = randomClass.getMethod("a", long.class, int.class, int.class);
        Method featureSeed = randomClass.getMethod("b", long.class, int.class, int.class);
        Method nextLong = randomClass.getMethod("g");
        Method nextBound = randomClass.getMethod("a", int.class);
        Method nextFloat = randomClass.getMethod("i");
        Method nextDouble = randomClass.getMethod("j");
        long[] worldSeeds = {0L, 1L, -1L, 123456789L, Long.MIN_VALUE};
        int[][] chunks = {{0, 0}, {1, -2}, {-17, 23}, {12345, -67890}};
        int[][] features = {{1, 0}, {6, 19}, {9, 78}, {10, 0}};
        for (long worldSeed : worldSeeds) {
            for (int[] chunk : chunks) {
                int blockX = chunk[0] * 16;
                int blockZ = chunk[1] * 16;
                long decoration = (long) decorationSeed.invoke(random, worldSeed, blockX, blockZ);
                System.out.printf("decoration %d %d %d %d%n", worldSeed, chunk[0], chunk[1], decoration);
                for (int[] feature : features) {
                    featureSeed.invoke(random, decoration, feature[1], feature[0]);
                    System.out.printf("feature_seed %d %d %d %d %d %d%n", worldSeed,
                        chunk[0], chunk[1], feature[0], feature[1], nextLong.invoke(random));
                    featureSeed.invoke(random, decoration, feature[1], feature[0]);
                    System.out.printf("feature_random %d %d %d %d %d %d %d %d%n", worldSeed,
                        chunk[0], chunk[1], feature[0], feature[1], nextBound.invoke(random, 37),
                        Float.floatToRawIntBits((float) nextFloat.invoke(random)),
                        Double.doubleToRawLongBits((double) nextDouble.invoke(random)));
                }
            }
        }
    }
}
