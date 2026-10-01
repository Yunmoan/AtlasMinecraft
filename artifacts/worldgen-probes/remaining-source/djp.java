/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.serialization.MapCodec
 *  javax.annotation.Nullable
 */
import com.mojang.serialization.MapCodec;
import javax.annotation.Nullable;

public class djp
extends dfy
implements dna {
    public static final MapCodec<djp> a = djp.b(djp::new);
    private static final dtt c = dts.C;
    protected static final exv b = dfy.a(2.0, 10.0, 2.0, 14.0, 16.0, 14.0);

    public MapCodec<djp> a() {
        return a;
    }

    protected djp(dtb.d $$0) {
        super($$0);
        this.k((dtc)((dtc)this.E.b()).a(c, false));
    }

    @Override
    protected void a(dtd.a<dfy, dtc> $$0) {
        $$0.a(new duf[]{c});
    }

    @Override
    protected epe b_(dtc $$0) {
        if ($$0.c(c).booleanValue()) {
            return epf.c.a(false);
        }
        return super.b_($$0);
    }

    @Override
    @Nullable
    public dtc a(cyd $$0) {
        dtc $$1 = super.a($$0);
        if ($$1 != null) {
            epe $$2 = $$0.q().b_($$0.a());
            return (dtc)$$1.a(c, $$2.a() == epf.c);
        }
        return null;
    }

    @Override
    protected boolean a(dtc $$0, dcz $$1, jd $$2) {
        jd $$3 = $$2.d();
        dtc $$4 = $$1.a_($$3);
        return $$4.d($$1, $$3, ji.a);
    }

    @Override
    protected exv a(dtc $$0, dcc $$1, jd $$2, exh $$3) {
        return b;
    }

    @Override
    protected dtc a(dtc $$0, ji $$1, dtc $$2, dcx $$3, jd $$4, jd $$5) {
        if ($$1 == ji.b && !this.a($$0, $$3, $$4)) {
            return dga.a.o();
        }
        if ($$0.c(c).booleanValue()) {
            $$3.a($$4, epf.c, epf.c.a($$3));
        }
        return super.a($$0, $$1, $$2, $$3, $$4, $$5);
    }
}
