/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.serialization.MapCodec
 */
import com.mojang.serialization.MapCodec;

public class dkx
extends dfy {
    public static final MapCodec<dkx> a = dkx.b(dkx::new);
    protected static final exv b = dfy.a(0.0, 0.0, 0.0, 16.0, 14.0, 16.0);

    public MapCodec<dkx> a() {
        return a;
    }

    public dkx(dtb.d $$0) {
        super($$0);
    }

    @Override
    protected exv b(dtc $$0, dcc $$1, jd $$2, exh $$3) {
        return b;
    }

    @Override
    protected exv b_(dtc $$0, dcc $$1, jd $$2) {
        return exs.b();
    }

    @Override
    protected exv c(dtc $$0, dcc $$1, jd $$2, exh $$3) {
        return exs.b();
    }

    @Override
    protected boolean a(dtc $$0, ept $$1) {
        return false;
    }

    @Override
    protected float d(dtc $$0, dcc $$1, jd $$2) {
        return 0.2f;
    }
}
