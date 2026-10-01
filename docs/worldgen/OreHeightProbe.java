package atlasprobe;

import java.io.ByteArrayInputStream;
import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Base64;
import java.util.function.Predicate;
import java.util.zip.GZIPInputStream;

public final class OreHeightProbe {
    public static void main(String[] args) throws Exception {
        Object version = Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a", Class.forName("ae")).invoke(null, version);
        Class.forName("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out), true));
        Object heightmap = Class.forName("dyy$a").getField("c").get(null);
        @SuppressWarnings("unchecked")
        Predicate<Object> solid = (Predicate<Object>) Class.forName("dyy$a")
            .getMethod("e").invoke(heightmap);
        var decodeState = Class.forName("dfy").getMethod("a", int.class);
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
                        if (solid.test(decodeState.invoke(null, states.getShort(index * 2) & 65535))) {
                            height = blockY + 1;
                            break;
                        }
                    }
                    System.out.printf("ore_height %s %s %s %d %d %d%n", fields[1], fields[2],
                        fields[3], blockX, blockZ, height);
                }
            }
        }
    }
}
