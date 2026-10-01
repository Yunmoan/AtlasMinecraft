/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.google.common.collect.Lists
 *  com.mojang.datafixers.util.Pair
 *  com.mojang.serialization.Codec
 */
import com.google.common.collect.Lists;
import com.mojang.datafixers.util.Pair;
import com.mojang.serialization.Codec;
import java.util.ArrayList;
import java.util.LinkedList;
import java.util.List;
import java.util.function.Predicate;

public class eck
extends ece<eej> {
    private static final ji[] a = ji.values();

    public eck(Codec<eej> $$0) {
        super($$0);
    }

    @Override
    public boolean a(ecg<eej> $$0) {
        eej $$1 = $$0.f();
        ayw $$2 = $$0.d();
        jd $$3 = $$0.e();
        dds $$4 = $$0.b();
        int $$5 = $$1.l;
        int $$6 = $$1.n;
        LinkedList $$7 = Lists.newLinkedList();
        int $$8 = $$1.j.a($$2);
        dzx $$9 = new dzx(new dyz($$4.C()));
        eoh $$10 = eoh.a($$9, -4, 1.0);
        LinkedList $$11 = Lists.newLinkedList();
        double $$12 = (double)$$8 / (double)$$1.i.b();
        dyx $$13 = $$1.d;
        dyv $$14 = $$1.c;
        dyw $$15 = $$1.e;
        double $$16 = 1.0 / Math.sqrt($$13.b);
        double $$17 = 1.0 / Math.sqrt($$13.c + $$12);
        double $$18 = 1.0 / Math.sqrt($$13.d + $$12);
        double $$19 = 1.0 / Math.sqrt($$13.e + $$12);
        double $$20 = 1.0 / Math.sqrt($$15.c + $$2.j() / 2.0 + ($$8 > 3 ? $$12 : 0.0));
        boolean $$21 = (double)$$2.i() < $$15.b;
        int $$22 = 0;
        for (int $$23 = 0; $$23 < $$8; ++$$23) {
            int $$26;
            int $$25;
            int $$24 = $$1.i.a($$2);
            jd $$27 = $$3.b($$24, $$25 = $$1.i.a($$2), $$26 = $$1.i.a($$2));
            dtc $$28 = $$4.a_($$27);
            if (($$28.i() || $$28.a(awe.bR)) && ++$$22 > $$1.p) {
                return false;
            }
            $$7.add(Pair.of((Object)$$27, (Object)$$1.k.a($$2)));
        }
        if ($$21) {
            int $$29 = $$2.a(4);
            int $$30 = $$8 * 2 + 1;
            if ($$29 == 0) {
                $$11.add($$3.b($$30, 7, 0));
                $$11.add($$3.b($$30, 5, 0));
                $$11.add($$3.b($$30, 1, 0));
            } else if ($$29 == 1) {
                $$11.add($$3.b(0, 7, $$30));
                $$11.add($$3.b(0, 5, $$30));
                $$11.add($$3.b(0, 1, $$30));
            } else if ($$29 == 2) {
                $$11.add($$3.b($$30, 7, $$30));
                $$11.add($$3.b($$30, 5, $$30));
                $$11.add($$3.b($$30, 1, $$30));
            } else {
                $$11.add($$3.b(0, 7, 0));
                $$11.add($$3.b(0, 5, 0));
                $$11.add($$3.b(0, 1, 0));
            }
        }
        ArrayList $$31 = Lists.newArrayList();
        Predicate<dtc> $$32 = eck.a($$1.c.g);
        for (jd $$33 : jd.c($$3.b($$5, $$5, $$5), $$3.b($$6, $$6, $$6))) {
            double $$34 = $$10.a($$33.u(), $$33.v(), $$33.w()) * $$1.o;
            double $$35 = 0.0;
            double $$36 = 0.0;
            for (Pair $$37 : $$7) {
                $$35 += ayo.f($$33.j((kh)$$37.getFirst()) + (double)((Integer)$$37.getSecond()).intValue()) + $$34;
            }
            for (jd $$38 : $$11) {
                $$36 += ayo.f($$33.j($$38) + (double)$$15.d) + $$34;
            }
            if ($$35 < $$19) continue;
            if ($$21 && $$36 >= $$20 && $$35 < $$16) {
                this.a($$4, $$33, dga.a.o(), $$32);
                for (ji $$39 : a) {
                    jd $$40 = $$33.a($$39);
                    epe $$41 = $$4.b_($$40);
                    if ($$41.c()) continue;
                    $$4.a($$40, $$41.a(), 0);
                }
                continue;
            }
            if ($$35 >= $$16) {
                this.a($$4, $$33, $$14.a.a($$2, $$33), $$32);
                continue;
            }
            if ($$35 >= $$17) {
                boolean $$42;
                boolean bl = $$42 = (double)$$2.i() < $$1.g;
                if ($$42) {
                    this.a($$4, $$33, $$14.c.a($$2, $$33), $$32);
                } else {
                    this.a($$4, $$33, $$14.b.a($$2, $$33), $$32);
                }
                if ($$1.h && !$$42 || !((double)$$2.i() < $$1.f)) continue;
                $$31.add($$33.j());
                continue;
            }
            if ($$35 >= $$18) {
                this.a($$4, $$33, $$14.d.a($$2, $$33), $$32);
                continue;
            }
            if (!($$35 >= $$19)) continue;
            this.a($$4, $$33, $$14.e.a($$2, $$33), $$32);
        }
        List<dtc> $$43 = $$14.f;
        block5: for (jd $$44 : $$31) {
            dtc $$45 = ad.a($$43, $$2);
            for (ji $$46 : a) {
                if ($$45.b(dts.P)) {
                    $$45 = (dtc)$$45.a(dts.P, $$46);
                }
                jd $$47 = $$44.a($$46);
                dtc $$48 = $$4.a_($$47);
                if ($$45.b(dts.C)) {
                    $$45 = (dtc)$$45.a(dts.C, $$48.u().b());
                }
                if (!dgg.g($$48)) continue;
                this.a($$4, $$47, $$45, $$32);
                continue block5;
            }
        }
        return true;
    }
}
