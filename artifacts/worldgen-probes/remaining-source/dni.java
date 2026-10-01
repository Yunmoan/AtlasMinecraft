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
 *  ddf
 *  dfy
 *  dga
 *  dni$1
 *  dtb$d
 *  dtc
 *  dtd$a
 *  dts
 *  duc
 *  duf
 *  ept
 *  exh
 *  exs
 *  exv
 *  javax.annotation.Nullable
 *  jd
 *  ji
 */
import com.mojang.serialization.MapCodec;
import javax.annotation.Nullable;

public class dni
extends dfy {
    public static final MapCodec<dni> a = dni.b(dni::new);
    public static final int b = 8;
    public static final duc c = dts.aF;
    protected static final exv[] d = new exv[]{exs.a(), dfy.a((double)0.0, (double)0.0, (double)0.0, (double)16.0, (double)2.0, (double)16.0), dfy.a((double)0.0, (double)0.0, (double)0.0, (double)16.0, (double)4.0, (double)16.0), dfy.a((double)0.0, (double)0.0, (double)0.0, (double)16.0, (double)6.0, (double)16.0), dfy.a((double)0.0, (double)0.0, (double)0.0, (double)16.0, (double)8.0, (double)16.0), dfy.a((double)0.0, (double)0.0, (double)0.0, (double)16.0, (double)10.0, (double)16.0), dfy.a((double)0.0, (double)0.0, (double)0.0, (double)16.0, (double)12.0, (double)16.0), dfy.a((double)0.0, (double)0.0, (double)0.0, (double)16.0, (double)14.0, (double)16.0), dfy.a((double)0.0, (double)0.0, (double)0.0, (double)16.0, (double)16.0, (double)16.0)};
    public static final int e = 5;

    public MapCodec<dni> a() {
        return a;
    }

    protected dni(dtb.d $$0) {
        super($$0);
        this.k((dtc)((dtc)this.E.b()).a((duf)c, (Comparable)Integer.valueOf(1)));
    }

    protected boolean a(dtc $$0, ept $$1) {
        switch (1.a[$$1.ordinal()]) {
            case 1: {
                return (Integer)$$0.c((duf)c) < 5;
            }
            case 2: {
                return false;
            }
            case 3: {
                return false;
            }
        }
        return false;
    }

    protected exv a(dtc $$0, dcc $$1, jd $$2, exh $$3) {
        return d[(Integer)$$0.c((duf)c)];
    }

    protected exv b(dtc $$0, dcc $$1, jd $$2, exh $$3) {
        return d[(Integer)$$0.c((duf)c) - 1];
    }

    protected exv b_(dtc $$0, dcc $$1, jd $$2) {
        return d[(Integer)$$0.c((duf)c)];
    }

    protected exv c(dtc $$0, dcc $$1, jd $$2, exh $$3) {
        return d[(Integer)$$0.c((duf)c)];
    }

    protected boolean f_(dtc $$0) {
        return true;
    }

    protected float d(dtc $$0, dcc $$1, jd $$2) {
        return (Integer)$$0.c((duf)c) == 8 ? 0.2f : 1.0f;
    }

    protected boolean a(dtc $$0, dcz $$1, jd $$2) {
        dtc $$3 = $$1.a_($$2.e());
        if ($$3.a(awe.cp)) {
            return false;
        }
        if ($$3.a(awe.cq)) {
            return true;
        }
        return dfy.a((exv)$$3.k((dcc)$$1, $$2.e()), (ji)ji.b) || $$3.a((dfy)this) && (Integer)$$3.c((duf)c) == 8;
    }

    protected dtc a(dtc $$0, ji $$1, dtc $$2, dcx $$3, jd $$4, jd $$5) {
        if (!$$0.a((dcz)$$3, $$4)) {
            return dga.a.o();
        }
        return super.a($$0, $$1, $$2, $$3, $$4, $$5);
    }

    protected void b(dtc $$0, aqu $$1, jd $$2, ayw $$3) {
        if ($$1.a(ddf.b, $$2) > 11) {
            dni.c((dtc)$$0, (dcw)$$1, (jd)$$2);
            $$1.a($$2, false);
        }
    }

    protected boolean a(dtc $$0, cyd $$1) {
        int $$2 = (Integer)$$0.c((duf)c);
        if ($$1.n().a(this.r()) && $$2 < 8) {
            if ($$1.c()) {
                return $$1.k() == ji.b;
            }
            return true;
        }
        return $$2 == 1;
    }

    @Nullable
    public dtc a(cyd $$0) {
        dtc $$1 = $$0.q().a_($$0.a());
        if ($$1.a((dfy)this)) {
            int $$2 = (Integer)$$1.c((duf)c);
            return (dtc)$$1.a((duf)c, (Comparable)Integer.valueOf(Math.min(8, $$2 + 1)));
        }
        return super.a($$0);
    }

    protected void a(dtd.a<dfy, dtc> $$0) {
        $$0.a(new duf[]{c});
    }
}
