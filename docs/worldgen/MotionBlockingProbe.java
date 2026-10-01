package atlasprobe;

import java.io.ByteArrayInputStream;
import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.lang.reflect.Method;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Base64;
import java.util.Map;
import java.util.TreeMap;
import java.util.function.Predicate;
import java.util.zip.GZIPInputStream;

public final class MotionBlockingProbe {
    public static void main(String[] args) throws Exception {
        Object version = Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a", Class.forName("ae")).invoke(null, version);
        Class.forName("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out), true));
        Object heightmap = Class.forName("dyy$a").getField("e").get(null);
        Object oceanFloor = Class.forName("dyy$a").getField("c").get(null);
        @SuppressWarnings("unchecked")
        Predicate<Object> blocksMotion = (Predicate<Object>) Class.forName("dyy$a")
            .getMethod("e").invoke(heightmap);
        @SuppressWarnings("unchecked")
        Predicate<Object> oceanFloorSolid = (Predicate<Object>) Class.forName("dyy$a")
            .getMethod("e").invoke(oceanFloor);
        Method decodeState = Class.forName("dfy").getMethod("a", int.class);
        Method blockOfState = Class.forName("dtc").getMethod("b");
        Method propertiesOfState = Class.forName("dte").getMethod("C");
        Method propertyName = Class.forName("duf").getMethod("f");
        Object blockRegistry = Class.forName("lt").getField("e").get(null);
        Method blockName = Class.forName("jz").getMethod("b", Object.class);
        Method defaultState = Class.forName("dfy").getMethod("o");
        Map<String, int[]> blockKinds = new TreeMap<>();
        Map<String, int[]> oceanKinds = new TreeMap<>();
        byte[] flags = new byte[(26684 + 7) / 8];
        byte[] oceanFlags = new byte[flags.length];
        for (int stateID = 0; stateID < 26684; stateID++) {
            Object state = decodeState.invoke(null, stateID);
            boolean blocking = blocksMotion.test(state);
            Object block = blockOfState.invoke(state);
            String name = blockName.invoke(blockRegistry, block).toString();
            int[] counts = blockKinds.computeIfAbsent(name, ignored -> new int[3]);
            counts[0]++;
            if (blocking) counts[1]++;
            int[] oceanCounts = oceanKinds.computeIfAbsent(name, ignored -> new int[2]);
            oceanCounts[0]++;
            if (oceanFloorSolid.test(state)) oceanCounts[1]++;
            if (blocksMotion.test(defaultState.invoke(block))) counts[2] = 1;
            boolean waterlogged = false;
            for (Map.Entry<?, ?> property : ((Map<?, ?>) propertiesOfState.invoke(state)).entrySet()) {
                if (propertyName.invoke(property.getKey()).equals("waterlogged"))
                    waterlogged = property.getValue().equals(Boolean.TRUE);
            }
            if (blocking != waterlogged && counts[0] > 0) {
                counts[2] |= 2;
            }
            if (blocking)
                flags[stateID / 8] |= 1 << (stateID % 8);
            if (oceanFloorSolid.test(state))
                oceanFlags[stateID / 8] |= 1 << (stateID % 8);
        }
        System.out.println("motion_blocking_states 26684 " + Base64.getEncoder().encodeToString(flags));
        System.out.println("ocean_floor_states 26684 " + Base64.getEncoder().encodeToString(oceanFlags));
        for (Map.Entry<String, int[]> entry : blockKinds.entrySet()) {
            int[] counts = entry.getValue();
            String kind;
            if (counts[1] == 0) kind = "never";
            else if (counts[1] == counts[0]) kind = "always";
            else {
                if ((counts[2] & 2) != 0)
                    throw new IllegalStateException("mixed motion blocking without waterlogged rule: " + entry.getKey());
                kind = (counts[2] & 1) != 0 ? "waterlogged_default_true" : "waterlogged_default_false";
            }
            System.out.printf("motion_blocking_name %s %s%n", entry.getKey(), kind);
        }
        for (Map.Entry<String, int[]> entry : oceanKinds.entrySet()) {
            int[] counts = entry.getValue();
            if (counts[1] != 0 && counts[1] != counts[0])
                throw new IllegalStateException("mixed ocean floor states: " + entry.getKey());
            System.out.printf("ocean_floor_name %s %s%n", entry.getKey(),
                counts[1] == 0 ? "never" : "always");
        }
        for (String line : Files.readAllLines(Path.of(args[0]))) {
            if (!line.startsWith("carved_chunk ")) continue;
            String[] fields = line.split(" ");
            byte[] blocks;
            try (GZIPInputStream gzip = new GZIPInputStream(new ByteArrayInputStream(
                Base64.getDecoder().decode(fields[12])))) {
                blocks = gzip.readAllBytes();
            }
            ByteBuffer states = ByteBuffer.wrap(blocks).order(ByteOrder.BIG_ENDIAN);
            for (int blockX = 0; blockX < 16; blockX++) {
                for (int blockZ = 0; blockZ < 16; blockZ++) {
                    int height = -64;
                    for (int blockY = 319; blockY >= -64; blockY--) {
                        int index = (blockY + 64) * 256 + blockZ * 16 + blockX;
                        int stateID = states.getShort(index * 2) & 65535;
                        if ((flags[stateID / 8] & (1 << (stateID % 8))) != 0) {
                            height = blockY + 1;
                            break;
                        }
                    }
                    System.out.printf("motion_blocking_height %s %s %s %d %d %d%n",
                        fields[1], fields[2], fields[3], blockX, blockZ, height);
                }
            }
        }
    }
}
