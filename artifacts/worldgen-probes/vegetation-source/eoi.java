/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.google.common.annotations.VisibleForTesting
 *  com.google.common.collect.ImmutableList
 *  com.mojang.datafixers.util.Pair
 *  it.unimi.dsi.fastutil.doubles.DoubleArrayList
 *  it.unimi.dsi.fastutil.doubles.DoubleList
 *  it.unimi.dsi.fastutil.ints.IntBidirectionalIterator
 *  it.unimi.dsi.fastutil.ints.IntRBTreeSet
 *  it.unimi.dsi.fastutil.ints.IntSortedSet
 *  javax.annotation.Nullable
 */
import com.google.common.annotations.VisibleForTesting;
import com.google.common.collect.ImmutableList;
import com.mojang.datafixers.util.Pair;
import it.unimi.dsi.fastutil.doubles.DoubleArrayList;
import it.unimi.dsi.fastutil.doubles.DoubleList;
import it.unimi.dsi.fastutil.ints.IntBidirectionalIterator;
import it.unimi.dsi.fastutil.ints.IntRBTreeSet;
import it.unimi.dsi.fastutil.ints.IntSortedSet;
import java.util.Arrays;
import java.util.Collection;
import java.util.List;
import java.util.Locale;
import java.util.Objects;
import java.util.stream.IntStream;
import javax.annotation.Nullable;

public class eoi {
    private static final int a = 0x2000000;
    private final eof[] b;
    private final int c;
    private final DoubleList d;
    private final double e;
    private final double f;
    private final double g;

    @Deprecated
    public static eoi a(ayw $$0, IntStream $$1) {
        return new eoi($$0, eoi.a((IntSortedSet)new IntRBTreeSet((Collection)$$1.boxed().collect(ImmutableList.toImmutableList()))), false);
    }

    @Deprecated
    public static eoi a(ayw $$0, int $$1, DoubleList $$2) {
        return new eoi($$0, (Pair<Integer, DoubleList>)Pair.of((Object)$$1, (Object)$$2), false);
    }

    public static eoi b(ayw $$0, IntStream $$1) {
        return eoi.a($$0, (List)$$1.boxed().collect(ImmutableList.toImmutableList()));
    }

    public static eoi a(ayw $$0, List<Integer> $$1) {
        return new eoi($$0, eoi.a((IntSortedSet)new IntRBTreeSet($$1)), true);
    }

    public static eoi a(ayw $$0, int $$1, double $$2, double ... $$3) {
        DoubleArrayList $$4 = new DoubleArrayList($$3);
        $$4.add(0, $$2);
        return new eoi($$0, (Pair<Integer, DoubleList>)Pair.of((Object)$$1, (Object)$$4), true);
    }

    public static eoi b(ayw $$0, int $$1, DoubleList $$2) {
        return new eoi($$0, (Pair<Integer, DoubleList>)Pair.of((Object)$$1, (Object)$$2), true);
    }

    private static Pair<Integer, DoubleList> a(IntSortedSet $$0) {
        int $$2;
        if ($$0.isEmpty()) {
            throw new IllegalArgumentException("Need some octaves!");
        }
        int $$1 = -$$0.firstInt();
        int $$3 = $$1 + ($$2 = $$0.lastInt()) + 1;
        if ($$3 < 1) {
            throw new IllegalArgumentException("Total number of octaves needs to be >= 1");
        }
        DoubleArrayList $$4 = new DoubleArrayList(new double[$$3]);
        IntBidirectionalIterator $$5 = $$0.iterator();
        while ($$5.hasNext()) {
            int $$6 = $$5.nextInt();
            $$4.set($$6 + $$1, 1.0);
        }
        return Pair.of((Object)(-$$1), (Object)$$4);
    }

    protected eoi(ayw $$02, Pair<Integer, DoubleList> $$1, boolean $$2) {
        this.c = (Integer)$$1.getFirst();
        this.d = (DoubleList)$$1.getSecond();
        int $$3 = this.d.size();
        int $$4 = -this.c;
        this.b = new eof[$$3];
        if ($$2) {
            dzl $$5 = $$02.e();
            for (int $$6 = 0; $$6 < $$3; ++$$6) {
                if (this.d.getDouble($$6) == 0.0) continue;
                int $$7 = this.c + $$6;
                this.b[$$6] = new eof($$5.a("octave_" + $$7));
            }
        } else {
            double $$9;
            eof $$8 = new eof($$02);
            if ($$4 >= 0 && $$4 < $$3 && ($$9 = this.d.getDouble($$4)) != 0.0) {
                this.b[$$4] = $$8;
            }
            for (int $$10 = $$4 - 1; $$10 >= 0; --$$10) {
                if ($$10 < $$3) {
                    double $$11 = this.d.getDouble($$10);
                    if ($$11 != 0.0) {
                        this.b[$$10] = new eof($$02);
                        continue;
                    }
                    eoi.a($$02);
                    continue;
                }
                eoi.a($$02);
            }
            if (Arrays.stream(this.b).filter(Objects::nonNull).count() != this.d.stream().filter($$0 -> $$0 != 0.0).count()) {
                throw new IllegalStateException("Failed to create correct number of noise levels for given non-zero amplitudes");
            }
            if ($$4 < $$3 - 1) {
                throw new IllegalArgumentException("Positive octaves are temporarily disabled");
            }
        }
        this.f = Math.pow(2.0, -$$4);
        this.e = Math.pow(2.0, $$3 - 1) / (Math.pow(2.0, $$3) - 1.0);
        this.g = this.c(2.0);
    }

    protected double a() {
        return this.g;
    }

    private static void a(ayw $$0) {
        $$0.b(262);
    }

    public double a(double $$0, double $$1, double $$2) {
        return this.a($$0, $$1, $$2, 0.0, 0.0, false);
    }

    @Deprecated
    public double a(double $$0, double $$1, double $$2, double $$3, double $$4, boolean $$5) {
        double $$6 = 0.0;
        double $$7 = this.f;
        double $$8 = this.e;
        for (int $$9 = 0; $$9 < this.b.length; ++$$9) {
            eof $$10 = this.b[$$9];
            if ($$10 != null) {
                double $$11 = $$10.a(eoi.b($$0 * $$7), $$5 ? -$$10.b : eoi.b($$1 * $$7), eoi.b($$2 * $$7), $$3 * $$7, $$4 * $$7);
                $$6 += this.d.getDouble($$9) * $$11 * $$8;
            }
            $$7 *= 2.0;
            $$8 /= 2.0;
        }
        return $$6;
    }

    public double a(double $$0) {
        return this.c($$0 + 2.0);
    }

    private double c(double $$0) {
        double $$1 = 0.0;
        double $$2 = this.e;
        for (int $$3 = 0; $$3 < this.b.length; ++$$3) {
            eof $$4 = this.b[$$3];
            if ($$4 != null) {
                $$1 += this.d.getDouble($$3) * $$0 * $$2;
            }
            $$2 /= 2.0;
        }
        return $$1;
    }

    @Nullable
    public eof a(int $$0) {
        return this.b[this.b.length - 1 - $$0];
    }

    public static double b(double $$0) {
        return $$0 - (double)ayo.b($$0 / 3.3554432E7 + 0.5) * 3.3554432E7;
    }

    protected int b() {
        return this.c;
    }

    protected DoubleList c() {
        return this.d;
    }

    @VisibleForTesting
    public void a(StringBuilder $$02) {
        $$02.append("PerlinNoise{");
        List<String> $$1 = this.d.stream().map($$0 -> String.format(Locale.ROOT, "%.2f", $$0)).toList();
        $$02.append("first octave: ").append(this.c).append(", amplitudes: ").append($$1).append(", noise levels: [");
        for (int $$2 = 0; $$2 < this.b.length; ++$$2) {
            $$02.append($$2).append(": ");
            eof $$3 = this.b[$$2];
            if ($$3 == null) {
                $$02.append("null");
            } else {
                $$3.a($$02);
            }
            $$02.append(", ");
        }
        $$02.append("]");
        $$02.append("}");
    }
}

