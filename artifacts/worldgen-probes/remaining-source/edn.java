/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.serialization.Codec
 *  dcz
 *  dds
 *  ddw
 *  dga
 *  dnj
 *  dtc
 *  duf
 *  dyy$a
 *  ece
 *  ecg
 *  eep
 *  jd
 *  jd$a
 *  ji
 *  kh
 */
import com.mojang.serialization.Codec;

public class edn
extends ece<eep> {
    public edn(Codec<eep> $$0) {
        super($$0);
    }

    public boolean a(ecg<eep> $$0) {
        dds $$1 = $$0.b();
        jd $$2 = $$0.e();
        jd.a $$3 = new jd.a();
        jd.a $$4 = new jd.a();
        for (int $$5 = 0; $$5 < 16; ++$$5) {
            for (int $$6 = 0; $$6 < 16; ++$$6) {
                int $$7 = $$2.u() + $$5;
                int $$8 = $$2.w() + $$6;
                int $$9 = $$1.a(dyy.a.e, $$7, $$8);
                $$3.d($$7, $$9, $$8);
                $$4.g((kh)$$3).c(ji.a, 1);
                ddw $$10 = (ddw)$$1.t((jd)$$3).a();
                if ($$10.a((dcz)$$1, (jd)$$4, false)) {
                    $$1.a((jd)$$4, dga.dO.o(), 2);
                }
                if (!$$10.b((dcz)$$1, (jd)$$3)) continue;
                $$1.a((jd)$$3, dga.dN.o(), 2);
                dtc $$11 = $$1.a_((jd)$$4);
                if (!$$11.b((duf)dnj.c)) continue;
                $$1.a((jd)$$4, (dtc)$$11.a((duf)dnj.c, (Comparable)Boolean.valueOf(true)), 2);
            }
        }
        return true;
    }
}
