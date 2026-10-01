/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.serialization.MapCodec
 */
import com.mojang.serialization.MapCodec;

public class dny
extends dfy {
    public static final MapCodec<dny> a = dny.b(dny::new);
    private static final double b = 5.0;
    private static final exv c = dfy.a(5.0, 5.0, 5.0, 11.0, 11.0, 11.0);

    public MapCodec<dny> a() {
        return a;
    }

    protected dny(dtb.d $$0) {
        super($$0);
    }

    @Override
    protected dmf a_(dtc $$0) {
        return dmf.a;
    }

    @Override
    protected exv a(dtc $$0, dcc $$1, jd $$2, exh $$3) {
        return c;
    }

    @Override
    protected float d(dtc $$0, dcc $$1, jd $$2) {
        return 1.0f;
    }
}
