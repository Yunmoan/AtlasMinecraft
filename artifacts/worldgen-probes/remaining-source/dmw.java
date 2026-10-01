/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  aqu
 *  awe
 *  ayw
 *  com.mojang.serialization.MapCodec
 *  cyd
 *  dcc
 *  dcw
 *  dcx
 *  dcz
 *  dfy
 *  dga
 *  dgb
 *  dgh
 *  dna
 *  dtb$d
 *  dtc
 *  dtd$a
 *  dts
 *  dtt
 *  duc
 *  duf
 *  epd
 *  epe
 *  epf
 *  ept
 *  exh
 *  exv
 *  javax.annotation.Nullable
 *  jd
 *  ji
 */
import com.mojang.serialization.MapCodec;
import javax.annotation.Nullable;

public class dmw
extends dgh
implements dgb,
dna {
    public static final MapCodec<dmw> a = dmw.b(dmw::new);
    public static final int b = 4;
    public static final duc c = dts.aS;
    public static final dtt d = dts.C;
    protected static final exv e = dfy.a((double)6.0, (double)0.0, (double)6.0, (double)10.0, (double)6.0, (double)10.0);
    protected static final exv f = dfy.a((double)3.0, (double)0.0, (double)3.0, (double)13.0, (double)6.0, (double)13.0);
    protected static final exv g = dfy.a((double)2.0, (double)0.0, (double)2.0, (double)14.0, (double)6.0, (double)14.0);
    protected static final exv h = dfy.a((double)2.0, (double)0.0, (double)2.0, (double)14.0, (double)7.0, (double)14.0);

    public MapCodec<dmw> a() {
        return a;
    }

    protected dmw(dtb.d $$0) {
        super($$0);
        this.k((dtc)((dtc)((dtc)this.E.b()).a((duf)c, (Comparable)Integer.valueOf(1))).a((duf)d, (Comparable)Boolean.valueOf(true)));
    }

    @Nullable
    public dtc a(cyd $$0) {
        dtc $$1 = $$0.q().a_($$0.a());
        if ($$1.a((dfy)this)) {
            return (dtc)$$1.a((duf)c, (Comparable)Integer.valueOf(Math.min(4, (Integer)$$1.c((duf)c) + 1)));
        }
        epe $$2 = $$0.q().b_($$0.a());
        boolean $$3 = $$2.a() == epf.c;
        return (dtc)super.a($$0).a((duf)d, (Comparable)Boolean.valueOf($$3));
    }

    public static boolean m(dtc $$0) {
        return (Boolean)$$0.c((duf)d) == false;
    }

    protected boolean b(dtc $$0, dcc $$1, jd $$2) {
        return !$$0.k($$1, $$2).a(ji.b).c() || $$0.d($$1, $$2, ji.b);
    }

    protected boolean a(dtc $$0, dcz $$1, jd $$2) {
        jd $$3 = $$2.e();
        return this.b($$1.a_($$3), (dcc)$$1, $$3);
    }

    protected dtc a(dtc $$0, ji $$1, dtc $$2, dcx $$3, jd $$4, jd $$5) {
        if (!$$0.a((dcz)$$3, $$4)) {
            return dga.a.o();
        }
        if (((Boolean)$$0.c((duf)d)).booleanValue()) {
            $$3.a($$4, (epd)epf.c, epf.c.a((dcz)$$3));
        }
        return super.a($$0, $$1, $$2, $$3, $$4, $$5);
    }

    protected boolean a(dtc $$0, cyd $$1) {
        if (!$$1.h() && $$1.n().a(this.r()) && (Integer)$$0.c((duf)c) < 4) {
            return true;
        }
        return super.a($$0, $$1);
    }

    protected exv a(dtc $$0, dcc $$1, jd $$2, exh $$3) {
        switch ((Integer)$$0.c((duf)c)) {
            default: {
                return e;
            }
            case 2: {
                return f;
            }
            case 3: {
                return g;
            }
            case 4: 
        }
        return h;
    }

    protected epe b_(dtc $$0) {
        if (((Boolean)$$0.c((duf)d)).booleanValue()) {
            return epf.c.a(false);
        }
        return super.b_($$0);
    }

    protected void a(dtd.a<dfy, dtc> $$0) {
        $$0.a(new duf[]{c, d});
    }

    public boolean b(dcz $$0, jd $$1, dtc $$2) {
        return true;
    }

    public boolean a(dcw $$0, ayw $$1, jd $$2, dtc $$3) {
        return true;
    }

    public void a(aqu $$0, ayw $$1, jd $$2, dtc $$3) {
        if (!dmw.m($$3) && $$0.a_($$2.e()).a(awe.ar)) {
            int $$4 = 5;
            int $$5 = 1;
            int $$6 = 2;
            int $$7 = 0;
            int $$8 = $$2.u() - 2;
            int $$9 = 0;
            for (int $$10 = 0; $$10 < 5; ++$$10) {
                for (int $$11 = 0; $$11 < $$5; ++$$11) {
                    int $$12 = 2 + $$2.v() - 1;
                    for (int $$13 = $$12 - 2; $$13 < $$12; ++$$13) {
                        dtc $$15;
                        jd $$14 = new jd($$8 + $$10, $$13, $$2.w() - $$9 + $$11);
                        if ($$14 == $$2 || $$1.a(6) != 0 || !$$0.a_($$14).a(dga.G) || !($$15 = $$0.a_($$14.e())).a(awe.ar)) continue;
                        $$0.a($$14, (dtc)dga.mV.o().a((duf)c, (Comparable)Integer.valueOf($$1.a(4) + 1)), 3);
                    }
                }
                if ($$7 < 2) {
                    $$5 += 2;
                    ++$$9;
                } else {
                    $$5 -= 2;
                    --$$9;
                }
                ++$$7;
            }
            $$0.a($$2, (dtc)$$3.a((duf)c, (Comparable)Integer.valueOf(4)), 2);
        }
    }

    protected boolean a(dtc $$0, ept $$1) {
        return false;
    }
}
