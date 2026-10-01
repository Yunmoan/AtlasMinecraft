package atlasprobe;

import com.mojang.serialization.Lifecycle;
import com.mojang.datafixers.util.Pair;
import com.google.gson.JsonParser;
import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.lang.reflect.Method;
import java.lang.reflect.Proxy;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Base64;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.LinkedHashSet;
import java.util.Set;
import java.util.function.Consumer;
import java.util.function.Function;
import java.util.stream.Stream;
import java.util.zip.GZIPInputStream;
import java.util.zip.GZIPOutputStream;
import java.util.zip.ZipFile;

public final class SurfaceChunkProbe {
    public static void main(String[] args) throws Exception {
        Object version = Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a", Class.forName("ae")).invoke(null, version);
        Class.forName("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out), true));
        Object registries = Class.forName("pd").getMethod("a").invoke(null);
        Object lookup = Class.forName("jo$a").getMethod("b").invoke(registries);
        Method lookupRegistry = Class.forName("jn$a").getMethod("b", Class.forName("akq"));
        Object settingsKey = Class.forName("dzd").getField("c").get(null);
        Object settingsRegistry = lookupRegistry.invoke(lookup, Class.forName("lu").getField("aO").get(null));
        Object settingsHolder = Class.forName("jn").getMethod("b", Class.forName("akq"))
            .invoke(settingsRegistry, settingsKey);
        Object settings = Class.forName("jm").getMethod("a").invoke(settingsHolder);
        Object noiseSettings = Class.forName("dzd").getMethod("f").invoke(settings);
        Object biomeLookup = lookupRegistry.invoke(lookup, Class.forName("lu").getField("aF").get(null));
        Map<String, Object> holders = new HashMap<>();
        Object biomeRegistry = Class.forName("ju").getConstructor(Class.forName("akq"), Lifecycle.class)
            .newInstance(Class.forName("lu").getField("aF").get(null), Lifecycle.stable());
        Method register = Class.forName("jz").getMethod("a", Class.forName("jz"), Class.forName("akq"), Object.class);
        for (Object holder : ((Stream<?>) Class.forName("jo").getMethod("b").invoke(biomeLookup)).toList()) {
            Object key = Class.forName("jm$c").getMethod("h").invoke(holder);
            holders.put(Class.forName("akq").getMethod("a").invoke(key).toString(), holder);
            register.invoke(null, biomeRegistry, key, Class.forName("jm").getMethod("a").invoke(holder));
        }
        if (holders.size() != 64) throw new IllegalStateException("unexpected biome registry");
        Object registryAccess = Proxy.newProxyInstance(Class.forName("ka").getClassLoader(), new Class<?>[]{Class.forName("ka")},
            (proxy, method, arguments) -> {
                if (arguments != null && arguments.length == 1 && arguments[0].equals(Class.forName("lu").getField("aF").get(null))) {
                    if (method.getName().equals("d")) return biomeRegistry;
                    if (method.getName().equals("c")) return java.util.Optional.of(biomeRegistry);
                }
                throw new UnsupportedOperationException(method.toString());
            });
        List<Pair<Object, Object>> points = new ArrayList<>();
        Method addPoints = Class.forName("dem").getDeclaredMethod("a", Consumer.class);
        addPoints.setAccessible(true);
        addPoints.invoke(Class.forName("dem").getConstructor().newInstance(),
            (Consumer<Pair<Object, Object>>) pair -> {
                try {
                    String name = Class.forName("akq").getMethod("a").invoke(pair.getSecond()).toString();
                    points.add(Pair.of(pair.getFirst(), holders.get(name)));
                } catch (ReflectiveOperationException error) {
                    throw new RuntimeException(error);
                }
            });
        Object parameterList = Class.forName("def$c").getConstructor(List.class).newInstance(points);
        Object naturalBiomeSource = Class.forName("dej").getMethod("a", Class.forName("def$c"))
            .invoke(null, parameterList);
        Method naturalBiomeAt = Class.forName("dej").getMethod("getNoiseBiome", int.class, int.class,
            int.class, Class.forName("def$f"));
        boolean carved = args.length > 1 && args[1].equals("carved");
        Map<String, Object> carvers = new HashMap<>();
        if (carved) {
            Object carverLookup = lookupRegistry.invoke(lookup, Class.forName("lu").getField("aH").get(null));
            for (Object holder : ((Stream<?>) Class.forName("jo").getMethod("b").invoke(carverLookup)).toList()) {
                Object key = Class.forName("jm$c").getMethod("h").invoke(holder);
                carvers.put(Class.forName("akq").getMethod("a").invoke(key).toString(),
                    Class.forName("jm").getMethod("a").invoke(holder));
            }
            Set<String> replacementNames = new LinkedHashSet<>();
            try (ZipFile server = new ZipFile(args[2])) {
                expandTag(server, "overworld_carver_replaceables", replacementNames);
            }
            Object blocks = Class.forName("lt").getField("e").get(null);
            List<Object> replacementHolders = new ArrayList<>();
            Method location = Class.forName("akr").getMethod("a", String.class);
            Method blockHolder = Class.forName("jz").getMethod("c", Class.forName("akr"));
            for (String name : replacementNames) {
                replacementHolders.add(((java.util.Optional<?>) blockHolder.invoke(blocks,
                    location.invoke(null, name))).orElseThrow());
            }
            Object directReplacements = Class.forName("jq").getMethod("a", List.class)
                .invoke(null, replacementHolders);
            var replaceableField = Class.forName("eax").getField("i");
            replaceableField.setAccessible(true);
            for (String name : new String[] {"minecraft:cave", "minecraft:cave_extra_underground", "minecraft:canyon"}) {
                Object carver = carvers.get(name);
                Object config = Class.forName("ebc").getMethod("b").invoke(carver);
                replaceableField.set(config, directReplacements);
                Object stone = Class.forName("dga").getField("b").get(null);
                Object stoneState = stone.getClass().getMethod("o").invoke(stone);
                if (!(Boolean) Class.forName("dtc").getMethod("a", Class.forName("jq"))
                    .invoke(stoneState, directReplacements)) throw new IllegalStateException("stone not replaceable");
            }
        }
        Object heights = Class.forName("dcy").getMethod("e", int.class, int.class).invoke(null, -64, 384);
        Object blender = Class.forName("eaa").getMethod("a").invoke(null);
        Object beardifier = Class.forName("dys$b").getField("a").get(null);
        Method fluidFactory = Class.forName("dzb").getDeclaredMethod("a", Class.forName("dzd"));
        fluidFactory.setAccessible(true);
        Object fluidPicker = fluidFactory.invoke(null, settings);
        Method decodeState = Class.forName("dfy").getMethod("a", int.class);
        Method stateID = Class.forName("dfy").getMethod("i", Class.forName("dtc"));
        Method getState = Class.forName("dvs").getMethod("a_", Class.forName("jd"));
        Method setState = Class.forName("dvs").getMethod("a", Class.forName("jd"), Class.forName("dtc"), boolean.class);
        Method applySurface = Class.forName("dzq").getMethod("a", Class.forName("dzm"), Class.forName("ddy"),
            Class.forName("jz"), boolean.class, Class.forName("dzv"), Class.forName("duy"),
            Class.forName("dzc"), Class.forName("dzp$o"));
        Class<?> getterClass = Class.forName("ddy$a");
        String[] biomes = {"minecraft:plains", "minecraft:desert", "minecraft:badlands",
            "minecraft:wooded_badlands", "minecraft:eroded_badlands", "minecraft:frozen_ocean",
            "minecraft:deep_frozen_ocean", "minecraft:snowy_slopes", "natural"};
        if (args.length > 1 && args[1].equals("natural")) biomes = new String[]{"natural"};
        List<String> lines = Files.readAllLines(Path.of(args[0]), StandardCharsets.UTF_8);
        for (String line : lines) {
            if (!line.startsWith("chunk ")) continue;
            String[] fields = line.split(" ");
            long seed = Long.parseLong(fields[1]);
            int cx = Integer.parseInt(fields[2]), cz = Integer.parseInt(fields[3]);
            byte[] compressed = Base64.getDecoder().decode(fields[4]);
            byte[] blocks;
            try (GZIPInputStream gzip = new GZIPInputStream(new ByteArrayInputStream(compressed))) {
                blocks = gzip.readAllBytes();
            }
            Object randomState = Class.forName("dzm").getMethod("a", Class.forName("jn$a"),
                Class.forName("akq"), long.class).invoke(null, lookup, settingsKey, seed);
            Object climateSampler = Class.forName("dzm").getMethod("b").invoke(randomState);
            Object surface = Class.forName("dzm").getMethod("c").invoke(randomState);
            for (String biomeName : biomes) {
                if (carved && !biomeName.equals("natural")) continue;
                int carveStarts = 0, carveReturns = 0, maskBits = 0;
                int[] maskByCarver = new int[3];
                Object carvingMask = null;
                Object biome = holders.get(biomeName.equals("natural") ? "minecraft:plains" : biomeName);
                Object biomeSource = Class.forName("deh").getConstructor(Class.forName("jm")).newInstance(biome);
                Object generator = Class.forName("dzb").getConstructor(Class.forName("dea"), Class.forName("jm"))
                    .newInstance(biomeSource, settingsHolder);
                Object generationContext = Class.forName("dzv").getConstructor(Class.forName("duz"), Class.forName("dcy"))
                    .newInstance(generator, heights);
                Object position = Class.forName("dcd").getConstructor(int.class, int.class).newInstance(cx, cz);
                Object protoChunk = Class.forName("dvs").getConstructor(Class.forName("dcd"), Class.forName("dvv"),
                    Class.forName("dcy"), Class.forName("jz"), Class.forName("eab"))
                    .newInstance(position, Class.forName("dvv").getField("a").get(null), heights, biomeRegistry, null);
                for (int blockY = -64; blockY < 320; blockY++) {
                    for (int blockZ = 0; blockZ < 16; blockZ++) {
                        for (int blockX = 0; blockX < 16; blockX++) {
                            int index = ((blockY + 64) * 256 + blockZ * 16 + blockX) * 3;
                            int id = ((blocks[index] & 255) << 8) | (blocks[index + 1] & 255);
                            if (id == 0) continue;
                            Object blockPosition = Class.forName("jd").getConstructor(int.class, int.class, int.class)
                                .newInstance(cx * 16 + blockX, blockY, cz * 16 + blockZ);
                            setState.invoke(protoChunk, blockPosition, decodeState.invoke(null, id), false);
                        }
                    }
                }
                Object noiseChunk = Class.forName("dzc").getConstructor(int.class, Class.forName("dzm"), int.class,
                    int.class, Class.forName("dzg"), Class.forName("dys$c"), Class.forName("dzd"),
                    Class.forName("dyk$a"), Class.forName("eaa"))
                    .newInstance(4, randomState, cx * 16, cz * 16, noiseSettings, beardifier, settings,
                        fluidPicker, blender);
                Object fixedBiome = Proxy.newProxyInstance(getterClass.getClassLoader(), new Class<?>[] {getterClass},
                    (proxy, method, arguments) -> biomeName.equals("natural")
                        ? naturalBiomeAt.invoke(naturalBiomeSource, arguments[0], arguments[1], arguments[2], climateSampler)
                        : biome);
                long biomeSeed = (Long) Class.forName("ddy").getMethod("a", long.class).invoke(null, seed);
                Object manager = Class.forName("ddy").getConstructor(getterClass, long.class).newInstance(fixedBiome, biomeSeed);
                applySurface.invoke(surface, randomState, manager, biomeRegistry, false, generationContext,
                    protoChunk, noiseChunk, Class.forName("dzd").getMethod("j").invoke(settings));
                if (carved) {
                    Object carvingContext = Class.forName("eaz").getConstructor(Class.forName("dzb"),
                        Class.forName("ka"), Class.forName("dcy"), Class.forName("dzc"),
                        Class.forName("dzm"), Class.forName("dzp$o"))
                        .newInstance(generator, registryAccess, heights, noiseChunk, randomState,
                            Class.forName("dzd").getMethod("j").invoke(settings));
                    Object mask = Class.forName("dux").getConstructor(int.class, int.class).newInstance(384, -64);
                    carvingMask = mask;
                    String[] carverNames = {"minecraft:cave", "minecraft:cave_extra_underground", "minecraft:canyon"};
                    Method seedCarver = Class.forName("dzx").getMethod("c", long.class, int.class, int.class);
                    Method isStart = Class.forName("ebc").getMethod("a", Class.forName("ayw"));
                    Method carve = Class.forName("ebc").getMethod("a", Class.forName("eaz"), Class.forName("duy"),
                        Function.class, Class.forName("ayw"), Class.forName("dyk"), Class.forName("dcd"), Class.forName("dux"));
                    Function<Object, Object> biomeAtBlock = blockPosition -> {
                        try {
                            return Class.forName("ddy").getMethod("a", Class.forName("jd"))
                                .invoke(manager, blockPosition);
                        } catch (ReflectiveOperationException error) {
                            throw new RuntimeException(error);
                        }
                    };
                    Object aquifer = Class.forName("dzc").getMethod("i").invoke(noiseChunk);
                    for (int dx = -8; dx <= 8; dx++) {
                        for (int dz = -8; dz <= 8; dz++) {
                            int sourceX = cx + dx, sourceZ = cz + dz;
                            Object sourcePosition = Class.forName("dcd").getConstructor(int.class, int.class)
                                .newInstance(sourceX, sourceZ);
                            for (int index = 0; index < carverNames.length; index++) {
                                Object random = Class.forName("dzx").getConstructor(Class.forName("ayw"))
                                    .newInstance(Class.forName("dyz").getConstructor(long.class).newInstance(0L));
                                seedCarver.invoke(random, seed + index, sourceX, sourceZ);
                                Object carver = carvers.get(carverNames[index]);
                                if ((Boolean) isStart.invoke(carver, random)) {
                                    carveStarts++;
                                    if ((Boolean) carve.invoke(carver, carvingContext, protoChunk, biomeAtBlock,
                                        random, aquifer, sourcePosition, mask)) carveReturns++;
                                    int newMaskBits = 0;
                                    for (long word : (long[]) Class.forName("dux").getMethod("a").invoke(mask))
                                        newMaskBits += Long.bitCount(word);
                                    maskByCarver[index] += newMaskBits - maskBits;
                                    maskBits = newMaskBits;
                                }
                            }
                        }
                    }
                }
                ByteBuffer output = ByteBuffer.allocate(16 * 16 * 384 * 2).order(ByteOrder.BIG_ENDIAN);
                for (int blockY = -64; blockY < 320; blockY++) {
                    for (int blockZ = 0; blockZ < 16; blockZ++) {
                        for (int blockX = 0; blockX < 16; blockX++) {
                            Object blockPosition = Class.forName("jd").getConstructor(int.class, int.class, int.class)
                                .newInstance(cx * 16 + blockX, blockY, cz * 16 + blockZ);
                            Object state = getState.invoke(protoChunk, blockPosition);
                            int index = ((blockY + 64) * 256 + blockZ * 16 + blockX) * 2;
                            output.putShort(index, ((Integer) stateID.invoke(null, state)).shortValue());
                        }
                    }
                }
                ByteArrayOutputStream compressedOutput = new ByteArrayOutputStream();
                try (GZIPOutputStream gzip = new GZIPOutputStream(compressedOutput)) {
                    gzip.write(output.array());
                }
                if (carved) {
                    ByteBuffer maskOutput = ByteBuffer.allocate(16 * 16 * 384 / 8).order(ByteOrder.LITTLE_ENDIAN);
                    long[] maskWords = (long[]) Class.forName("dux").getMethod("a").invoke(carvingMask);
                    for (int index = 0; index < maskWords.length; index++) maskOutput.putLong(index * 8, maskWords[index]);
                    ByteArrayOutputStream compressedMask = new ByteArrayOutputStream();
                    try (GZIPOutputStream gzip = new GZIPOutputStream(compressedMask)) {
                        gzip.write(maskOutput.array());
                    }
                    System.out.printf("carved_chunk %d %d %d %s %d %d %d %d %d %d %s %s%n", seed, cx, cz, biomeName,
                        carveStarts, carveReturns, maskBits, maskByCarver[0], maskByCarver[1], maskByCarver[2],
                        Base64.getEncoder().encodeToString(compressedMask.toByteArray()),
                        Base64.getEncoder().encodeToString(compressedOutput.toByteArray()));
                } else {
                    System.out.printf("surface_chunk %d %d %d %s %s%n", seed, cx, cz, biomeName,
                        Base64.getEncoder().encodeToString(compressedOutput.toByteArray()));
                }
            }
        }
    }

    private static void expandTag(ZipFile server, String tag, Set<String> blocks) throws Exception {
        String file = "data/minecraft/tags/block/" + tag + ".json";
        var values = JsonParser.parseString(new String(server.getInputStream(server.getEntry(file)).readAllBytes(),
            StandardCharsets.UTF_8)).getAsJsonObject().getAsJsonArray("values");
        for (var value : values) {
            String name = value.getAsString();
            if (name.startsWith("#minecraft:")) expandTag(server, name.substring(11), blocks);
            else blocks.add(name);
        }
    }
}
