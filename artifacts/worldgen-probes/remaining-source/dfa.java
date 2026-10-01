/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.serialization.MapCodec
 */
import com.mojang.serialization.MapCodec;

public class dfa
extends dgh
implements dgb {
    public static final MapCodec<dfa> a = dfa.b(dfa::new);
    private static final exv b = exs.a(dfy.a(0.0, 8.0, 0.0, 16.0, 16.0, 16.0), dfy.a(6.0, 0.0, 6.0, 10.0, 8.0, 10.0));

    public MapCodec<dfa> a() {
        return a;
    }

    protected dfa(dtb.d $$0) {
        super($$0);
    }

    @Override
    protected exv a(dtc $$0, dcc $$1, jd $$2, exh $$3) {
        return b;
    }

    @Override
    protected boolean b(dtc $$0, dcc $$1, jd $$2) {
        return $$0.a(dga.dR) || super.b($$0, $$1, $$2);
    }

    @Override
    public boolean b(dcz $$0, jd $$1, dtc $$2) {
        return $$0.b_($$1.d()).c();
    }

    @Override
    public boolean a(dcw $$0, ayw $$1, jd $$2, dtc $$3) {
        return (double)$$0.z.i() < 0.45;
    }

    @Override
    public void a(aqu $$0, ayw $$1, jd $$2, dtc $$3) {
        dsr.e.a($$0, $$0.l().g(), $$2, $$3, $$1);
    }
}
