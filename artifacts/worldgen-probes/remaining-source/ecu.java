/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.serialization.Codec
 *  javax.annotation.Nullable
 */
import com.mojang.serialization.Codec;
import java.util.Optional;
import javax.annotation.Nullable;

public class ecu
extends ece<eel> {
    public ecu(Codec<eel> $$0) {
        super($$0);
    }

    @Override
    public boolean a(ecg<eel> $$0) {
        b $$13;
        dds $$1 = $$0.b();
        jd $$2 = $$0.e();
        eel $$3 = $$0.f();
        ayw $$4 = $$0.d();
        if (!ebz.a($$1, $$2)) {
            return false;
        }
        Optional<dyo> $$5 = dyo.a($$1, $$2, $$3.b, ebz::c, ebz::a);
        if ($$5.isEmpty() || !($$5.get() instanceof dyo.b)) {
            return false;
        }
        dyo.b $$6 = (dyo.b)$$5.get();
        if ($$6.g() < 4) {
            return false;
        }
        int $$7 = (int)((float)$$6.g() * $$3.e);
        int $$8 = ayo.a($$7, $$3.c.a(), $$3.c.b());
        int $$9 = ayo.b($$4, $$3.c.a(), $$8);
        a $$10 = ecu.a($$2.h($$6.e() - 1), false, $$4, $$9, $$3.f, $$3.d);
        a $$11 = ecu.a($$2.h($$6.f() + 1), true, $$4, $$9, $$3.g, $$3.d);
        if ($$10.a($$3) && $$11.a($$3)) {
            b $$12 = new b($$2.v(), $$4, $$3.h);
        } else {
            $$13 = b.a();
        }
        boolean $$14 = $$10.a($$1, $$13);
        boolean $$15 = $$11.a($$1, $$13);
        if ($$14) {
            $$10.a($$1, $$4, $$13);
        }
        if ($$15) {
            $$11.a($$1, $$4, $$13);
        }
        return true;
    }

    private static a a(jd $$0, boolean $$1, ayw $$2, int $$3, bpu $$4, bpu $$5) {
        return new a($$0, $$1, $$3, $$4.a($$2), $$5.a($$2));
    }

    private void a(dds $$0, jd $$1, dyo.b $$2, b $$3) {
        $$0.a($$3.a($$1.h($$2.e() - 1)), dga.cz.o(), 2);
        $$0.a($$3.a($$1.h($$2.f() + 1)), dga.ch.o(), 2);
        jd.a $$4 = $$1.h($$2.f() + 2).k();
        while ($$4.v() < $$2.e() - 1) {
            jd $$5 = $$3.a($$4);
            if (ebz.a($$0, $$5) || $$0.a_($$5).a(dga.su)) {
                $$0.a($$5, dga.gM.o(), 2);
            }
            $$4.c(ji.b);
        }
    }

    static final class a {
        private jd a;
        private final boolean b;
        private int c;
        private final double d;
        private final double e;

        a(jd $$0, boolean $$1, int $$2, double $$3, double $$4) {
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

        boolean a(dds $$0, b $$1) {
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

        void a(dds $$0, ayw $$1, b $$2) {
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

    static final class b {
        private final int a;
        @Nullable
        private final exc b;

        b(int $$0, ayw $$1, bpu $$2) {
            this.a = $$0;
            float $$3 = $$2.a($$1);
            float $$4 = ayo.b($$1, 0.0f, (float)Math.PI);
            this.b = new exc(ayo.b($$4) * $$3, 0.0, ayo.a($$4) * $$3);
        }

        private b() {
            this.a = 0;
            this.b = null;
        }

        static b a() {
            return new b();
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
}
