/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.serialization.MapCodec
 */
import com.mojang.serialization.MapCodec;
import java.util.OptionalInt;

public class dki
extends dfy
implements dna {
    public static final MapCodec<dki> b = dki.b(dki::new);
    public static final int c = 7;
    public static final duc d = dts.aC;
    public static final dtt e = dts.v;
    public static final dtt f = dts.C;
    private static final int a = 1;

    public MapCodec<? extends dki> a() {
        return b;
    }

    public dki(dtb.d $$0) {
        super($$0);
        this.k((dtc)((dtc)((dtc)((dtc)this.E.b()).a(d, 7)).a(e, false)).a(f, false));
    }

    @Override
    protected exv b_(dtc $$0, dcc $$1, jd $$2) {
        return exs.a();
    }

    @Override
    protected boolean d_(dtc $$0) {
        return $$0.c(d) == 7 && $$0.c(e) == false;
    }

    @Override
    protected void b(dtc $$0, aqu $$1, jd $$2, ayw $$3) {
        if (this.m($$0)) {
            dki.c($$0, $$1, $$2);
            $$1.a($$2, false);
        }
    }

    protected boolean m(dtc $$0) {
        return $$0.c(e) == false && $$0.c(d) == 7;
    }

    @Override
    protected void a(dtc $$0, aqu $$1, jd $$2, ayw $$3) {
        $$1.a($$2, dki.a($$0, $$1, $$2), 3);
    }

    @Override
    protected int g(dtc $$0, dcc $$1, jd $$2) {
        return 1;
    }

    @Override
    protected dtc a(dtc $$0, ji $$1, dtc $$2, dcx $$3, jd $$4, jd $$5) {
        int $$6;
        if ($$0.c(f).booleanValue()) {
            $$3.a($$4, epf.c, epf.c.a($$3));
        }
        if (($$6 = dki.o($$2) + 1) != 1 || $$0.c(d) != $$6) {
            $$3.a($$4, this, 1);
        }
        return $$0;
    }

    private static dtc a(dtc $$0, dcx $$1, jd $$2) {
        int $$3 = 7;
        jd.a $$4 = new jd.a();
        for (ji $$5 : ji.values()) {
            $$4.a((kh)$$2, $$5);
            $$3 = Math.min($$3, dki.o($$1.a_($$4)) + 1);
            if ($$3 == 1) break;
        }
        return (dtc)$$0.a(d, $$3);
    }

    private static int o(dtc $$0) {
        return dki.n($$0).orElse(7);
    }

    public static OptionalInt n(dtc $$0) {
        if ($$0.a(awe.u)) {
            return OptionalInt.of(0);
        }
        if ($$0.b(d)) {
            return OptionalInt.of($$0.c(d));
        }
        return OptionalInt.empty();
    }

    @Override
    protected epe b_(dtc $$0) {
        if ($$0.c(f).booleanValue()) {
            return epf.c.a(false);
        }
        return super.b_($$0);
    }

    @Override
    public void a(dtc $$0, dcw $$1, jd $$2, ayw $$3) {
        if (!$$1.r($$2.d())) {
            return;
        }
        if ($$3.a(15) != 1) {
            return;
        }
        jd $$4 = $$2.e();
        dtc $$5 = $$1.a_($$4);
        if ($$5.p() && $$5.d($$1, $$4, ji.b)) {
            return;
        }
        ays.a($$1, $$2, $$3, (lk)lm.l);
    }

    @Override
    protected void a(dtd.a<dfy, dtc> $$0) {
        $$0.a(d, e, f);
    }

    @Override
    public dtc a(cyd $$0) {
        epe $$1 = $$0.q().b_($$0.a());
        dtc $$2 = (dtc)((dtc)this.o().a(e, true)).a(f, $$1.a() == epf.c);
        return dki.a($$2, $$0.q(), $$0.a());
    }
}
