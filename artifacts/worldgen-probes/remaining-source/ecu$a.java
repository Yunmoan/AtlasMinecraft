/*
 * Decompiled with CFR 0.152.
 */
static final class ecu.a {
    private jd a;
    private final boolean b;
    private int c;
    private final double d;
    private final double e;

    ecu.a(jd $$0, boolean $$1, int $$2, double $$3, double $$4) {
        this.a = $$0;
        this.b = $$1;
        this.c = $$2;
        this.d = $$3;
        this.e = $$4;
    }

    private int a() {
        return this.a(0.0f);
    }

    private int b() {
        if (this.b) {
            return this.a.v();
        }
        return this.a.v() - this.a();
    }

    private int c() {
        if (!this.b) {
            return this.a.v();
        }
        return this.a.v() + this.a();
    }

    boolean a(dds $$0, ecu.b $$1) {
        while (this.c > 1) {
            jd.a $$2 = this.a.k();
            int $$3 = Math.min(10, this.a());
            for (int $$4 = 0; $$4 < $$3; ++$$4) {
                if ($$0.a_($$2).a(dga.H)) {
                    return false;
                }
                if (ebz.a($$0, $$1.a($$2), this.c)) {
                    this.a = $$2;
                    return true;
                }
                $$2.c(this.b ? ji.a : ji.b);
            }
            this.c /= 2;
        }
        return false;
    }

    private int a(float $$0) {
        return (int)ebz.a($$0, (double)this.c, this.e, this.d);
    }

    void a(dds $$0, ayw $$1, ecu.b $$2) {
        for (int $$3 = -this.c; $$3 <= this.c; ++$$3) {
            block1: for (int $$4 = -this.c; $$4 <= this.c; ++$$4) {
                int $$6;
                float $$5 = ayo.c((float)($$3 * $$3 + $$4 * $$4));
                if ($$5 > (float)this.c || ($$6 = this.a($$5)) <= 0) continue;
                if ((double)$$1.i() < 0.2) {
                    $$6 = (int)((float)$$6 * ayo.b($$1, 0.8f, 1.0f));
                }
                jd.a $$7 = this.a.b($$3, 0, $$4).k();
                boolean $$8 = false;
                int $$9 = this.b ? $$0.a(dyy.a.a, $$7.u(), $$7.w()) : Integer.MAX_VALUE;
                for (int $$10 = 0; $$10 < $$6 && $$7.v() < $$9; ++$$10) {
                    jd $$11 = $$2.a($$7);
                    if (ebz.b($$0, $$11)) {
                        $$8 = true;
                        dfy $$12 = dga.su;
                        $$0.a($$11, $$12.o(), 2);
                    } else if ($$8 && $$0.a_($$11).a(awe.bf)) continue block1;
                    $$7.c(this.b ? ji.b : ji.a);
                }
            }
        }
    }

    boolean a(eel $$0) {
        return this.c >= $$0.i && this.d >= (double)$$0.j;
    }
}
