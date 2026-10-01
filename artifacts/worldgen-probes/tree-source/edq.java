/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.google.common.collect.Iterables
 *  com.google.common.collect.Lists
 *  com.google.common.collect.Sets
 *  com.mojang.serialization.Codec
 */
import com.google.common.collect.Iterables;
import com.google.common.collect.Lists;
import com.google.common.collect.Sets;
import com.mojang.serialization.Codec;
import java.util.HashSet;
import java.util.List;
import java.util.OptionalInt;
import java.util.Set;
import java.util.function.BiConsumer;

public class edq
extends ece<efe> {
    private static final int a = 19;

    public edq(Codec<efe> $$0) {
        super($$0);
    }

    private static boolean d(ddc $$02, jd $$1) {
        return $$02.a($$1, (dtc $$0) -> $$0.a(dga.ff));
    }

    public static boolean b(ddc $$02, jd $$1) {
        return $$02.a($$1, (dtc $$0) -> $$0.i() || $$0.a(awe.P));
    }

    private static void b(dde $$0, jd $$1, dtc $$2) {
        $$0.a($$1, $$2, 19);
    }

    public static boolean c(ddc $$02, jd $$1) {
        return $$02.a($$1, (dtc $$0) -> $$0.i() || $$0.a(awe.co));
    }

    private boolean a(dds $$0, ayw $$1, jd $$22, BiConsumer<jd, dtc> $$3, BiConsumer<jd, dtc> $$4, efu.b $$5, efe $$6) {
        int $$72 = $$6.d.a($$1);
        int $$8 = $$6.f.a($$1, $$72, $$6);
        int $$9 = $$72 - $$8;
        int $$10 = $$6.f.a($$1, $$9);
        jd $$11 = $$6.g.map($$2 -> $$2.a($$22, $$1)).orElse($$22);
        int $$12 = Math.min($$22.v(), $$11.v());
        int $$13 = Math.max($$22.v(), $$11.v()) + $$72 + 1;
        if ($$12 < $$0.I_() + 1 || $$13 > $$0.am()) {
            return false;
        }
        OptionalInt $$14 = $$6.h.c();
        int $$15 = this.a((ddc)$$0, $$72, $$11, $$6);
        if ($$15 < $$72 && ($$14.isEmpty() || $$15 < $$14.getAsInt())) {
            return false;
        }
        if ($$6.g.isPresent() && !$$6.g.get().a($$0, $$3, $$1, $$22, $$11, $$6)) {
            return false;
        }
        List<efu.a> $$16 = $$6.d.a((ddc)$$0, $$4, $$1, $$15, $$11, $$6);
        $$16.forEach($$7 -> $$0.f.a((ddc)$$0, $$5, $$1, $$6, $$15, (efu.a)$$7, $$8, $$10));
        return true;
    }

    private int a(ddc $$0, int $$1, jd $$2, efe $$3) {
        jd.a $$4 = new jd.a();
        for (int $$5 = 0; $$5 <= $$1 + 1; ++$$5) {
            int $$6 = $$3.h.a($$1, $$5);
            for (int $$7 = -$$6; $$7 <= $$6; ++$$7) {
                for (int $$8 = -$$6; $$8 <= $$6; ++$$8) {
                    $$4.a($$2, $$7, $$5, $$8);
                    if ($$3.d.b($$0, $$4) && ($$3.j || !edq.d($$0, $$4))) continue;
                    return $$5 - 2;
                }
            }
        }
        return $$1;
    }

    @Override
    protected void a(dde $$0, jd $$1, dtc $$2) {
        edq.b($$0, $$1, $$2);
    }

    @Override
    public final boolean a(ecg<efe> $$0) {
        final dds $$12 = $$0.b();
        ayw $$22 = $$0.d();
        jd $$32 = $$0.e();
        efe $$42 = $$0.f();
        HashSet $$5 = Sets.newHashSet();
        HashSet $$6 = Sets.newHashSet();
        final HashSet $$7 = Sets.newHashSet();
        HashSet $$8 = Sets.newHashSet();
        BiConsumer<jd, dtc> $$9 = ($$2, $$3) -> {
            $$5.add($$2.j());
            $$12.a((jd)$$2, (dtc)$$3, 19);
        };
        BiConsumer<jd, dtc> $$10 = ($$2, $$3) -> {
            $$6.add($$2.j());
            $$12.a((jd)$$2, (dtc)$$3, 19);
        };
        efu.b $$11 = new efu.b(){

            @Override
            public void a(jd $$0, dtc $$1) {
                $$7.add($$0.j());
                $$12.a($$0, $$1, 19);
            }

            @Override
            public boolean a(jd $$0) {
                return $$7.contains($$0);
            }
        };
        BiConsumer<jd, dtc> $$122 = ($$2, $$3) -> {
            $$8.add($$2.j());
            $$12.a((jd)$$2, (dtc)$$3, 19);
        };
        boolean $$13 = this.a($$12, $$22, $$32, $$9, $$10, $$11, $$42);
        if (!$$13 || $$6.isEmpty() && $$7.isEmpty()) {
            return false;
        }
        if (!$$42.i.isEmpty()) {
            eha.a $$14 = new eha.a($$12, $$122, $$22, $$6, $$7, $$5);
            $$42.i.forEach($$1 -> $$1.a($$14));
        }
        return ejj.a(Iterables.concat((Iterable)$$5, (Iterable)$$6, (Iterable)$$7, (Iterable)$$8)).map($$4 -> {
            exl $$5 = edq.a((dcx)$$12, $$4, (Set<jd>)$$6, (Set<jd>)$$8, $$5);
            ent.a($$12, 3, $$5, $$4.h(), $$4.i(), $$4.j());
            return true;
        }).orElse(false);
    }

    /*
     * Unable to fully structure code
     */
    private static exl a(dcx $$0, ejj $$1, Set<jd> $$2, Set<jd> $$3, Set<jd> $$4) {
        $$5 = new exf($$1.d(), $$1.e(), $$1.f());
        $$6 = 7;
        $$7 = Lists.newArrayList();
        for ($$8 = 0; $$8 < 7; ++$$8) {
            $$7.add(Sets.newHashSet());
        }
        for (jd $$9 : Lists.newArrayList((Iterable)Sets.union($$3, $$4))) {
            if (!$$1.b($$9)) continue;
            $$5.c($$9.u() - $$1.h(), $$9.v() - $$1.i(), $$9.w() - $$1.j());
        }
        $$10 = new jd.a();
        $$11 = 0;
        ((Set)$$7.get(0)).addAll($$2);
        block2: while (true) {
            if ($$11 < 7 && ((Set)$$7.get($$11)).isEmpty()) {
                ++$$11;
                continue;
            }
            if ($$11 >= 7) break;
            $$12 = ((Set)$$7.get($$11)).iterator();
            $$13 = (jd)$$12.next();
            $$12.remove();
            if (!$$1.b($$13)) continue;
            if ($$11 != 0) {
                $$14 = $$0.a_($$13);
                edq.b($$0, $$13, (dtc)$$14.a(dts.aC, $$11));
            }
            $$5.c($$13.u() - $$1.h(), $$13.v() - $$1.i(), $$13.w() - $$1.j());
            var12_14 = ji.values();
            var13_15 = var12_14.length;
            var14_16 = 0;
            while (true) {
                if (var14_16 < var13_15) ** break;
                continue block2;
                $$15 = var12_14[var14_16];
                $$10.a((kh)$$13, $$15);
                if ($$1.b($$10) && !$$5.b($$16 = $$10.u() - $$1.h(), $$17 = $$10.v() - $$1.i(), $$18 = $$10.w() - $$1.j()) && !($$20 = dki.n($$19 = $$0.a_($$10))).isEmpty() && ($$21 = Math.min($$20.getAsInt(), $$11 + 1)) < 7) {
                    ((Set)$$7.get($$21)).add($$10.j());
                    $$11 = Math.min($$11, $$21);
                }
                ++var14_16;
            }
            break;
        }
        return $$5;
    }
}
