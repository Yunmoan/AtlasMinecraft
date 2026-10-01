/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.serialization.MapCodec
 */
import com.mojang.serialization.MapCodec;

public abstract class dgh
extends dfy {
    protected dgh(dtb.d $$0) {
        super($$0);
    }

    protected abstract MapCodec<? extends dgh> a();

    protected boolean b(dtc $$0, dcc $$1, jd $$2) {
        return $$0.a(awe.ag) || $$0.a(dga.cC);
    }

    @Override
    protected dtc a(dtc $$0, ji $$1, dtc $$2, dcx $$3, jd $$4, jd $$5) {
        if (!$$0.a($$3, $$4)) {
            return dga.a.o();
        }
        return super.a($$0, $$1, $$2, $$3, $$4, $$5);
    }

    @Override
    protected boolean a(dtc $$0, dcz $$1, jd $$2) {
        jd $$3 = $$2.e();
        return this.b($$1.a_($$3), $$1, $$3);
    }

    @Override
    protected boolean a_(dtc $$0, dcc $$1, jd $$2) {
        return $$0.u().c();
    }

    @Override
    protected boolean a(dtc $$0, ept $$1) {
        if ($$1 == ept.c && !this.aG) {
            return true;
        }
        return super.a($$0, $$1);
    }
}

