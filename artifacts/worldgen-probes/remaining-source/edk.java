/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  ayw
 *  com.mojang.serialization.Codec
 *  dcz
 *  dds
 *  dga
 *  dof
 *  dtc
 *  dty
 *  duf
 *  dyy$a
 *  ece
 *  ecg
 *  ees
 *  jd
 */
import com.mojang.serialization.Codec;

public class edk
extends ece<ees> {
    public edk(Codec<ees> $$0) {
        super($$0);
    }

    public boolean a(ecg<ees> $$0) {
        boolean $$1 = false;
        ayw $$2 = $$0.d();
        dds $$3 = $$0.b();
        jd $$4 = $$0.e();
        ees $$5 = (ees)$$0.f();
        int $$6 = $$2.a(8) - $$2.a(8);
        int $$7 = $$2.a(8) - $$2.a(8);
        int $$8 = $$3.a(dyy.a.d, $$4.u() + $$6, $$4.w() + $$7);
        jd $$9 = new jd($$4.u() + $$6, $$8, $$4.w() + $$7);
        if ($$3.a_($$9).a(dga.G)) {
            dtc $$11;
            boolean $$10 = $$2.j() < (double)$$5.l;
            dtc dtc2 = $$11 = $$10 ? dga.bx.o() : dga.bw.o();
            if ($$11.a((dcz)$$3, $$9)) {
                if ($$10) {
                    dtc $$12 = (dtc)$$11.a((duf)dof.d, (Comparable)dty.a);
                    jd $$13 = $$9.d();
                    if ($$3.a_($$13).a(dga.G)) {
                        $$3.a($$9, $$11, 2);
                        $$3.a($$13, $$12, 2);
                    }
                } else {
                    $$3.a($$9, $$11, 2);
                }
                $$1 = true;
            }
        }
        return $$1;
    }
}
