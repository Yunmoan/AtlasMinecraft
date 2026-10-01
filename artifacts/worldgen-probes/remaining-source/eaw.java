/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  ayo
 *  ayw
 *  com.mojang.serialization.Codec
 *  dcd
 *  ddw
 *  dux
 *  duy
 *  dyk
 *  dzv
 *  eav
 *  eax
 *  eaz
 *  ebe
 *  jd
 *  jm
 */
import com.mojang.serialization.Codec;
import java.util.function.Function;

public class eaw
extends ebe<eav> {
    public eaw(Codec<eav> $$0) {
        super($$0);
    }

    public boolean a(eav $$0, ayw $$1) {
        return $$1.i() <= $$0.l;
    }

    public boolean a(eaz $$0, eav $$1, duy $$2, Function<jd, jm<ddw>> $$3, ayw $$4, dyk $$5, dcd $$6, dux $$7) {
        int $$8 = (this.d() * 2 - 1) * 16;
        double $$9 = $$6.a($$4.a(16));
        int $$10 = $$1.e.a($$4, (dzv)$$0);
        double $$11 = $$6.b($$4.a(16));
        float $$12 = $$4.i() * ((float)Math.PI * 2);
        float $$13 = $$1.b.a($$4);
        double $$14 = $$1.f.a($$4);
        float $$15 = $$1.c.c.a($$4);
        int $$16 = (int)((float)$$8 * $$1.c.b.a($$4));
        boolean $$17 = false;
        this.a($$0, $$1, $$2, $$3, $$4.g(), $$5, $$9, $$10, $$11, $$15, $$12, $$13, 0, $$16, $$14, $$7);
        return true;
    }

    private void a(eaz $$0, eav $$12, duy $$22, Function<jd, jm<ddw>> $$32, long $$42, dyk $$52, double $$6, double $$7, double $$8, float $$9, float $$10, float $$11, int $$122, int $$13, double $$14, dux $$15) {
        ayw $$16 = ayw.a((long)$$42);
        float[] $$17 = this.a($$0, $$12, $$16);
        float $$18 = 0.0f;
        float $$19 = 0.0f;
        for (int $$20 = $$122; $$20 < $$13; ++$$20) {
            double $$21 = 1.5 + (double)(ayo.a((float)((float)$$20 * (float)Math.PI / (float)$$13)) * $$9);
            double $$222 = $$21 * $$14;
            $$21 *= (double)$$12.c.e.a($$16);
            $$222 = this.a($$12, $$16, $$222, $$13, $$20);
            float $$23 = ayo.b((float)$$11);
            float $$24 = ayo.a((float)$$11);
            $$6 += (double)(ayo.b((float)$$10) * $$23);
            $$7 += (double)$$24;
            $$8 += (double)(ayo.a((float)$$10) * $$23);
            $$11 *= 0.7f;
            $$11 += $$19 * 0.05f;
            $$10 += $$18 * 0.05f;
            $$19 *= 0.8f;
            $$18 *= 0.5f;
            $$19 += ($$16.i() - $$16.i()) * $$16.i() * 2.0f;
            $$18 += ($$16.i() - $$16.i()) * $$16.i() * 4.0f;
            if ($$16.a(4) == 0) continue;
            if (!eaw.a((dcd)$$22.f(), (double)$$6, (double)$$8, (int)$$20, (int)$$13, (float)$$9)) {
                return;
            }
            this.a($$0, (eax)$$12, $$22, $$32, $$52, $$6, $$7, $$8, $$21, $$222, $$15, ($$1, $$2, $$3, $$4, $$5) -> this.a($$1, $$17, $$2, $$3, $$4, $$5));
        }
    }

    private float[] a(eaz $$0, eav $$1, ayw $$2) {
        int $$3 = $$0.b();
        float[] $$4 = new float[$$3];
        float $$5 = 1.0f;
        for (int $$6 = 0; $$6 < $$3; ++$$6) {
            if ($$6 == 0 || $$2.a($$1.c.d) == 0) {
                $$5 = 1.0f + $$2.i() * $$2.i();
            }
            $$4[$$6] = $$5 * $$5;
        }
        return $$4;
    }

    private double a(eav $$0, ayw $$1, double $$2, float $$3, float $$4) {
        float $$5 = 1.0f - ayo.e((float)(0.5f - $$4 / $$3)) * 2.0f;
        float $$6 = $$0.c.f + $$0.c.g * $$5;
        return (double)$$6 * $$2 * (double)ayo.b((ayw)$$1, (float)0.75f, (float)1.0f);
    }

    private boolean a(eaz $$0, float[] $$1, double $$2, double $$3, double $$4, int $$5) {
        int $$6 = $$5 - $$0.a();
        return ($$2 * $$2 + $$4 * $$4) * (double)$$1[$$6 - 1] + $$3 * $$3 / 6.0 >= 1.0;
    }
}
