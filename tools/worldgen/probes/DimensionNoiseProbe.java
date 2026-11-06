package atlasprobe;

import java.io.ByteArrayOutputStream;
import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.lang.reflect.Method;
import java.lang.reflect.Field;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.util.Base64;
import java.util.List;
import java.util.zip.GZIPOutputStream;

public final class DimensionNoiseProbe {
    public static void main(String[] args) throws Exception {
        Object version = Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a", Class.forName("ae")).invoke(null, version);
        Class.forName("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out), true));
        Object registries = Class.forName("pd").getMethod("a").invoke(null);
        Object lookup = Class.forName("jo$a").getMethod("b").invoke(registries);
        Object settingsKey = Class.forName("dzd").getField(args[0].equals("nether") ? "f" : "g").get(null);
        Object settingsRegistry = Class.forName("jn$a").getMethod("b", Class.forName("akq"))
            .invoke(lookup, Class.forName("lu").getField("aO").get(null));
        Object settingsHolder = Class.forName("jn").getMethod("b", Class.forName("akq"))
            .invoke(settingsRegistry, settingsKey);
        Object settings = Class.forName("jm").getMethod("a").invoke(settingsHolder);
        Object noiseSettings = Class.forName("dzd").getMethod("f").invoke(settings);
        int minY=(Integer)Class.forName("dzg").getMethod("c").invoke(noiseSettings);
        int height=(Integer)Class.forName("dzg").getMethod("d").invoke(noiseSettings);
        int width=(Integer)Class.forName("dzg").getMethod("b").invoke(noiseSettings);
        int cellHeight=(Integer)Class.forName("dzg").getMethod("a").invoke(noiseSettings);
        Object defaultBlock = Class.forName("dzd").getMethod("g").invoke(settings);
        Method fluidFactory = Class.forName("dzb").getDeclaredMethod("a", Class.forName("dzd"));
        fluidFactory.setAccessible(true);
        Object fluidPicker = fluidFactory.invoke(null, settings);
        Object beardifier = Class.forName("dys$b").getField("a").get(null);
        Object blender = Class.forName("eaa").getMethod("a").invoke(null);
        Class<?> chunkClass = Class.forName("dzc");
        Class<?> densityClass = Class.forName("dyr");
        Field cellCaches = chunkClass.getDeclaredField("j");
        cellCaches.setAccessible(true);
        Method getState = chunkClass.getDeclaredMethod("e");
        getState.setAccessible(true);
        Method blockID = Class.forName("dfy").getMethod("i", Class.forName("dtc"));
        Method compute = densityClass.getMethod("a", Class.forName("dyr$b"));
        Method start = chunkClass.getMethod("f");
        Method advance = chunkClass.getMethod("b", int.class);
        Method selectCell = chunkClass.getMethod("b", int.class, int.class);
        Method updateY = chunkClass.getMethod("a", int.class, double.class);
        Method updateX = chunkClass.getMethod("b", int.class, double.class);
        Method updateZ = chunkClass.getMethod("c", int.class, double.class);
        Method swap = chunkClass.getMethod("h");
        Method stop = chunkClass.getMethod("g");
        Method getAquifer = chunkClass.getMethod("i");
        Method scheduleFluid = Class.forName("dyk").getMethod("a");
        long[] seeds = {0L, 1L, -1L, 7122209284156038298L};
        int[][] chunks = {{0, 0}, {-3, 2}, {11, -27}, {70,-90}, {-1874999,1874998}};
        java.util.List<long[]> cases = new java.util.ArrayList<>();
        if (args.length > 1) {
            for (String line : java.nio.file.Files.readAllLines(java.nio.file.Path.of(args[1]))) {
                if (line.isBlank() || line.startsWith("#")) continue;
                String[] values = line.trim().split("\\s+");
                if (values.length != 3) throw new IllegalArgumentException("expected seed cx cz: " + line);
                cases.add(new long[]{Long.parseLong(values[0]), Integer.parseInt(values[1]), Integer.parseInt(values[2])});
            }
        } else {
            for (long seed : seeds) for (int[] position : chunks) cases.add(new long[]{seed, position[0], position[1]});
        }
        for (long[] testCase : cases) {
            long seed = testCase[0];
            Object randomState = Class.forName("dzm").getMethod("a", Class.forName("jn$a"),
                Class.forName("akq"), long.class).invoke(null, lookup, settingsKey, seed);
            {
                int chunkX = (int)testCase[1], chunkZ = (int)testCase[2];
                Object chunk = chunkClass.getConstructor(int.class, Class.forName("dzm"), int.class,
                    int.class, Class.forName("dzg"), Class.forName("dys$c"), Class.forName("dzd"),
                    Class.forName("dyk$a"), Class.forName("eaa"))
                    .newInstance(16/width, randomState, chunkX * 16, chunkZ * 16, noiseSettings,
                        beardifier, settings, fluidPicker, blender);
                List<?> caches = (List<?>) cellCaches.get(chunk);
                if (caches.size() != 1) throw new IllegalStateException("unexpected cell cache count");
                Object density = caches.get(0);
                Object aquifer = getAquifer.invoke(chunk);
                ByteBuffer blocks = ByteBuffer.allocate(16 * 16 * height * 3).order(ByteOrder.BIG_ENDIAN);
                start.invoke(chunk);
                for (int cellX = 0; cellX < 16/width; cellX++) {
                    advance.invoke(chunk, cellX);
                    for (int cellZ = 0; cellZ < 16/width; cellZ++) {
                        for (int cellY = height/cellHeight-1; cellY >= 0; cellY--) {
                            selectCell.invoke(chunk, cellY, cellZ);
                            for (int localY = cellHeight-1; localY >= 0; localY--) {
                                int blockY = minY + cellY * cellHeight + localY;
                                updateY.invoke(chunk, blockY, localY / (double)cellHeight);
                                for (int localX = 0; localX < width; localX++) {
                                    int blockX = cellX * width + localX;
                                    updateX.invoke(chunk, chunkX * 16 + blockX, localX / (double)width);
                                    for (int localZ = 0; localZ < width; localZ++) {
                                        int blockZ = cellZ * width + localZ;
                                        updateZ.invoke(chunk, chunkZ * 16 + blockZ, localZ / (double)width);
                                        Object state = getState.invoke(chunk);
                                        if (state == null) state = defaultBlock;
                                        int index = ((blockY - minY) * 256 + blockZ * 16 + blockX) * 3;
                                        blocks.putShort(index, ((Integer) blockID.invoke(null, state)).shortValue());
                                        blocks.put(index + 2, (byte) ((Boolean) scheduleFluid.invoke(aquifer) ? 1 : 0));
                                        if (blockX == 5 && blockZ == 9 && blockY % 64 == 0) {
                                            double value = (Double) compute.invoke(density, chunk);
                                            System.out.printf("density %d %d %d %d %s%n", seed, chunkX, chunkZ,
                                                blockY, Long.toUnsignedString(Double.doubleToRawLongBits(value)));
                                        }
                                    }
                                }
                            }
                        }
                    }
                    swap.invoke(chunk);
                }
                stop.invoke(chunk);
                ByteArrayOutputStream compressed = new ByteArrayOutputStream();
                try (GZIPOutputStream gzip = new GZIPOutputStream(compressed)) {
                    gzip.write(blocks.array());
                }
                System.out.printf("chunk %d %d %d %s%n", seed, chunkX, chunkZ,
                    Base64.getEncoder().encodeToString(compressed.toByteArray()));
            }
        }
    }
}
