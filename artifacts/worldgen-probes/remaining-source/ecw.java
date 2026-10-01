/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  ayw
 *  com.mojang.serialization.Codec
 *  dcc
 *  dcx
 *  dds
 *  dfy
 *  dga
 *  dtc
 *  ece
 *  ecg
 *  een
 *  jd
 *  jd$a
 *  ji
 *  kh
 */
import com.mojang.serialization.Codec;
import java.util.List;

public class ecw
extends ece<een> {
    public ecw(Codec<een> $$0) {
        super($$0);
    }

    public boolean a(ecg<een> $$0) {
        dds $$1 = $$0.b();
        jd $$2 = $$0.e();
        ayw $$3 = $$0.d();
        een $$4 = (een)$$0.f();
        if (!ecw.c($$1.a_($$2))) {
            return false;
        }
        List $$5 = $$4.a($$3);
        if (ecw.a($$1, $$2, $$1.a_($$2), $$4, $$3, $$5)) {
            return true;
        }
        jd.a $$6 = $$2.k();
        block0: for (ji $$7 : $$5) {
            $$6.g((kh)$$2);
            List $$8 = $$4.a($$3, $$7.g());
            for (int $$9 = 0; $$9 < $$4.c; ++$$9) {
                $$6.a((kh)$$2, $$7);
                dtc $$10 = $$1.a_((jd)$$6);
                if (!ecw.c($$10) && !$$10.a((dfy)$$4.b)) continue block0;
                if (!ecw.a($$1, (jd)$$6, $$10, $$4, $$3, $$8)) continue;
                return true;
            }
        }
        return false;
    }

    public static boolean a(dds $$0, jd $$1, dtc $$2, een $$3, ayw $$4, List<ji> $$5) {
        jd.a $$6 = $$1.k();
        for (ji $$7 : $$5) {
            dtc $$8 = $$0.a_((jd)$$6.a((kh)$$1, $$7));
            if (!$$8.a($$3.h)) continue;
            dtc $$9 = $$3.b.c($$2, (dcc)$$0, $$1, $$7);
            if ($$9 == null) {
                return false;
            }
            $$0.a($$1, $$9, 3);
            $$0.y($$1).e($$1);
            if ($$4.i() < $$3.g) {
                $$3.b.c().a($$9, (dcx)$$0, $$1, $$7, $$4, true);
            }
            return true;
        }
        return false;
    }

    private static boolean c(dtc $$0) {
        return $$0.i() || $$0.a(dga.G);
    }
}
