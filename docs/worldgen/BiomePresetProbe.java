package atlasprobe;

import com.mojang.datafixers.util.Pair;
import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.util.ArrayList;
import java.util.List;
import java.util.Random;
import java.util.function.Consumer;

// Run only while developing Atlas. Prints the built-in 1.21.1 overworld
// climate parameter list and independent vanilla nearest-biome answers.
public final class BiomePresetProbe {
    public static void main(String[] args) throws Exception {
        Object version = Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a", Class.forName("ae")).invoke(null, version);
        Class.forName("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out), true));
        var points = new ArrayList<Pair<Object, Object>>();
        var addPoints = Class.forName("dem").getDeclaredMethod("a", Consumer.class);
        addPoints.setAccessible(true);
        addPoints.invoke(Class.forName("dem").getConstructor().newInstance(),
            (Consumer<Pair<Object, Object>>) points::add);
        var selector = Class.forName("def$c").getConstructor(List.class).newInstance(points);
        var treeField = selector.getClass().getDeclaredField("b");
        treeField.setAccessible(true);
        var tree = treeField.get(selector);
        var rootField = tree.getClass().getDeclaredField("b");
        rootField.setAccessible(true);
        printTree(rootField.get(tree));
        var pointClass = Class.forName("def$h");
        var pointCtor = pointClass.getConstructor(long.class, long.class, long.class, long.class, long.class, long.class);
        var nearest = selector.getClass().getMethod("a", pointClass);
        long[][] queries = {
            {0, 0, 0, 0, 0, 0},
            {8000, 2000, -4000, 3000, 1000, -5000},
            {-7000, -6000, 9000, -2000, 5000, 1000},
            {2000, 7000, -8000, 8000, -6000, 9000},
            {10000, 10000, 10000, 10000, 10000, 10000},
            {-10000, -10000, -10000, -10000, -10000, -10000},
        };
        for (var q : queries) {
            printQuery(selector, pointCtor, nearest, q);
        }
        var random = new Random(767L);
        for (int i = 0; i < 256; i++) {
            long[] q = new long[6];
            for (int j = 0; j < q.length; j++) q[j] = random.nextInt(24001) - 12000;
            printQuery(selector, pointCtor, nearest, q);
        }
        clearPrevious(selector);
        long[][] sequence = {
            {0, 0, 9000, 0, 5000, 0},
            {3652, 10195, 8501, 6629, 3278, -2759},
            {0, 8000, 0, 0, 5000, 0},
            {3652, 10195, 8501, 6629, 3278, -2759},
        };
        for (var q : sequence) printQueryResult(selector, pointCtor, nearest, q, "sequence");
    }

    private static void printTree(Object node) throws Exception {
        if (node.getClass().getName().equals("def$e$c")) {
            var childrenField = node.getClass().getDeclaredField("b");
            childrenField.setAccessible(true);
            for (Object child : (Object[]) childrenField.get(node)) printTree(child);
            return;
        }
        var rangesField = node.getClass().getSuperclass().getDeclaredField("a");
        rangesField.setAccessible(true);
        var ranges = (Object[]) rangesField.get(node);
        var valueField = node.getClass().getDeclaredField("b");
        valueField.setAccessible(true);
        var biome = valueField.get(node);
        var line = new StringBuilder("point ").append(biome.getClass().getMethod("a").invoke(biome));
        for (int i = 0; i < 6; i++) {
            line.append(' ').append(ranges[i].getClass().getMethod("a").invoke(ranges[i]));
            line.append(' ').append(ranges[i].getClass().getMethod("b").invoke(ranges[i]));
        }
        System.out.println(line.append(' ').append(ranges[6].getClass().getMethod("a").invoke(ranges[6])));
    }

    private static void printQuery(Object selector, java.lang.reflect.Constructor<?> pointCtor,
                                   java.lang.reflect.Method nearest, long[] q) throws Exception {
        // Clear Climate.ParameterList's per-thread last-biome cache so every
        // vector checks the tree's choice independently of query order.
        clearPrevious(selector);
        printQueryResult(selector, pointCtor, nearest, q, "query");
    }

    private static void clearPrevious(Object selector) throws Exception {
        var treeField = selector.getClass().getDeclaredField("b");
        treeField.setAccessible(true);
        var tree = treeField.get(selector);
        var cacheField = tree.getClass().getDeclaredField("c");
        cacheField.setAccessible(true);
        ((ThreadLocal<?>) cacheField.get(tree)).remove();
    }

    private static void printQueryResult(Object selector, java.lang.reflect.Constructor<?> pointCtor,
                                   java.lang.reflect.Method nearest, long[] q, String label) throws Exception {
        var point = pointCtor.newInstance(q[0], q[1], q[2], q[3], q[4], q[5]);
        var biome = nearest.invoke(selector, point);
        var line = new StringBuilder(label);
        for (long value : q) line.append(' ').append(value);
        System.out.println(line.append(' ').append(biome.getClass().getMethod("a").invoke(biome)));
    }
}
