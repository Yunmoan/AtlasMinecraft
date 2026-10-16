/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.serialization.Codec
 */
import com.mojang.serialization.Codec;
import java.util.function.Function;

public class ebb
extends ebe<eba> {
    public ebb(Codec<eba> $$0) {
        super($$0);
    }

    @Override
    public boolean a(eba $$0, ayw $$1) {
        return $$1.i() <= $$0.l;
    }

    @Override
    public boolean a(eaz $$0, eba $$12, duy $$22, Function<jd, jm<ddw>> $$32, ayw $$42, dyk $$52, dcd $$6, dux $$7) {
        int $$8 = kf.c(this.d() * 2 - 1);
        int $$9 = $$42.a($$42.a($$42.a(this.a()) + 1) + 1);
        for (int $$10 = 0; $$10 < $$9; ++$$10) {
            double $$11 = $$6.a($$42.a(16));
            double $$122 = $$12.e.a($$42, $$0);
            double $$13 = $$6.b($$42.a(16));
            double $$14 = $$12.b.a($$42);
            double $$15 = $$12.c.a($$42);
            double $$16 = $$12.j.a($$42);
            ebe.a $$17 = ($$1, $$2, $$3, $$4, $$5) -> ebb.a($$2, $$3, $$4, $$16);
            int $$18 = 1;
            if ($$42.a(4) == 0) {
                double $$19 = $$12.f.a($$42);
                float $$20 = 1.0f + $$42.i() * 6.0f;
                this.a($$0, $$12, $$22, $$32, $$52, $$11, $$122, $$13, $$20, $$19, $$7, $$17);
                $$18 += $$42.a(4);
            }
            for (int $$21 = 0; $$21 < $$18; ++$$21) {
                float $$222 = $$42.i() * ((float)Math.PI * 2);
                float $$23 = ($$42.i() - 0.5f) / 4.0f;
                float $$24 = this.a($$42);
                int $$25 = $$8 - $$42.a($$8 / 4);
                boolean $$26 = false;
                this.a($$0, $$12, $$22, $$32, $$42.g(), $$52, $$11, $$122, $$13, $$14, $$15, $$24, $$222, $$23, 0, $$25, this.b(), $$7, $$17);
            }
        }
        return true;
    }

    protected int a() {
        return 15;
    }

    protected float a(ayw $$0) {
        float $$1 = $$0.i() * 2.0f + $$0.i();
        if ($$0.a(10) == 0) {
            $$1 *= $$0.i() * $$0.i() * 3.0f + 1.0f;
        }
        return $$1;
    }

    protected double b() {
        return 1.0;
    }

    protected void a(eaz $$0, eba $$1, duy $$2, Function<jd, jm<ddw>> $$3, dyk $$4, double $$5, double $$6, double $$7, float $$8, double $$9, dux $$10, ebe.a $$11) {
        double $$12 = 1.5 + (double)(ayo.a(1.5707964f) * $$8);
        double $$13 = $$12 * $$9;
        this.a($$0, $$1, $$2, $$3, $$4, $$5 + 1.0, $$6, $$7, $$12, $$13, $$10, $$11);
    }

    protected void a(eaz $$0, eba $$1, duy $$2, Function<jd, jm<ddw>> $$3, long $$4, dyk $$5, double $$6, double $$7, double $$8, double $$9, double $$10, float $$11, float $$12, float $$13, int $$14, int $$15, double $$16, dux $$17, ebe.a $$18) {
        ayw $$19 = ayw.a($$4);
        int $$20 = $$19.a($$15 / 2) + $$15 / 4;
        boolean $$21 = $$19.a(6) == 0;
        float $$22 = 0.0f;
        float $$23 = 0.0f;
        for (int $$24 = $$14; $$24 < $$15; ++$$24) {
            double $$25 = 1.5 + (double)(ayo.a((float)Math.PI * (float)$$24 / (float)$$15) * $$11);
            double $$26 = $$25 * $$16;
            float $$27 = ayo.b($$13);
            $$6 += (double)(ayo.b($$12) * $$27);
            $$7 += (double)ayo.a($$13);
            $$8 += (double)(ayo.a($$12) * $$27);
            $$13 *= $$21 ? 0.92f : 0.7f;
            $$13 += $$23 * 0.1f;
            $$12 += $$22 * 0.1f;
            $$23 *= 0.9f;
            $$22 *= 0.75f;
            $$23 += ($$19.i() - $$19.i()) * $$19.i() * 2.0f;
            $$22 += ($$19.i() - $$19.i()) * $$19.i() * 4.0f;
            if ($$24 == $$20 && $$11 > 1.0f) {
                this.a($$0, $$1, $$2, $$3, $$19.g(), $$5, $$6, $$7, $$8, $$9, $$10, $$19.i() * 0.5f + 0.5f, $$12 - 1.5707964f, $$13 / 3.0f, $$24, $$15, 1.0, $$17, $$18);
                this.a($$0, $$1, $$2, $$3, $$19.g(), $$5, $$6, $$7, $$8, $$9, $$10, $$19.i() * 0.5f + 0.5f, $$12 + 1.5707964f, $$13 / 3.0f, $$24, $$15, 1.0, $$17, $$18);
                return;
            }
            if ($$19.a(4) == 0) continue;
            if (!ebb.a($$2.f(), $$6, $$8, $$24, $$15, $$11)) {
                return;
            }
            this.a($$0, $$1, $$2, $$3, $$5, $$6, $$7, $$8, $$25 * $$9, $$26 * $$10, $$17, $$18);
        }
    }

    private static boolean a(double $$0, double $$1, double $$2, double $$3) {
        if ($$1 <= $$3) {
            return true;
        }
        return $$0 * $$0 + $$1 * $$1 + $$2 * $$2 >= 1.0;
    }
}
