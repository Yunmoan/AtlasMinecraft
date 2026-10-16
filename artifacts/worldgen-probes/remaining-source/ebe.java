/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.google.common.collect.ImmutableSet
 *  com.mojang.serialization.Codec
 *  com.mojang.serialization.MapCodec
 *  javax.annotation.Nullable
 *  org.apache.commons.lang3.mutable.MutableBoolean
 */
import com.google.common.collect.ImmutableSet;
import com.mojang.serialization.Codec;
import com.mojang.serialization.MapCodec;
import java.util.Set;
import java.util.function.Function;
import javax.annotation.Nullable;
import org.apache.commons.lang3.mutable.MutableBoolean;

public abstract class ebe<C extends eax> {
    public static final ebe<eba> a = ebe.a("cave", new ebb(eba.a));
    public static final ebe<eba> b = ebe.a("nether_cave", new ebd(eba.a));
    public static final ebe<eav> c = ebe.a("canyon", new eaw(eav.a));
    protected static final dtc d = dga.a.o();
    protected static final dtc e = dga.nc.o();
    protected static final epe f = epf.c.g();
    protected static final epe g = epf.e.g();
    protected Set<epd> h = ImmutableSet.of((Object)epf.c);
    private final MapCodec<ebc<C>> i;

    private static <C extends eax, F extends ebe<C>> F a(String $$0, F $$1) {
        return (F)jz.a(lt.N, $$0, $$1);
    }

    public ebe(Codec<C> $$0) {
        this.i = $$0.fieldOf("config").xmap(this::a, ebc::b);
    }

    public ebc<C> a(C $$0) {
        return new ebc<C>(this, $$0);
    }

    public MapCodec<ebc<C>> c() {
        return this.i;
    }

    public int d() {
        return 4;
    }

    protected boolean a(eaz $$0, C $$1, duy $$2, Function<jd, jm<ddw>> $$3, dyk $$4, double $$5, double $$6, double $$7, double $$8, double $$9, dux $$10, a $$11) {
        dcd $$12 = $$2.f();
        double $$13 = $$12.b();
        double $$14 = $$12.c();
        double $$15 = 16.0 + $$8 * 2.0;
        if (Math.abs($$5 - $$13) > $$15 || Math.abs($$7 - $$14) > $$15) {
            return false;
        }
        int $$16 = $$12.d();
        int $$17 = $$12.e();
        int $$18 = Math.max(ayo.a($$5 - $$8) - $$16 - 1, 0);
        int $$19 = Math.min(ayo.a($$5 + $$8) - $$16, 15);
        int $$20 = Math.max(ayo.a($$6 - $$9) - 1, $$0.a() + 1);
        int $$21 = $$2.y() ? 0 : 7;
        int $$22 = Math.min(ayo.a($$6 + $$9) + 1, $$0.a() + $$0.b() - 1 - $$21);
        int $$23 = Math.max(ayo.a($$7 - $$8) - $$17 - 1, 0);
        int $$24 = Math.min(ayo.a($$7 + $$8) - $$17, 15);
        boolean $$25 = false;
        jd.a $$26 = new jd.a();
        jd.a $$27 = new jd.a();
        for (int $$28 = $$18; $$28 <= $$19; ++$$28) {
            int $$29 = $$12.a($$28);
            double $$30 = ((double)$$29 + 0.5 - $$5) / $$8;
            for (int $$31 = $$23; $$31 <= $$24; ++$$31) {
                int $$32 = $$12.b($$31);
                double $$33 = ((double)$$32 + 0.5 - $$7) / $$8;
                if ($$30 * $$30 + $$33 * $$33 >= 1.0) continue;
                MutableBoolean $$34 = new MutableBoolean(false);
                for (int $$35 = $$22; $$35 > $$20; --$$35) {
                    double $$36 = ((double)$$35 - 0.5 - $$6) / $$9;
                    if ($$11.shouldSkip($$0, $$30, $$36, $$33, $$35) || $$10.b($$28, $$35, $$31) && !ebe.b($$1)) continue;
                    $$10.a($$28, $$35, $$31);
                    $$26.d($$29, $$35, $$32);
                    $$25 |= this.a($$0, $$1, $$2, $$3, $$10, $$26, $$27, $$4, $$34);
                }
            }
        }
        return $$25;
    }

    protected boolean a(eaz $$0, C $$1, duy $$22, Function<jd, jm<ddw>> $$3, dux $$4, jd.a $$5, jd.a $$6, dyk $$7, MutableBoolean $$8) {
        dtc $$9 = $$22.a_($$5);
        if ($$9.a(dga.i) || $$9.a(dga.fl)) {
            $$8.setTrue();
        }
        if (!this.a($$1, $$9) && !ebe.b($$1)) {
            return false;
        }
        dtc $$10 = this.a($$0, $$1, $$5, $$7);
        if ($$10 == null) {
            return false;
        }
        $$22.a($$5, $$10, false);
        if ($$7.a() && !$$10.u().c()) {
            $$22.e($$5);
        }
        if ($$8.isTrue()) {
            $$6.a((kh)$$5, ji.a);
            if ($$22.a_($$6).a(dga.j)) {
                $$0.a($$3, $$22, $$6, !$$10.u().c()).ifPresent($$2 -> {
                    $$22.a($$6, (dtc)$$2, false);
                    if (!$$2.u().c()) {
                        $$22.e($$6);
                    }
                });
            }
        }
        return true;
    }

    @Nullable
    private dtc a(eaz $$0, C $$1, jd $$2, dyk $$3) {
        if ($$2.v() <= ((eax)$$1).g.a($$0)) {
            return g.g();
        }
        dtc $$4 = $$3.a(new dyr.e($$2.u(), $$2.v(), $$2.w()), 0.0);
        if ($$4 == null) {
            return ebe.b($$1) ? ((eax)$$1).h.e() : null;
        }
        return ebe.b($$1) ? ebe.b($$1, $$4) : $$4;
    }

    private static dtc b(eax $$0, dtc $$1) {
        if ($$1.a(dga.a)) {
            return $$0.h.b();
        }
        if ($$1.a(dga.G)) {
            dtc $$2 = $$0.h.c();
            if ($$2.b(dts.C)) {
                return (dtc)$$2.a(dts.C, true);
            }
            return $$2;
        }
        if ($$1.a(dga.H)) {
            return $$0.h.d();
        }
        return $$1;
    }

    public abstract boolean a(eaz var1, C var2, duy var3, Function<jd, jm<ddw>> var4, ayw var5, dyk var6, dcd var7, dux var8);

    public abstract boolean a(C var1, ayw var2);

    protected boolean a(C $$0, dtc $$1) {
        return $$1.a(((eax)$$0).i);
    }

    protected static boolean a(dcd $$0, double $$1, double $$2, int $$3, int $$4, float $$5) {
        double $$11;
        double $$10;
        double $$7;
        double $$9;
        double $$6 = $$0.b();
        double $$8 = $$1 - $$6;
        return $$8 * $$8 + ($$9 = $$2 - ($$7 = (double)$$0.c())) * $$9 - ($$10 = (double)($$4 - $$3)) * $$10 <= ($$11 = (double)($$5 + 2.0f + 16.0f)) * $$11;
    }

    private static boolean b(eax $$0) {
        return $$0.h.a();
    }

    public static interface a {
        public boolean shouldSkip(eaz var1, double var2, double var4, double var6, int var8);
    }
}
