/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  awe
 *  awk
 *  ayo
 *  ayw
 *  bpq
 *  com.mojang.serialization.Codec
 *  dcx
 *  dcz
 *  ddc
 *  dds
 *  dga
 *  dtc
 *  dyo
 *  ebz
 *  ece
 *  ecg
 *  eeg
 *  jd
 *  jd$a
 *  ji
 *  ji$c
 */
import com.mojang.serialization.Codec;
import java.util.Optional;
import java.util.OptionalInt;

public class eby
extends ece<eeg> {
    public eby(Codec<eeg> $$0) {
        super($$0);
    }

    public boolean a(ecg<eeg> $$0) {
        dds $$1 = $$0.b();
        jd $$2 = $$0.e();
        eeg $$3 = (eeg)$$0.f();
        ayw $$4 = $$0.d();
        if (!ebz.a((dcx)$$1, (jd)$$2)) {
            return false;
        }
        int $$5 = $$3.c.a($$4);
        float $$6 = $$3.i.a($$4);
        float $$7 = $$3.h.a($$4);
        int $$8 = $$3.d.a($$4);
        int $$9 = $$3.d.a($$4);
        for (int $$10 = -$$8; $$10 <= $$8; ++$$10) {
            for (int $$11 = -$$9; $$11 <= $$9; ++$$11) {
                double $$12 = this.a($$8, $$9, $$10, $$11, $$3);
                jd $$13 = $$2.b($$10, 0, $$11);
                this.a($$1, $$4, $$13, $$10, $$11, $$6, $$12, $$5, $$7, $$3);
            }
        }
        return true;
    }

    private void a(dds $$0, ayw $$1, jd $$2, int $$3, int $$4, float $$5, double $$6, int $$7, float $$8, eeg $$9) {
        boolean $$39;
        int $$38;
        int $$37;
        int $$28;
        boolean $$24;
        int $$23;
        boolean $$18;
        dyo $$16;
        boolean $$13;
        Optional $$10 = dyo.a((ddc)$$0, (jd)$$2, (int)$$9.b, ebz::c, ebz::d);
        if ($$10.isEmpty()) {
            return;
        }
        OptionalInt $$11 = ((dyo)$$10.get()).b();
        OptionalInt $$12 = ((dyo)$$10.get()).c();
        if ($$11.isEmpty() && $$12.isEmpty()) {
            return;
        }
        boolean bl = $$13 = $$1.i() < $$5;
        if ($$13 && $$12.isPresent() && this.b($$0, $$2.h($$12.getAsInt()))) {
            int $$14 = $$12.getAsInt();
            dyo $$15 = ((dyo)$$10.get()).a(OptionalInt.of($$14 - 1));
            $$0.a($$2.h($$14), dga.G.o(), 2);
        } else {
            $$16 = (dyo)$$10.get();
        }
        OptionalInt $$17 = $$16.c();
        boolean bl2 = $$18 = $$1.j() < $$6;
        if ($$11.isPresent() && $$18 && !this.a((dcz)$$0, $$2.h($$11.getAsInt()))) {
            int $$21;
            int $$19 = $$9.g.a($$1);
            this.a($$0, $$2.h($$11.getAsInt()), $$19, ji.b);
            if ($$17.isPresent()) {
                int $$20 = Math.min($$7, $$11.getAsInt() - $$17.getAsInt());
            } else {
                $$21 = $$7;
            }
            int $$22 = this.a($$1, $$3, $$4, $$8, $$21, $$9);
        } else {
            $$23 = 0;
        }
        boolean bl3 = $$24 = $$1.j() < $$6;
        if ($$17.isPresent() && $$24 && !this.a((dcz)$$0, $$2.h($$17.getAsInt()))) {
            int $$25 = $$9.g.a($$1);
            this.a($$0, $$2.h($$17.getAsInt()), $$25, ji.a);
            if ($$11.isPresent()) {
                int $$26 = Math.max(0, $$23 + ayo.b((ayw)$$1, (int)(-$$9.e), (int)$$9.e));
            } else {
                int $$27 = this.a($$1, $$3, $$4, $$8, $$7, $$9);
            }
        } else {
            $$28 = 0;
        }
        if ($$11.isPresent() && $$17.isPresent() && $$11.getAsInt() - $$23 <= $$17.getAsInt() + $$28) {
            int $$29 = $$17.getAsInt();
            int $$30 = $$11.getAsInt();
            int $$31 = Math.max($$30 - $$23, $$29 + 1);
            int $$32 = Math.min($$29 + $$28, $$30 - 1);
            int $$33 = ayo.b((ayw)$$1, (int)$$31, (int)($$32 + 1));
            int $$34 = $$33 - 1;
            int $$35 = $$30 - $$33;
            int $$36 = $$34 - $$29;
        } else {
            $$37 = $$23;
            $$38 = $$28;
        }
        boolean bl4 = $$39 = $$1.h() && $$37 > 0 && $$38 > 0 && $$16.d().isPresent() && $$37 + $$38 == $$16.d().getAsInt();
        if ($$11.isPresent()) {
            ebz.a((dcx)$$0, (jd)$$2.h($$11.getAsInt() - 1), (ji)ji.a, (int)$$37, (boolean)$$39);
        }
        if ($$17.isPresent()) {
            ebz.a((dcx)$$0, (jd)$$2.h($$17.getAsInt() + 1), (ji)ji.b, (int)$$38, (boolean)$$39);
        }
    }

    private boolean a(dcz $$0, jd $$1) {
        return $$0.a_($$1).a(dga.H);
    }

    private int a(ayw $$0, int $$1, int $$2, float $$3, int $$4, eeg $$5) {
        if ($$0.i() > $$3) {
            return 0;
        }
        int $$6 = Math.abs($$1) + Math.abs($$2);
        float $$7 = (float)ayo.a((double)$$6, (double)0.0, (double)$$5.l, (double)((double)$$4 / 2.0), (double)0.0);
        return (int)eby.a($$0, 0.0f, $$4, $$7, $$5.f);
    }

    private boolean b(dds $$0, jd $$1) {
        dtc $$2 = $$0.a_($$1);
        if ($$2.a(dga.G) || $$2.a(dga.su) || $$2.a(dga.st)) {
            return false;
        }
        if ($$0.a_($$1.d()).u().a(awk.a)) {
            return false;
        }
        for (ji $$3 : ji.c.a) {
            if (this.a((dcx)$$0, $$1.a($$3))) continue;
            return false;
        }
        return this.a((dcx)$$0, $$1.e());
    }

    private boolean a(dcx $$0, jd $$1) {
        dtc $$2 = $$0.a_($$1);
        return $$2.a(awe.bf) || $$2.u().a(awk.a);
    }

    private void a(dds $$0, jd $$1, int $$2, ji $$3) {
        jd.a $$4 = $$1.k();
        for (int $$5 = 0; $$5 < $$2; ++$$5) {
            if (!ebz.c((dcx)$$0, (jd)$$4)) {
                return;
            }
            $$4.c($$3);
        }
    }

    private double a(int $$0, int $$1, int $$2, int $$3, eeg $$4) {
        int $$5 = $$0 - Math.abs($$2);
        int $$6 = $$1 - Math.abs($$3);
        int $$7 = Math.min($$5, $$6);
        return ayo.b((float)$$7, (float)0.0f, (float)$$4.k, (float)$$4.j, (float)1.0f);
    }

    private static float a(ayw $$0, float $$1, float $$2, float $$3, float $$4) {
        return bpq.a((ayw)$$0, (float)$$3, (float)$$4, (float)$$1, (float)$$2);
    }
}
