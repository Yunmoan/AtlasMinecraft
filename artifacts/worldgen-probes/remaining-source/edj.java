/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  ayw
 *  com.mojang.serialization.Codec
 *  dcz
 *  dds
 *  dga
 *  dmw
 *  dtc
 *  duf
 *  dyy$a
 *  ece
 *  ecg
 *  eed
 *  jd
 */
import com.mojang.serialization.Codec;

public class edj
extends ece<eed> {
    public edj(Codec<eed> $$0) {
        super($$0);
    }

    public boolean a(ecg<eed> $$0) {
        int $$1 = 0;
        ayw $$2 = $$0.d();
        dds $$3 = $$0.b();
        jd $$4 = $$0.e();
        int $$5 = ((eed)$$0.f()).a().a($$2);
        for (int $$6 = 0; $$6 < $$5; ++$$6) {
            int $$7 = $$2.a(8) - $$2.a(8);
            int $$8 = $$2.a(8) - $$2.a(8);
            int $$9 = $$3.a(dyy.a.d, $$4.u() + $$7, $$4.w() + $$8);
            jd $$10 = new jd($$4.u() + $$7, $$9, $$4.w() + $$8);
            dtc $$11 = (dtc)dga.mV.o().a((duf)dmw.c, (Comparable)Integer.valueOf($$2.a(4) + 1));
            if (!$$3.a_($$10).a(dga.G) || !$$11.a((dcz)$$3, $$10)) continue;
            $$3.a($$10, $$11, 2);
            ++$$1;
        }
        return $$1 > 0;
    }
}
