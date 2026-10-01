/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.serialization.MapCodec
 *  javax.annotation.Nullable
 */
import com.mojang.serialization.MapCodec;
import javax.annotation.Nullable;

public class dne
extends dig
implements dgb,
dna {
    public static final MapCodec<dne> c = dne.b(dne::new);
    private static final dtt g = dts.C;
    public static final dtw d = dts.R;
    protected static final float e = 6.0f;
    protected static final exv f = dfy.a(2.0, 0.0, 2.0, 14.0, 13.0, 14.0);

    public MapCodec<dne> a() {
        return c;
    }

    public dne(dtb.d $$0) {
        super($$0);
        this.k((dtc)((dtc)((dtc)((dtc)this.E.b()).a(b, dty.b)).a(g, false)).a(d, ji.c));
    }

    @Override
    protected exv a(dtc $$0, dcc $$1, jd $$2, exh $$3) {
        return f;
    }

    @Override
    protected boolean b(dtc $$0, dcc $$1, jd $$2) {
        return $$0.a(awe.by) || $$1.b_($$2.d()).a(epf.c) && super.b($$0, $$1, $$2);
    }

    @Override
    @Nullable
    public dtc a(cyd $$0) {
        dtc $$1 = super.a($$0);
        if ($$1 != null) {
            return dne.c($$0.q(), $$0.a(), (dtc)$$1.a(d, $$0.g().g()));
        }
        return null;
    }

    @Override
    public void a(dcw $$0, jd $$1, dtc $$2, btn $$3, cuq $$4) {
        if (!$$0.x_()) {
            jd $$5 = $$1.d();
            dtc $$6 = dig.c($$0, $$5, (dtc)((dtc)this.o().a(b, dty.a)).a(d, $$2.c(d)));
            $$0.a($$5, $$6, 3);
        }
    }

    @Override
    protected epe b_(dtc $$0) {
        if ($$0.c(g).booleanValue()) {
            return epf.c.a(false);
        }
        return super.b_($$0);
    }

    @Override
    protected boolean a(dtc $$0, dcz $$1, jd $$2) {
        if ($$0.c(b) == dty.a) {
            return super.a($$0, $$1, $$2);
        }
        jd $$3 = $$2.e();
        dtc $$4 = $$1.a_($$3);
        return this.b($$4, $$1, $$3);
    }

    @Override
    protected dtc a(dtc $$0, ji $$1, dtc $$2, dcx $$3, jd $$4, jd $$5) {
        if ($$0.c(g).booleanValue()) {
            $$3.a($$4, epf.c, epf.c.a($$3));
        }
        return super.a($$0, $$1, $$2, $$3, $$4, $$5);
    }

    @Override
    protected void a(dtd.a<dfy, dtc> $$0) {
        $$0.a(b, g, d);
    }

    @Override
    public boolean b(dcz $$0, jd $$1, dtc $$2) {
        return true;
    }

    @Override
    public boolean a(dcw $$0, ayw $$1, jd $$2, dtc $$3) {
        return true;
    }

    @Override
    public void a(aqu $$0, ayw $$1, jd $$2, dtc $$3) {
        if ($$3.c(dig.b) == dty.b) {
            jd $$4 = $$2.d();
            $$0.a($$4, $$0.b_($$4).g(), 18);
            dfv.a((dcx)$$0, $$1, $$2, $$3.c(d));
        } else {
            jd $$5 = $$2.e();
            this.a($$0, $$1, $$5, $$0.a_($$5));
        }
    }

    @Override
    protected dtc a(dtc $$0, dmm $$1) {
        return (dtc)$$0.a(d, $$1.a($$0.c(d)));
    }

    @Override
    protected dtc a(dtc $$0, dkv $$1) {
        return $$0.a($$1.a($$0.c(d)));
    }

    @Override
    protected float ax_() {
        return 0.1f;
    }
}
