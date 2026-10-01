/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.serialization.Codec
 */
import com.mojang.serialization.Codec;
import java.util.function.Predicate;

public class edg
extends ece<eey> {
    public edg(Codec<eey> $$0) {
        super($$0);
    }

    @Override
    public boolean a(ecg<eey> $$0) {
        jd $$2;
        dds $$1 = $$0.b();
        if (!$$1.a_($$2 = $$0.e()).i()) {
            return false;
        }
        ayw $$3 = $$0.d();
        jd $$4 = $$0.e();
        eey $$5 = $$0.f();
        jd.a $$6 = $$4.k();
        if (edg.a($$1, $$0.c(), $$5, $$3, $$6, $$4)) {
            edg.a($$1, $$5, $$3, $$4, $$6);
        }
        return true;
    }

    private static boolean a(dds $$0, eey $$1, jd $$2) {
        jd.a $$3 = $$2.k();
        for (int $$4 = 1; $$4 <= $$1.c; ++$$4) {
            $$3.c(ji.b);
            dtc $$5 = $$0.a_($$3);
            if (edg.a($$5, $$4, $$1.n)) continue;
            return false;
        }
        return true;
    }

    private static boolean a(dtc $$0, int $$1, int $$2) {
        if ($$0.i()) {
            return true;
        }
        int $$3 = $$1 + 1;
        return $$3 <= $$2 && $$0.u().a(awk.a);
    }

    private static boolean a(dds $$0, duz $$1, eey $$2, ayw $$3, jd.a $$4, jd $$5) {
        for (int $$6 = 0; $$6 < $$2.h; ++$$6) {
            $$4.c(ji.b);
            if (!$$2.o.test($$0, $$4) || !edg.a($$0, $$2, $$4)) continue;
            jd $$7 = $$4.e();
            if ($$0.b_($$7).a(awk.b) || !$$0.a_($$7).e()) {
                return false;
            }
            if (!$$2.b.a().a($$0, $$1, $$3, $$4)) continue;
            edg.a($$5, $$5.v() + $$6, $$0, $$2, $$3);
            return true;
        }
        return false;
    }

    private static void a(jd $$0, int $$1, dds $$2, eey $$3, ayw $$4) {
        int $$5 = $$0.u();
        int $$6 = $$0.w();
        jd.a $$7 = $$0.k();
        for (int $$8 = $$0.v(); $$8 < $$1; ++$$8) {
            edg.a($$2, $$3, $$4, $$5, $$6, $$7.d($$5, $$8, $$6));
        }
    }

    private static void a(dds $$0, eey $$12, ayw $$2, int $$3, int $$4, jd.a $$5) {
        int $$6 = $$12.d;
        Predicate<dtc> $$7 = $$1 -> $$1.a($$0.e);
        for (int $$8 = 0; $$8 < $$12.g; ++$$8) {
            $$5.a($$5, $$2.a($$6) - $$2.a($$6), 0, $$2.a($$6) - $$2.a($$6));
            if ($$7.test($$0.a_($$5))) {
                $$0.a((jd)$$5, $$12.f.a($$2, $$5), 2);
            }
            $$5.p($$3);
            $$5.r($$4);
        }
    }

    private static void a(dds $$0, eey $$1, ayw $$2, jd $$3, jd.a $$4) {
        int $$5 = $$1.i;
        int $$6 = $$1.j;
        for (int $$7 = 0; $$7 < $$1.l; ++$$7) {
            dtc $$8;
            $$4.a($$3, $$2.a($$5) - $$2.a($$5), $$2.a($$6) - $$2.a($$6), $$2.a($$5) - $$2.a($$5));
            if (!$$0.u($$4) || !($$8 = $$1.k.a($$2, $$4)).a($$0, (jd)$$4) || !$$0.a_($$4.d()).d($$0, $$4, ji.a)) continue;
            $$0.a((jd)$$4, $$8, 2);
        }
    }
}
