package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.lang.reflect.Method;

/**
 * Dumps vanilla's climate-based initial spawn search (Climate$SpawnFinder via
 * Climate$Sampler.findSpawnPosition) for fixed seeds, so the Go port can be
 * compared exactly. The search is deterministic per seed: it ring-samples the
 * overworld climate density functions and picks the chunk whose climate best
 * matches the noise settings' spawn_target.
 */
public final class SpawnSearchProbe {
    private static final long[] SEEDS = {
        0L, 1L, -1L, 1234L, 7122209284156038298L, 2173291061L, -417282937141413L, 9223372036854775807L,
        9212937586943976949L,
    };

    public static void main(String[] args) throws Exception {
        Object version = Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a", Class.forName("ae")).invoke(null, version);
        Class.forName("akt").getMethod("a").invoke(null); // Bootstrap registries.
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out), true));
        Object registries = Class.forName("pd").getMethod("a").invoke(null); // VanillaRegistries.createLookup()
        Object provider = Class.forName("jo$a").getMethod("b").invoke(registries); // asGetterLookup()
        Object overworldSettings = Class.forName("dzd").getField("c").get(null);
        Method create = Class.forName("dzm").getMethod("a", Class.forName("jn$a"), Class.forName("akq"), long.class);
        Method sampler = Class.forName("dzm").getMethod("b");
        Method findSpawnPosition = Class.forName("def$f").getMethod("a");
        Class<?> blockPos = Class.forName("jd");
        Method getX = blockPos.getMethod("u");
        Method getZ = blockPos.getMethod("w");
        for (long seed : SEEDS) {
            Object state = create.invoke(null, provider, overworldSettings, seed);
            Object position = findSpawnPosition.invoke(sampler.invoke(state));
            System.out.printf("spawn_search %d %d %d%n",
                seed, (int) getX.invoke(position), (int) getZ.invoke(position));
        }
    }
}
