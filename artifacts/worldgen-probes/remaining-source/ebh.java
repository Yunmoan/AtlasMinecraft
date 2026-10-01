/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  ayw
 *  com.mojang.serialization.Codec
 *  dcz
 *  dds
 *  dfc
 *  dga
 *  dtc
 *  dto
 *  duf
 *  dyy$a
 *  ece
 *  ecg
 *  ees
 *  jd
 *  jd$a
 *  ji
 */
import com.mojang.serialization.Codec;

public class ebh
extends ece<ees> {
    private static final dtc a = (dtc)((dtc)((dtc)dga.mZ.o().a((duf)dfc.h, (Comparable)Integer.valueOf(1))).a((duf)dfc.i, (Comparable)dto.a)).a((duf)dfc.j, (Comparable)Integer.valueOf(0));
    private static final dtc b = (dtc)((dtc)a.a((duf)dfc.i, (Comparable)dto.c)).a((duf)dfc.j, (Comparable)Integer.valueOf(1));
    private static final dtc c = (dtc)a.a((duf)dfc.i, (Comparable)dto.c);
    private static final dtc d = (dtc)a.a((duf)dfc.i, (Comparable)dto.b);

    public ebh(Codec<ees> $$0) {
        super($$0);
    }

    public boolean a(ecg<ees> $$0) {
        int $$1 = 0;
        jd $$2 = $$0.e();
        dds $$3 = $$0.b();
        ayw $$4 = $$0.d();
        ees $$5 = (ees)$$0.f();
        jd.a $$6 = $$2.k();
        jd.a $$7 = $$2.k();
        if ($$3.u((jd)$$6)) {
            if (dga.mZ.o().a((dcz)$$3, (jd)$$6)) {
                int $$8 = $$4.a(12) + 5;
                if ($$4.i() < $$5.l) {
                    int $$9 = $$4.a(4) + 1;
                    for (int $$10 = $$2.u() - $$9; $$10 <= $$2.u() + $$9; ++$$10) {
                        for (int $$11 = $$2.w() - $$9; $$11 <= $$2.w() + $$9; ++$$11) {
                            int $$13;
                            int $$12 = $$10 - $$2.u();
                            if ($$12 * $$12 + ($$13 = $$11 - $$2.w()) * $$13 > $$9 * $$9) continue;
                            $$7.d($$10, $$3.a(dyy.a.b, $$10, $$11) - 1, $$11);
                            if (!ebh.b((dtc)$$3.a_((jd)$$7))) continue;
                            $$3.a((jd)$$7, dga.l.o(), 2);
                        }
                    }
                }
                for (int $$14 = 0; $$14 < $$8 && $$3.u((jd)$$6); ++$$14) {
                    $$3.a((jd)$$6, a, 2);
                    $$6.c(ji.b, 1);
                }
                if ($$6.v() - $$2.v() >= 3) {
                    $$3.a((jd)$$6, b, 2);
                    $$3.a((jd)$$6.c(ji.a, 1), c, 2);
                    $$3.a((jd)$$6.c(ji.a, 1), d, 2);
                }
            }
            ++$$1;
        }
        return $$1 > 0;
    }
}
