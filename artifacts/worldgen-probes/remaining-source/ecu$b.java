/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  javax.annotation.Nullable
 */
import javax.annotation.Nullable;

static final class ecu.b {
    private final int a;
    @Nullable
    private final exc b;

    ecu.b(int $$0, ayw $$1, bpu $$2) {
        this.a = $$0;
        float $$3 = $$2.a($$1);
        float $$4 = ayo.b($$1, 0.0f, (float)Math.PI);
        this.b = new exc(ayo.b($$4) * $$3, 0.0, ayo.a($$4) * $$3);
    }

    private ecu.b() {
        this.a = 0;
        this.b = null;
    }

    static ecu.b a() {
        return new ecu.b();
    }

    jd a(jd $$0) {
        if (this.b == null) {
            return $$0;
        }
        int $$1 = this.a - $$0.v();
        exc $$2 = this.b.a((double)$$1);
        return $$0.b(ayo.a($$2.c), 0, ayo.a($$2.e));
    }
}
