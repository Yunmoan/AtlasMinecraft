/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.datafixers.kinds.App
 *  com.mojang.datafixers.kinds.Applicative
 *  com.mojang.serialization.Codec
 *  com.mojang.serialization.codecs.RecordCodecBuilder
 */
import com.mojang.datafixers.kinds.App;
import com.mojang.datafixers.kinds.Applicative;
import com.mojang.serialization.Codec;
import com.mojang.serialization.codecs.RecordCodecBuilder;
import java.lang.invoke.MethodHandle;
import java.lang.runtime.ObjectMethods;

@Deprecated
public class ect
extends ece<a> {
    private static final dtc a = dga.nc.o();

    public ect(Codec<a> $$0) {
        super($$0);
    }

    @Override
    public boolean a(ecg<a> $$0) {
        jd $$1 = $$0.e();
        dds $$2 = $$0.b();
        ayw $$3 = $$0.d();
        a $$4 = $$0.f();
        if ($$1.v() <= $$2.I_() + 4) {
            return false;
        }
        $$1 = $$1.c(4);
        boolean[] $$5 = new boolean[2048];
        int $$6 = $$3.a(4) + 4;
        for (int $$7 = 0; $$7 < $$6; ++$$7) {
            double $$8 = $$3.j() * 6.0 + 3.0;
            double $$9 = $$3.j() * 4.0 + 2.0;
            double $$10 = $$3.j() * 6.0 + 3.0;
            double $$11 = $$3.j() * (16.0 - $$8 - 2.0) + 1.0 + $$8 / 2.0;
            double $$12 = $$3.j() * (8.0 - $$9 - 4.0) + 2.0 + $$9 / 2.0;
            double $$13 = $$3.j() * (16.0 - $$10 - 2.0) + 1.0 + $$10 / 2.0;
            for (int $$14 = 1; $$14 < 15; ++$$14) {
                for (int $$15 = 1; $$15 < 15; ++$$15) {
                    for (int $$16 = 1; $$16 < 7; ++$$16) {
                        double $$17 = ((double)$$14 - $$11) / ($$8 / 2.0);
                        double $$18 = ((double)$$16 - $$12) / ($$9 / 2.0);
                        double $$19 = ((double)$$15 - $$13) / ($$10 / 2.0);
                        double $$20 = $$17 * $$17 + $$18 * $$18 + $$19 * $$19;
                        if (!($$20 < 1.0)) continue;
                        $$5[($$14 * 16 + $$15) * 8 + $$16] = true;
                    }
                }
            }
        }
        dtc $$21 = $$4.a().a($$3, $$1);
        for (int $$22 = 0; $$22 < 16; ++$$22) {
            for (int $$23 = 0; $$23 < 16; ++$$23) {
                for (int $$24 = 0; $$24 < 8; ++$$24) {
                    boolean $$25;
                    boolean bl = $$25 = !$$5[($$22 * 16 + $$23) * 8 + $$24] && ($$22 < 15 && $$5[(($$22 + 1) * 16 + $$23) * 8 + $$24] || $$22 > 0 && $$5[(($$22 - 1) * 16 + $$23) * 8 + $$24] || $$23 < 15 && $$5[($$22 * 16 + $$23 + 1) * 8 + $$24] || $$23 > 0 && $$5[($$22 * 16 + ($$23 - 1)) * 8 + $$24] || $$24 < 7 && $$5[($$22 * 16 + $$23) * 8 + $$24 + 1] || $$24 > 0 && $$5[($$22 * 16 + $$23) * 8 + ($$24 - 1)]);
                    if (!$$25) continue;
                    dtc $$26 = $$2.a_($$1.b($$22, $$24, $$23));
                    if ($$24 >= 4 && $$26.k()) {
                        return false;
                    }
                    if ($$24 >= 4 || $$26.e() || $$2.a_($$1.b($$22, $$24, $$23)) == $$21) continue;
                    return false;
                }
            }
        }
        for (int $$27 = 0; $$27 < 16; ++$$27) {
            for (int $$28 = 0; $$28 < 16; ++$$28) {
                for (int $$29 = 0; $$29 < 8; ++$$29) {
                    jd $$30;
                    if (!$$5[($$27 * 16 + $$28) * 8 + $$29] || !this.c($$2.a_($$30 = $$1.b($$27, $$29, $$28)))) continue;
                    boolean $$31 = $$29 >= 4;
                    $$2.a($$30, $$31 ? a : $$21, 2);
                    if (!$$31) continue;
                    $$2.a($$30, a.b(), 0);
                    this.a($$2, $$30);
                }
            }
        }
        dtc $$32 = $$4.b().a($$3, $$1);
        if (!$$32.i()) {
            for (int $$33 = 0; $$33 < 16; ++$$33) {
                for (int $$34 = 0; $$34 < 16; ++$$34) {
                    for (int $$35 = 0; $$35 < 8; ++$$35) {
                        dtc $$37;
                        boolean $$36;
                        boolean bl = $$36 = !$$5[($$33 * 16 + $$34) * 8 + $$35] && ($$33 < 15 && $$5[(($$33 + 1) * 16 + $$34) * 8 + $$35] || $$33 > 0 && $$5[(($$33 - 1) * 16 + $$34) * 8 + $$35] || $$34 < 15 && $$5[($$33 * 16 + $$34 + 1) * 8 + $$35] || $$34 > 0 && $$5[($$33 * 16 + ($$34 - 1)) * 8 + $$35] || $$35 < 7 && $$5[($$33 * 16 + $$34) * 8 + $$35 + 1] || $$35 > 0 && $$5[($$33 * 16 + $$34) * 8 + ($$35 - 1)]);
                        if (!$$36 || $$35 >= 4 && $$3.a(2) == 0 || !($$37 = $$2.a_($$1.b($$33, $$35, $$34))).e() || $$37.a(awe.bQ)) continue;
                        jd $$38 = $$1.b($$33, $$35, $$34);
                        $$2.a($$38, $$32, 2);
                        this.a($$2, $$38);
                    }
                }
            }
        }
        if ($$21.u().a(awk.a)) {
            for (int $$39 = 0; $$39 < 16; ++$$39) {
                for (int $$40 = 0; $$40 < 16; ++$$40) {
                    int $$41 = 4;
                    jd $$42 = $$1.b($$39, 4, $$40);
                    if (!$$2.t($$42).a().a($$2, $$42, false) || !this.c($$2.a_($$42))) continue;
                    $$2.a($$42, dga.dO.o(), 2);
                }
            }
        }
        return true;
    }

    private boolean c(dtc $$0) {
        return !$$0.a(awe.bP);
    }

    public static final class a
    extends Record
    implements eei {
        private final egj b;
        private final egj c;
        public static final Codec<a> a = RecordCodecBuilder.create($$0 -> $$0.group((App)egj.a.fieldOf("fluid").forGetter(a::a), (App)egj.a.fieldOf("barrier").forGetter(a::b)).apply((Applicative)$$0, a::new));

        public a(egj $$0, egj $$1) {
            this.b = $$0;
            this.c = $$1;
        }

        @Override
        public final String toString() {
            return ObjectMethods.bootstrap("toString", new MethodHandle[]{a.class, "fluid;barrier", "b", "c"}, this);
        }

        @Override
        public final int hashCode() {
            return (int)ObjectMethods.bootstrap("hashCode", new MethodHandle[]{a.class, "fluid;barrier", "b", "c"}, this);
        }

        @Override
        public final boolean equals(Object $$0) {
            return (boolean)ObjectMethods.bootstrap("equals", new MethodHandle[]{a.class, "fluid;barrier", "b", "c"}, this, $$0);
        }

        public egj a() {
            return this.b;
        }

        public egj b() {
            return this.c;
        }
    }
}
