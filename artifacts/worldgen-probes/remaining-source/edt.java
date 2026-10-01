/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.serialization.Codec
 */
import com.mojang.serialization.Codec;
import java.util.HashSet;
import java.util.Set;
import java.util.function.Predicate;

public class edt
extends ece<efh> {
    public edt(Codec<efh> $$0) {
        super($$0);
    }

    @Override
    public boolean a(ecg<efh> $$0) {
        dds $$12 = $$0.b();
        efh $$2 = $$0.f();
        ayw $$3 = $$0.d();
        jd $$4 = $$0.e();
        Predicate<dtc> $$5 = $$1 -> $$1.a($$0.b);
        int $$6 = $$2.j.a($$3) + 1;
        int $$7 = $$2.j.a($$3) + 1;
        Set<jd> $$8 = this.a($$12, $$2, $$3, $$4, $$5, $$6, $$7);
        this.a($$0, $$12, $$2, $$3, $$8, $$6, $$7);
        return !$$8.isEmpty();
    }

    protected Set<jd> a(dds $$02, efh $$1, ayw $$2, jd $$3, Predicate<dtc> $$4, int $$5, int $$6) {
        jd.a $$7 = $$3.k();
        jd.a $$8 = $$7.k();
        ji $$9 = $$1.e.a();
        ji $$10 = $$9.g();
        HashSet<jd> $$11 = new HashSet<jd>();
        for (int $$12 = -$$5; $$12 <= $$5; ++$$12) {
            boolean $$13 = $$12 == -$$5 || $$12 == $$5;
            for (int $$14 = -$$6; $$14 <= $$6; ++$$14) {
                int $$19;
                boolean $$18;
                boolean $$15 = $$14 == -$$6 || $$14 == $$6;
                boolean $$16 = $$13 || $$15;
                boolean $$17 = $$13 && $$15;
                boolean bl = $$18 = $$16 && !$$17;
                if ($$17 || $$18 && ($$1.k == 0.0f || $$2.i() > $$1.k)) continue;
                $$7.a($$3, $$12, 0, $$14);
                for ($$19 = 0; $$02.a((jd)$$7, dtb.a::i) && $$19 < $$1.h; ++$$19) {
                    $$7.c($$9);
                }
                for ($$19 = 0; $$02.a((jd)$$7, (dtc $$0) -> !$$0.i()) && $$19 < $$1.h; ++$$19) {
                    $$7.c($$10);
                }
                $$8.a((kh)$$7, $$1.e.a());
                dtc $$20 = $$02.a_($$8);
                if (!$$02.u($$7) || !$$20.d($$02, $$8, $$1.e.a().g())) continue;
                int $$21 = $$1.f.a($$2) + ($$1.g > 0.0f && $$2.i() < $$1.g ? 1 : 0);
                jd $$22 = $$8.j();
                boolean $$23 = this.a($$02, $$1, $$4, $$2, $$8, $$21);
                if (!$$23) continue;
                $$11.add($$22);
            }
        }
        return $$11;
    }

    protected void a(ecg<efh> $$0, dds $$1, efh $$2, ayw $$3, Set<jd> $$4, int $$5, int $$6) {
        for (jd $$7 : $$4) {
            if (!($$2.i > 0.0f) || !($$3.i() < $$2.i)) continue;
            this.a($$1, $$2, $$0.c(), $$3, $$7);
        }
    }

    protected boolean a(dds $$0, efh $$1, duz $$2, ayw $$3, jd $$4) {
        return $$1.d.a().a($$0, $$2, $$3, $$4.a($$1.e.a().g()));
    }

    protected boolean a(dds $$0, efh $$1, Predicate<dtc> $$2, ayw $$3, jd.a $$4, int $$5) {
        for (int $$6 = 0; $$6 < $$5; ++$$6) {
            dtc $$8;
            dtc $$7 = $$1.c.a($$3, $$4);
            if ($$7.a(($$8 = $$0.a_($$4)).b())) continue;
            if (!$$2.test($$8)) {
                return $$6 != 0;
            }
            $$0.a((jd)$$4, $$7, 2);
            $$4.c($$1.e.a());
        }
        return true;
    }
}
