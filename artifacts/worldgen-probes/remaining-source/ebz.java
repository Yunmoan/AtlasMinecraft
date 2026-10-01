/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  awe
 *  ayo
 *  dcx
 *  dds
 *  dga
 *  dlq
 *  dtc
 *  dtz
 *  duf
 *  jd
 *  jd$a
 *  ji
 */
import java.util.function.Consumer;

public class ebz {
    protected static double a(double $$0, double $$1, double $$2, double $$3) {
        if ($$0 < $$3) {
            $$0 = $$3;
        }
        double $$4 = 0.384;
        double $$5 = $$0 / $$1 * 0.384;
        double $$6 = 0.75 * Math.pow($$5, 1.3333333333333333);
        double $$7 = Math.pow($$5, 0.6666666666666666);
        double $$8 = 0.3333333333333333 * Math.log($$5);
        double $$9 = $$2 * ($$6 - $$7 - $$8);
        $$9 = Math.max($$9, 0.0);
        return $$9 / 0.384 * $$1;
    }

    protected static boolean a(dds $$0, jd $$1, int $$2) {
        if (ebz.b((dcx)$$0, $$1)) {
            return false;
        }
        float $$3 = 6.0f;
        float $$4 = 6.0f / (float)$$2;
        for (float $$5 = 0.0f; $$5 < (float)Math.PI * 2; $$5 += $$4) {
            int $$7;
            int $$6 = (int)(ayo.b((float)$$5) * (float)$$2);
            if (!ebz.b((dcx)$$0, $$1.b($$6, 0, $$7 = (int)(ayo.a((float)$$5) * (float)$$2)))) continue;
            return false;
        }
        return true;
    }

    protected static boolean a(dcx $$0, jd $$1) {
        return $$0.a($$1, ebz::c);
    }

    protected static boolean b(dcx $$0, jd $$1) {
        return $$0.a($$1, ebz::e);
    }

    protected static void a(ji $$0, int $$1, boolean $$2, Consumer<dtc> $$3) {
        if ($$1 >= 3) {
            $$3.accept(ebz.a($$0, dtz.e));
            for (int $$4 = 0; $$4 < $$1 - 3; ++$$4) {
                $$3.accept(ebz.a($$0, dtz.d));
            }
        }
        if ($$1 >= 2) {
            $$3.accept(ebz.a($$0, dtz.c));
        }
        if ($$1 >= 1) {
            $$3.accept(ebz.a($$0, $$2 ? dtz.a : dtz.b));
        }
    }

    protected static void a(dcx $$0, jd $$1, ji $$2, int $$32, boolean $$4) {
        if (!ebz.b($$0.a_($$1.a($$2.g())))) {
            return;
        }
        jd.a $$5 = $$1.k();
        ebz.a($$2, $$32, $$4, (dtc $$3) -> {
            if ($$3.a(dga.st)) {
                $$3 = (dtc)$$3.a((duf)dlq.d, (Comparable)Boolean.valueOf($$0.z((jd)$$5)));
            }
            $$0.a((jd)$$5, $$3, 2);
            $$5.c($$2);
        });
    }

    protected static boolean c(dcx $$0, jd $$1) {
        dtc $$2 = $$0.a_($$1);
        if ($$2.a(awe.bt)) {
            $$0.a($$1, dga.su.o(), 2);
            return true;
        }
        return false;
    }

    private static dtc a(ji $$0, dtz $$1) {
        return (dtc)((dtc)dga.st.o().a((duf)dlq.b, (Comparable)$$0)).a((duf)dlq.c, (Comparable)$$1);
    }

    public static boolean a(dtc $$0) {
        return ebz.b($$0) || $$0.a(dga.H);
    }

    public static boolean b(dtc $$0) {
        return $$0.a(dga.su) || $$0.a(awe.bt);
    }

    public static boolean c(dtc $$0) {
        return $$0.i() || $$0.a(dga.G);
    }

    public static boolean d(dtc $$0) {
        return !$$0.i() && !$$0.a(dga.G);
    }

    public static boolean e(dtc $$0) {
        return $$0.i() || $$0.a(dga.G) || $$0.a(dga.H);
    }
}
