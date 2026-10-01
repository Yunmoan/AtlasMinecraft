/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  ayw
 *  com.mojang.serialization.Codec
 *  dcz
 *  dds
 *  dga
 *  dkc
 *  dtc
 *  duf
 *  dyy$a
 *  ece
 *  ecg
 *  eep
 *  jd
 */
import com.mojang.serialization.Codec;

public class ecs
extends ece<eep> {
    public ecs(Codec<eep> $$0) {
        super($$0);
    }

    public boolean a(ecg<eep> $$0) {
        int $$1 = 0;
        dds $$2 = $$0.b();
        jd $$3 = $$0.e();
        ayw $$4 = $$0.d();
        int $$5 = $$2.a(dyy.a.d, $$3.u(), $$3.w());
        jd $$6 = new jd($$3.u(), $$5, $$3.w());
        if ($$2.a_($$6).a(dga.G)) {
            dtc $$7 = dga.mc.o();
            dtc $$8 = dga.md.o();
            int $$9 = 1 + $$4.a(10);
            for (int $$10 = 0; $$10 <= $$9; ++$$10) {
                if ($$2.a_($$6).a(dga.G) && $$2.a_($$6.d()).a(dga.G) && $$8.a((dcz)$$2, $$6)) {
                    if ($$10 == $$9) {
                        $$2.a($$6, (dtc)$$7.a((duf)dkc.e, (Comparable)Integer.valueOf($$4.a(4) + 20)), 2);
                        ++$$1;
                    } else {
                        $$2.a($$6, $$8, 2);
                    }
                } else if ($$10 > 0) {
                    jd $$11 = $$6.e();
                    if (!$$7.a((dcz)$$2, $$11) || $$2.a_($$11.e()).a(dga.mc)) break;
                    $$2.a($$11, (dtc)$$7.a((duf)dkc.e, (Comparable)Integer.valueOf($$4.a(4) + 20)), 2);
                    ++$$1;
                    break;
                }
                $$6 = $$6.d();
            }
        }
        return $$1 > 0;
    }
}
