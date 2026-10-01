/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.serialization.Codec
 */
import com.mojang.serialization.Codec;

public class edu
extends ece<eep> {
    public edu(Codec<eep> $$0) {
        super($$0);
    }

    @Override
    public boolean a(ecg<eep> $$0) {
        dds $$1 = $$0.b();
        jd $$2 = $$0.e();
        $$0.f();
        if (!$$1.u($$2)) {
            return false;
        }
        for (ji $$3 : ji.values()) {
            if ($$3 == ji.a || !dov.a((dcc)$$1, $$2.a($$3), $$3)) continue;
            $$1.a($$2, (dtc)dga.ff.o().a(dov.a($$3), true), 2);
            return true;
        }
        return false;
    }
}
