/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.google.common.annotations.VisibleForTesting
 *  com.mojang.datafixers.kinds.App
 *  com.mojang.datafixers.kinds.Applicative
 *  com.mojang.serialization.Codec
 *  com.mojang.serialization.codecs.RecordCodecBuilder
 *  it.unimi.dsi.fastutil.doubles.DoubleArrayList
 *  it.unimi.dsi.fastutil.doubles.DoubleList
 *  it.unimi.dsi.fastutil.doubles.DoubleListIterator
 */
import com.google.common.annotations.VisibleForTesting;
import com.mojang.datafixers.kinds.App;
import com.mojang.datafixers.kinds.Applicative;
import com.mojang.serialization.Codec;
import com.mojang.serialization.codecs.RecordCodecBuilder;
import it.unimi.dsi.fastutil.doubles.DoubleArrayList;
import it.unimi.dsi.fastutil.doubles.DoubleList;
import it.unimi.dsi.fastutil.doubles.DoubleListIterator;
import java.lang.invoke.MethodHandle;
import java.lang.runtime.ObjectMethods;
import java.util.List;

public class eoh {
    private static final double a = 1.0181268882175227;
    private static final double b = 0.3333333333333333;
    private final double c;
    private final eoi d;
    private final eoi e;
    private final double f;
    private final a g;

    @Deprecated
    public static eoh a(ayw $$0, a $$1) {
        return new eoh($$0, $$1, false);
    }

    public static eoh a(ayw $$0, int $$1, double ... $$2) {
        return eoh.b($$0, new a($$1, (DoubleList)new DoubleArrayList($$2)));
    }

    public static eoh b(ayw $$0, a $$1) {
        return new eoh($$0, $$1, true);
    }

    private eoh(ayw $$0, a $$1, boolean $$2) {
        int $$3 = $$1.c;
        DoubleList $$4 = $$1.d;
        this.g = $$1;
        if ($$2) {
            this.d = eoi.b($$0, $$3, $$4);
            this.e = eoi.b($$0, $$3, $$4);
        } else {
            this.d = eoi.a($$0, $$3, $$4);
            this.e = eoi.a($$0, $$3, $$4);
        }
        int $$5 = Integer.MAX_VALUE;
        int $$6 = Integer.MIN_VALUE;
        DoubleListIterator $$7 = $$4.iterator();
        while ($$7.hasNext()) {
            int $$8 = $$7.nextIndex();
            double $$9 = $$7.nextDouble();
            if ($$9 == 0.0) continue;
            $$5 = Math.min($$5, $$8);
            $$6 = Math.max($$6, $$8);
        }
        this.c = 0.16666666666666666 / eoh.a($$6 - $$5);
        this.f = (this.d.a() + this.e.a()) * this.c;
    }

    public double a() {
        return this.f;
    }

    private static double a(int $$0) {
        return 0.1 * (1.0 + 1.0 / (double)($$0 + 1));
    }

    public double a(double $$0, double $$1, double $$2) {
        double $$3 = $$0 * 1.0181268882175227;
        double $$4 = $$1 * 1.0181268882175227;
        double $$5 = $$2 * 1.0181268882175227;
        return (this.d.a($$0, $$1, $$2) + this.e.a($$3, $$4, $$5)) * this.c;
    }

    public a b() {
        return this.g;
    }

    @VisibleForTesting
    public void a(StringBuilder $$0) {
        $$0.append("NormalNoise {");
        $$0.append("first: ");
        this.d.a($$0);
        $$0.append(", second: ");
        this.e.a($$0);
        $$0.append("}");
    }

    public static final class a
    extends Record {
        final int c;
        final DoubleList d;
        public static final Codec<a> a = RecordCodecBuilder.create($$0 -> $$0.group((App)Codec.INT.fieldOf("firstOctave").forGetter(a::a), (App)Codec.DOUBLE.listOf().fieldOf("amplitudes").forGetter(a::b)).apply((Applicative)$$0, a::new));
        public static final Codec<jm<a>> b = akn.a(lu.aP, a);

        public a(int $$0, List<Double> $$1) {
            this($$0, (DoubleList)new DoubleArrayList($$1));
        }

        public a(int $$0, double $$12, double ... $$2) {
            this($$0, (DoubleList)ad.a(new DoubleArrayList($$2), (? super T $$1) -> $$1.add(0, $$12)));
        }

        public a(int $$0, DoubleList $$1) {
            this.c = $$0;
            this.d = $$1;
        }

        @Override
        public final String toString() {
            return ObjectMethods.bootstrap("toString", new MethodHandle[]{a.class, "firstOctave;amplitudes", "c", "d"}, this);
        }

        @Override
        public final int hashCode() {
            return (int)ObjectMethods.bootstrap("hashCode", new MethodHandle[]{a.class, "firstOctave;amplitudes", "c", "d"}, this);
        }

        @Override
        public final boolean equals(Object $$0) {
            return (boolean)ObjectMethods.bootstrap("equals", new MethodHandle[]{a.class, "firstOctave;amplitudes", "c", "d"}, this, $$0);
        }

        public int a() {
            return this.c;
        }

        public DoubleList b() {
            return this.d;
        }
    }
}

