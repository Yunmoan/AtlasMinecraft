/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.serialization.MapCodec
 */
import com.mojang.serialization.MapCodec;

public class doc
extends dgh
implements dgb {
    public static final MapCodec<doc> a = doc.b(doc::new);
    private static final float d = 0.003f;
    public static final int b = 3;
    public static final duc c = dts.as;
    private static final exv e = dfy.a(3.0, 0.0, 3.0, 13.0, 8.0, 13.0);
    private static final exv f = dfy.a(1.0, 0.0, 1.0, 15.0, 16.0, 15.0);

    public MapCodec<doc> a() {
        return a;
    }

    public doc(dtb.d $$0) {
        super($$0);
        this.k((dtc)((dtc)this.E.b()).a(c, 0));
    }

    @Override
    public cuq a(dcz $$0, jd $$1, dtc $$2) {
        return new cuq(cut.wu);
    }

    @Override
    protected exv a(dtc $$0, dcc $$1, jd $$2, exh $$3) {
        if ($$0.c(c) == 0) {
            return e;
        }
        if ($$0.c(c) < 3) {
            return f;
        }
        return super.a($$0, $$1, $$2, $$3);
    }

    @Override
    protected boolean d_(dtc $$0) {
        return $$0.c(c) < 3;
    }

    @Override
    protected void b(dtc $$0, aqu $$1, jd $$2, ayw $$3) {
        int $$4 = $$0.c(c);
        if ($$4 < 3 && $$3.a(5) == 0 && $$1.b($$2.d(), 0) >= 9) {
            dtc $$5 = (dtc)$$0.a(c, $$4 + 1);
            $$1.a($$2, $$5, 2);
            $$1.a(dxz.c, $$2, dxz.a.a($$5));
        }
    }

    @Override
    protected void a(dtc $$0, dcw $$1, jd $$2, bsr $$3) {
        if (!($$3 instanceof btn) || $$3.am() == bsx.Q || $$3.am() == bsx.h) {
            return;
        }
        $$3.a($$0, new exc(0.8f, 0.75, 0.8f));
        if (!($$1.B || $$0.c(c) <= 0 || $$3.ad == $$3.dt() && $$3.af == $$3.dz())) {
            double $$4 = Math.abs($$3.dt() - $$3.ad);
            double $$5 = Math.abs($$3.dz() - $$3.af);
            if ($$4 >= (double)0.003f || $$5 >= (double)0.003f) {
                $$3.a($$1.aj().t(), 1.0f);
            }
        }
    }

    @Override
    protected bqt a(cuq $$0, dtc $$1, dcw $$2, jd $$3, cmx $$4, bqq $$5, ewy $$6) {
        boolean $$8;
        int $$7 = $$1.c(c);
        boolean bl2 = $$8 = $$7 == 3;
        if (!$$8 && $$0.a(cut.ry)) {
            return bqt.e;
        }
        return super.a($$0, $$1, $$2, $$3, $$4, $$5, $$6);
    }

    @Override
    protected bqr a(dtc $$0, dcw $$1, jd $$2, cmx $$3, ewy $$4) {
        boolean $$6;
        int $$5 = $$0.c(c);
        boolean bl2 = $$6 = $$5 == 3;
        if ($$5 > 1) {
            int $$7 = 1 + $$1.z.a(2);
            doc.a($$1, $$2, new cuq(cut.wu, $$7 + ($$6 ? 1 : 0)));
            $$1.a(null, $$2, avp.zs, avq.e, 1.0f, 0.8f + $$1.z.i() * 0.4f);
            dtc $$8 = (dtc)$$0.a(c, 1);
            $$1.a($$2, $$8, 2);
            $$1.a(dxz.c, $$2, dxz.a.a($$3, $$8));
            return bqr.a($$1.B);
        }
        return super.a($$0, $$1, $$2, $$3, $$4);
    }

    @Override
    protected void a(dtd.a<dfy, dtc> $$0) {
        $$0.a(new duf[]{c});
    }

    @Override
    public boolean b(dcz $$0, jd $$1, dtc $$2) {
        return $$2.c(c) < 3;
    }

    @Override
    public boolean a(dcw $$0, ayw $$1, jd $$2, dtc $$3) {
        return true;
    }

    @Override
    public void a(aqu $$0, ayw $$1, jd $$2, dtc $$3) {
        int $$4 = Math.min(3, $$3.c(c) + 1);
        $$0.a($$2, (dtc)$$3.a(c, $$4), 2);
    }
}

