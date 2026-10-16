/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.serialization.MapCodec
 *  javax.annotation.Nullable
 */
import com.mojang.serialization.MapCodec;
import javax.annotation.Nullable;

public class dmx
extends dgh
implements dgb,
dkp {
    public static final MapCodec<dmx> a = dmx.b(dmx::new);
    protected static final float b = 6.0f;
    protected static final exv c = dfy.a(2.0, 0.0, 2.0, 14.0, 12.0, 14.0);

    public MapCodec<dmx> a() {
        return a;
    }

    protected dmx(dtb.d $$0) {
        super($$0);
    }

    @Override
    protected exv a(dtc $$0, dcc $$1, jd $$2, exh $$3) {
        return c;
    }

    @Override
    protected boolean b(dtc $$0, dcc $$1, jd $$2) {
        return $$0.d($$1, $$2, ji.b) && !$$0.a(dga.kJ);
    }

    @Override
    @Nullable
    public dtc a(cyd $$0) {
        epe $$1 = $$0.q().b_($$0.a());
        if ($$1.a(awk.a) && $$1.e() == 8) {
            return super.a($$0);
        }
        return null;
    }

    @Override
    protected dtc a(dtc $$0, ji $$1, dtc $$2, dcx $$3, jd $$4, jd $$5) {
        dtc $$6 = super.a($$0, $$1, $$2, $$3, $$4, $$5);
        if (!$$6.i()) {
            $$3.a($$4, epf.c, epf.c.a($$3));
        }
        return $$6;
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
    protected epe b_(dtc $$0) {
        return epf.c.a(false);
    }

    @Override
    public void a(aqu $$0, ayw $$1, jd $$2, dtc $$3) {
        dtc $$4 = dga.bx.o();
        dtc $$5 = (dtc)$$4.a(dof.d, dty.a);
        jd $$6 = $$2.d();
        if ($$0.a_($$6).a(dga.G)) {
            $$0.a($$2, $$4, 2);
            $$0.a($$6, $$5, 2);
        }
    }

    @Override
    public boolean a(@Nullable cmx $$0, dcc $$1, jd $$2, dtc $$3, epd $$4) {
        return false;
    }

    @Override
    public boolean a(dcx $$0, jd $$1, dtc $$2, epe $$3) {
        return false;
    }
}
